package tests

import (
	"context"
	"fmt"
	"net"

	card_handler "github.com/MamangRust/monolith-payment-gateway-card/handler"
	card_repository "github.com/MamangRust/monolith-payment-gateway-card/repository"
	card_service "github.com/MamangRust/monolith-payment-gateway-card/service"
	merchant_handler "github.com/MamangRust/monolith-payment-gateway-merchant/handler"
	merchant_repository "github.com/MamangRust/monolith-payment-gateway-merchant/repository"
	merchant_service "github.com/MamangRust/monolith-payment-gateway-merchant/service"
	pbcard "github.com/MamangRust/monolith-payment-gateway-pb/card"
	pbmerchant "github.com/MamangRust/monolith-payment-gateway-pb/merchant"
	pbrole "github.com/MamangRust/monolith-payment-gateway-pb/role"
	pbsaldo "github.com/MamangRust/monolith-payment-gateway-pb/saldo"
	pbuser "github.com/MamangRust/monolith-payment-gateway-pb/user"
	pbuserrole "github.com/MamangRust/monolith-payment-gateway-pb/user_role"
	adapter "github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	"github.com/MamangRust/monolith-payment-gateway-pkg/hash"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	saldo_handler "github.com/MamangRust/monolith-payment-gateway-saldo/handler"
	saldo_repository "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	saldo_service "github.com/MamangRust/monolith-payment-gateway-saldo/service"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	user_handler "github.com/MamangRust/monolith-payment-gateway-user/handler"
	user_repository "github.com/MamangRust/monolith-payment-gateway-user/repository"
	user_service "github.com/MamangRust/monolith-payment-gateway-user/service"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DependencyHarness hosts the real card, saldo and merchant gRPC services
// in-process so services that talk to them through gRPC adapters can be
// exercised end-to-end.
type DependencyHarness struct {
	CardAdapter     adapter.CardAdapter
	SaldoAdapter    adapter.SaldoAdapter
	MerchantAdapter adapter.MerchantAdapter
	UserAdapter     adapter.UserAdapter
	Conn            *grpc.ClientConn
	Server          *grpc.Server
}

// NewDependencyHarness starts card + saldo services backed by db and returns
// adapters wired to them. Call Close when the test suite tears down.
func NewDependencyHarness(db *mongo.Database, cacheStore *cache.CacheStore, log logger.LoggerInterface) (*DependencyHarness, error) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to listen for dependency gRPC server: %w", err)
	}

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial dependency gRPC server: %w", err)
	}

	cardAdapter := adapter.NewCardAdapter(pbcard.NewCardQueryServiceClient(conn), pbcard.NewCardCommandServiceClient(conn))
	saldoAdapter := adapter.NewSaldoAdapter(pbsaldo.NewSaldoQueryServiceClient(conn), pbsaldo.NewSaldoCommandServiceClient(conn))
	merchantAdapter := adapter.NewMerchantAdapter(pbmerchant.NewMerchantQueryServiceClient(conn))
	userAdapter := adapter.NewUserAdapter(pbuser.NewUserQueryServiceClient(conn))

	server := grpc.NewServer()

	// The role service backs the user service's role adapters: CreateUser
	// resolves and assigns the default ROLE_ADMIN through it.
	registerRoleService(server, db, cacheStore, log)
	if _, err := seedRole(context.Background(), db, DefaultRoleName); err != nil {
		return nil, fmt.Errorf("failed to seed %s role: %w", DefaultRoleName, err)
	}

	userSvc := user_service.NewService(&user_service.Deps{
		Cache: cacheStore,
		Repositories: user_repository.NewRepositories(db,
			pbrole.NewRoleQueryServiceClient(conn),
			pbuserrole.NewUserRoleServiceClient(conn)),
		Hash:   hash.NewHashingPassword(),
		Logger: log,
	})
	userH := user_handler.NewHandler(userSvc)
	pbuser.RegisterUserQueryServiceServer(server, userH)
	pbuser.RegisterUserCommandServiceServer(server, userH)

	cardSvc := card_service.NewService(&card_service.Deps{
		Cache:        cacheStore,
		Repositories: card_repository.NewRepositories(db, pbuser.NewUserQueryServiceClient(conn)),
		UserAdapter:  userAdapter,
		Logger:       log,
		Kafka:        nil,
	})
	cardH := card_handler.NewHandler(cardSvc)
	pbcard.RegisterCardQueryServiceServer(server, cardH)
	pbcard.RegisterCardCommandServiceServer(server, cardH)

	saldoSvc := saldo_service.NewService(&saldo_service.Deps{
		Cache:        cacheStore,
		Repositories: saldo_repository.NewRepositories(db, nil, nil),
		CardAdapter:  cardAdapter,
		Logger:       log,
		DB:           db,
	})
	saldoH := saldo_handler.NewHandler(saldoSvc)
	pbsaldo.RegisterSaldoQueryServiceServer(server, saldoH)
	pbsaldo.RegisterSaldoCommandServiceServer(server, saldoH)

	merchantSvc := merchant_service.NewService(&merchant_service.Deps{
		DB:           db,
		Repositories: merchant_repository.NewRepositories(db, nil),
		Logger:       log,
		Cache:        cacheStore,
	})
	merchantH := merchant_handler.NewHandler(merchantSvc)
	pbmerchant.RegisterMerchantQueryServiceServer(server, merchantH)
	pbmerchant.RegisterMerchantCommandServiceServer(server, merchantH)

	go func() { _ = server.Serve(lis) }()

	return &DependencyHarness{
		CardAdapter:     cardAdapter,
		SaldoAdapter:    saldoAdapter,
		MerchantAdapter: merchantAdapter,
		UserAdapter:     userAdapter,
		Conn:            conn,
		Server:          server,
	}, nil
}

// Close stops the harness gRPC server and closes the client connection.
func (h *DependencyHarness) Close() {
	if h.Conn != nil {
		_ = h.Conn.Close()
	}
	if h.Server != nil {
		h.Server.Stop()
	}
}
