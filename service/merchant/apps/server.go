package apps

import (
	"fmt"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-merchant/handler"
	"github.com/MamangRust/monolith-payment-gateway-merchant/repository"
	"github.com/MamangRust/monolith-payment-gateway-merchant/service"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/merchant"
	pbstats "github.com/MamangRust/monolith-payment-gateway-pb/merchant/stats"
	pbdocument "github.com/MamangRust/monolith-payment-gateway-pb/merchant_document"
	"github.com/MamangRust/monolith-payment-gateway-pb/user"
	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	"github.com/MamangRust/monolith-payment-gateway-pkg/kafka"
	"github.com/MamangRust/monolith-payment-gateway-pkg/resilience"
	"github.com/MamangRust/monolith-payment-gateway-pkg/server"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewServer(cfg *server.Config) (*server.GRPCServer, error) {
	srv, err := server.New(cfg)
	if err != nil {
		return nil, err
	}

	myKafka := kafka.NewKafka(srv.Logger, []string{viper.GetString("KAFKA_BROKERS")})
	srv.AddCleanupHook(myKafka.Close)

	connUser, err := grpc.NewClient(
		viper.GetString("GRPC_USER_ADDR"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to user service: %w", err)
	}
	srv.AddCleanupHook(func() error { return connUser.Close() })

	userClient := user.NewUserQueryServiceClient(connUser)

	guardUser := resilience.NewDependencyGuard("user", 5, 30, 100, 3*time.Second, srv.Logger)

	repos := repository.NewRepositories(srv.MongoDB, userClient,
		repository.GuardOptions{
			User: []adapter.GuardOption{adapter.WithDependencyGuard(guardUser)},
		})
	svc := service.NewService(&service.Deps{
		Cache:        srv.CacheStore,
		Logger:       srv.Logger,
		Repositories: repos,
		UserAdapter:  repos.UserAdapter(),
		DB:           srv.MongoDB,
	})
	h := handler.NewHandler(svc)

	srv.RegisterServices = func(gs *grpc.Server) {
		pb.RegisterMerchantQueryServiceServer(gs, h)
		pb.RegisterMerchantCommandServiceServer(gs, h)
		pb.RegisterMerchantTransactionServiceServer(gs, h)
		pbdocument.RegisterMerchantDocumentQueryServiceServer(gs, h)
		pbdocument.RegisterMerchantDocumentCommandServiceServer(gs, h)
		pbstats.RegisterMerchantStatsAmountServiceServer(gs, h)
		pbstats.RegisterMerchantStatsMethodServiceServer(gs, h)
		pbstats.RegisterMerchantStatsTotalAmountServiceServer(gs, h)
	}

	return srv, nil
}
