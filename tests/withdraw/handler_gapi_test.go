package withdraw_test

import (
	"context"
	card_handler "github.com/MamangRust/monolith-payment-gateway-card/handler"
	card_repo "github.com/MamangRust/monolith-payment-gateway-card/repository"
	card_service "github.com/MamangRust/monolith-payment-gateway-card/service"
	pbcard "github.com/MamangRust/monolith-payment-gateway-pb/card"
	pbsaldo "github.com/MamangRust/monolith-payment-gateway-pb/saldo"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/withdraw"
	adapter "github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	saldo_handler "github.com/MamangRust/monolith-payment-gateway-saldo/handler"
	saldo_repo "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	saldo_service "github.com/MamangRust/monolith-payment-gateway-saldo/service"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	"github.com/MamangRust/monolith-payment-gateway-test"
	user_repo "github.com/MamangRust/monolith-payment-gateway-user/repository"
	"github.com/MamangRust/monolith-payment-gateway-withdraw/handler"
	"github.com/MamangRust/monolith-payment-gateway-withdraw/repository"
	"github.com/MamangRust/monolith-payment-gateway-withdraw/service"
	"net"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type WithdrawGapiTestSuite struct {
	suite.Suite
	ts            *tests.TestSuite
	redisClient   *redis.Client
	grpcServer    *grpc.Server
	commandClient pb.WithdrawCommandServiceClient
	queryClient   pb.WithdrawQueryServiceClient
	conn          *grpc.ClientConn
	repos         repository.Repositories
	userRepo      user_repo.UserCommandRepository
	cardRepo      card_repo.CardCommandRepository
	saldoRepo     saldo_repo.Repositories

	cardNumber string
	withdrawID int32
}

func (s *WithdrawGapiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	mongoDB, err := s.ts.MongoDB()
	if err != nil {
		s.Require().NoError(err)
	}

	// Repositories for seeding and service dependencies
	userRepos := user_repo.NewUserCommandRepository(mongoDB.Collection("users"))
	cardRepos := card_repo.NewRepositories(mongoDB, nil)
	saldoRepos := saldo_repo.NewRepositories(mongoDB, nil, nil)

	s.userRepo = userRepos
	s.cardRepo = cardRepos.CardCommand
	s.saldoRepo = saldoRepos

	s.repos = repository.NewRepositoriesWithAdapters(mongoDB, &realCardRepo{query: cardRepos.CardQuery}, &realSaldoRepo{repo: saldoRepos})

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	lis, err := net.Listen("tcp", ":0")
	s.Require().NoError(err)

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	cardAdapter := adapter.NewCardAdapter(pbcard.NewCardQueryServiceClient(conn), pbcard.NewCardCommandServiceClient(conn))
	saldoAdapter := adapter.NewSaldoAdapter(pbsaldo.NewSaldoQueryServiceClient(conn), pbsaldo.NewSaldoCommandServiceClient(conn))

	saldoSvc := saldo_service.NewService(&saldo_service.Deps{
		Cache:        cacheStore,
		Repositories: saldo_repo.NewRepositories(mongoDB, nil, nil),
		CardAdapter:  cardAdapter,
		Logger:       log,
		DB:           mongoDB,
	})
	saldoH := saldo_handler.NewHandler(saldoSvc)

	cardSvc := card_service.NewService(&card_service.Deps{
		Cache:        cacheStore,
		Repositories: card_repo.NewRepositories(mongoDB, nil),
		Logger:       log,
		Kafka:        nil,
	})
	cardH := card_handler.NewHandler(cardSvc)

	withdrawService := service.NewService(&service.Deps{
		Kafka:        nil,
		Repositories: s.repos,
		CardAdapter:  cardAdapter,
		SaldoAdapter: saldoAdapter,
		Logger:       log,
		Cache:        cacheStore,
		DB:           mongoDB,
	})

	// Seed User, Card, Saldo
	user, _ := s.userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Withdraw", LastName: "Gapi", Email: "withdraw.gapi@test.com", Password: "password123",
	})
	card, _ := s.cardRepo.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID: int(user.UserID), CardType: "debit", ExpireDate: time.Now().AddDate(1, 0, 0), CVV: "999", CardProvider: "visa",
	})
	s.cardNumber = card.CardNumber
	s.saldoRepo.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{
		CardNumber: s.cardNumber, TotalBalance: 1000000,
	})

	// Start gRPC Server
	withdrawHandler := handler.NewHandler(withdrawService)
	server := grpc.NewServer()
	pb.RegisterWithdrawCommandServiceServer(server, withdrawHandler)
	pb.RegisterWithdrawQueryServiceServer(server, withdrawHandler)
	pbcard.RegisterCardQueryServiceServer(server, cardH)
	pbcard.RegisterCardCommandServiceServer(server, cardH)
	pbsaldo.RegisterSaldoQueryServiceServer(server, saldoH)
	pbsaldo.RegisterSaldoCommandServiceServer(server, saldoH)
	s.grpcServer = server

	go func() { _ = server.Serve(lis) }()

	s.commandClient = pb.NewWithdrawCommandServiceClient(conn)
	s.queryClient = pb.NewWithdrawQueryServiceClient(conn)
}

func (s *WithdrawGapiTestSuite) TearDownSuite() {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
	if s.redisClient != nil {
		s.redisClient.Close()
	}
	if s.ts != nil {
		s.ts.Teardown()
	}
}

func (s *WithdrawGapiTestSuite) Test1_Create() {
	ctx := context.Background()

	createReq := &pb.CreateWithdrawRequest{
		CardNumber:     s.cardNumber,
		WithdrawAmount: 100000,
		WithdrawTime:   timestamppb.New(time.Now()),
	}
	res, err := s.commandClient.CreateWithdraw(ctx, createReq)
	s.NoError(err)
	s.Equal(int32(100000), res.Data.WithdrawAmount)

	s.withdrawID = res.Data.WithdrawId

	// Verify balance
	saldo, _ := s.saldoRepo.FindByCardNumber(ctx, s.cardNumber)
	s.Equal(int64(900000), saldo.TotalBalance)
}

func (s *WithdrawGapiTestSuite) Test2_FindById() {
	s.Require().NotZero(s.withdrawID)

	ctx := context.Background()
	found, err := s.queryClient.FindByIdWithdraw(ctx, &pb.FindByIdWithdrawRequest{WithdrawId: s.withdrawID})
	s.NoError(err)
	s.Equal(s.withdrawID, found.Data.WithdrawId)
}

func (s *WithdrawGapiTestSuite) Test3_Update() {
	s.Require().NotZero(s.withdrawID)

	ctx := context.Background()
	updateReq := &pb.UpdateWithdrawRequest{
		WithdrawId:     s.withdrawID,
		CardNumber:     s.cardNumber,
		WithdrawAmount: 150000,
		WithdrawTime:   timestamppb.New(time.Now()),
	}
	updated, err := s.commandClient.UpdateWithdraw(ctx, updateReq)
	s.NoError(err)
	s.Equal(int32(150000), updated.Data.WithdrawAmount)

	// Verify adjusted balance (900k - 50k = 850k)
	saldo, _ := s.saldoRepo.FindByCardNumber(ctx, s.cardNumber)
	s.Equal(int64(850000), saldo.TotalBalance)
}

func (s *WithdrawGapiTestSuite) Test4_Trashed() {
	s.Require().NotZero(s.withdrawID)

	ctx := context.Background()
	_, err := s.commandClient.TrashedWithdraw(ctx, &pb.FindByIdWithdrawRequest{WithdrawId: s.withdrawID})
	s.NoError(err)
}

func (s *WithdrawGapiTestSuite) Test5_Restore() {
	s.Require().NotZero(s.withdrawID)

	ctx := context.Background()
	_, err := s.commandClient.RestoreWithdraw(ctx, &pb.FindByIdWithdrawRequest{WithdrawId: s.withdrawID})
	s.NoError(err)
}

func (s *WithdrawGapiTestSuite) Test6_PermanentDelete() {
	s.Require().NotZero(s.withdrawID)

	ctx := context.Background()
	_, err := s.commandClient.DeleteWithdrawPermanent(ctx, &pb.FindByIdWithdrawRequest{WithdrawId: s.withdrawID})
	s.NoError(err)
}

func TestWithdrawGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(WithdrawGapiTestSuite))
}
