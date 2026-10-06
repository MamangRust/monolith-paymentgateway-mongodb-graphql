package transfer_test

import (
	"context"
	"net"
	"testing"
	"time"

	card_repo "github.com/MamangRust/monolith-payment-gateway-card/repository"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/transfer"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	saldo_repo "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/MamangRust/monolith-payment-gateway-transfer/handler"
	"github.com/MamangRust/monolith-payment-gateway-transfer/repository"
	"github.com/MamangRust/monolith-payment-gateway-transfer/service"
	user_repo "github.com/MamangRust/monolith-payment-gateway-user/repository"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TransferGapiTestSuite struct {
	suite.Suite
	harness       *tests.DependencyHarness
	ts            *tests.TestSuite
	redisClient   *redis.Client
	grpcServer    *grpc.Server
	commandClient pb.TransferCommandServiceClient
	queryClient   pb.TransferQueryServiceClient
	conn          *grpc.ClientConn
	repos         repository.Repositories
	userRepo      user_repo.UserCommandRepository
	cardRepo      card_repo.Repositories
	saldoRepo     saldo_repo.Repositories

	senderCardNumber   string
	receiverCardNumber string
	transferID         int
}

func (s *TransferGapiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	mongoDB, err := s.ts.MongoDB()
	if err != nil {
		s.Require().NoError(err)
	}

	// Repositories for seeding
	s.userRepo = user_repo.NewUserCommandRepository(mongoDB.Collection("users"))
	s.cardRepo = *card_repo.NewRepositories(mongoDB, nil)
	s.saldoRepo = saldo_repo.NewRepositories(mongoDB, nil, nil)

	// Transfer repos
	s.repos = repository.NewRepositoriesWithAdapters(mongoDB, &realSaldoRepo{repo: s.saldoRepo}, &realCardRepo{query: s.cardRepo.CardQuery})

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	_, _ = observability.NewObservability("test", log)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	harness, err := tests.NewDependencyHarness(mongoDB, cacheStore, log)
	s.Require().NoError(err)
	s.harness = harness

	transferService := service.NewService(&service.Deps{
		Kafka:        nil,
		DB:           mongoDB,
		Repositories: s.repos,
		CardAdapter:  harness.CardAdapter,
		SaldoAdapter: harness.SaldoAdapter,
		Logger:       log,
		Cache:        cacheStore,
	})

	// Seed Sender
	sender, err := s.userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Sender",
		LastName:  "Gapi",
		Email:     "sender.gapi@test.com",
		Password:  "password123",
	})
	s.Require().NoError(err)

	sCard, err := s.cardRepo.CardCommand.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID:       int(sender.UserID),
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(1, 0, 0),
		CVV:          "111",
		CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.senderCardNumber = sCard.CardNumber

	_, err = s.saldoRepo.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{
		CardNumber:   s.senderCardNumber,
		TotalBalance: 1000000,
	})
	s.Require().NoError(err)

	// Seed Receiver
	receiver, err := s.userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Receiver",
		LastName:  "Gapi",
		Email:     "receiver.gapi@test.com",
		Password:  "password123",
	})
	s.Require().NoError(err)

	rCard, err := s.cardRepo.CardCommand.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID:       int(receiver.UserID),
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(1, 0, 0),
		CVV:          "222",
		CardProvider: "mastercard",
	})
	s.Require().NoError(err)
	s.receiverCardNumber = rCard.CardNumber

	_, err = s.saldoRepo.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{
		CardNumber:   s.receiverCardNumber,
		TotalBalance: 0,
	})
	s.Require().NoError(err)

	// Start gRPC Server
	transferHandlerGapi := handler.NewHandler(transferService)

	server := grpc.NewServer()
	pb.RegisterTransferCommandServiceServer(server, transferHandlerGapi)
	pb.RegisterTransferQueryServiceServer(server, transferHandlerGapi)
	s.grpcServer = server

	lis, err := net.Listen("tcp", ":0")
	s.Require().NoError(err)

	go func() {
		_ = server.Serve(lis)
	}()

	// Create gRPC Client
	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn
	s.commandClient = pb.NewTransferCommandServiceClient(conn)
	s.queryClient = pb.NewTransferQueryServiceClient(conn)
}

func (s *TransferGapiTestSuite) TearDownSuite() {
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

func (s *TransferGapiTestSuite) Test1_CreateTransfer() {
	ctx := context.Background()
	req := &pb.CreateTransferRequest{
		TransferFrom:   s.senderCardNumber,
		TransferTo:     s.receiverCardNumber,
		TransferAmount: 100000,
	}

	res, err := s.commandClient.CreateTransfer(ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(res.Data)
	s.transferID = int(res.Data.Id)

	// Verify balances
	senderSaldo, _ := s.saldoRepo.FindByCardNumber(ctx, s.senderCardNumber)
	s.Equal(int64(900000), senderSaldo.TotalBalance)

	receiverSaldo, _ := s.saldoRepo.FindByCardNumber(ctx, s.receiverCardNumber)
	s.Equal(int64(100000), receiverSaldo.TotalBalance)
}

func (s *TransferGapiTestSuite) Test2_FindTransferById() {
	s.Require().NotZero(s.transferID)
	ctx := context.Background()

	res, err := s.queryClient.FindByIdTransfer(ctx, &pb.FindByIdTransferRequest{
		TransferId: int32(s.transferID),
	})
	s.Require().NoError(err)
	s.Require().NotNil(res.Data)
	s.Equal(int32(s.transferID), res.Data.Id)
}

func (s *TransferGapiTestSuite) Test3_FindAllTransfers() {
	ctx := context.Background()
	res, err := s.queryClient.FindAllTransfer(ctx, &pb.FindAllTransferRequest{
		Page:     1,
		PageSize: 10,
	})
	s.Require().NoError(err)
	s.Require().NotNil(res.Data)
}

func (s *TransferGapiTestSuite) Test4_UpdateTransfer() {
	s.Require().NotZero(s.transferID)
	ctx := context.Background()

	req := &pb.UpdateTransferRequest{
		TransferId:     int32(s.transferID),
		TransferFrom:   s.senderCardNumber,
		TransferTo:     s.receiverCardNumber,
		TransferAmount: 150000, // Increase by 50000
	}

	res, err := s.commandClient.UpdateTransfer(ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(res.Data)

	// Verify adjusted balances (Sender 900k - 50k = 850k, Receiver 100k + 50k = 150k)
	senderSaldo, _ := s.saldoRepo.FindByCardNumber(ctx, s.senderCardNumber)
	s.Equal(int64(850000), senderSaldo.TotalBalance)

	receiverSaldo, _ := s.saldoRepo.FindByCardNumber(ctx, s.receiverCardNumber)
	s.Equal(int64(150000), receiverSaldo.TotalBalance)
}

func (s *TransferGapiTestSuite) Test5_TrashedTransfer() {
	s.Require().NotZero(s.transferID)
	ctx := context.Background()

	_, err := s.commandClient.TrashedTransfer(ctx, &pb.FindByIdTransferRequest{
		TransferId: int32(s.transferID),
	})
	s.Require().NoError(err)
}

func (s *TransferGapiTestSuite) Test6_RestoreTransfer() {
	s.Require().NotZero(s.transferID)
	ctx := context.Background()

	_, err := s.commandClient.RestoreTransfer(ctx, &pb.FindByIdTransferRequest{
		TransferId: int32(s.transferID),
	})
	s.Require().NoError(err)
}

func (s *TransferGapiTestSuite) Test7_PermanentDeleteTransfer() {
	s.Require().NotZero(s.transferID)
	ctx := context.Background()

	_, err := s.commandClient.DeleteTransferPermanent(ctx, &pb.FindByIdTransferRequest{
		TransferId: int32(s.transferID),
	})
	s.Require().NoError(err)
}

func TestTransferGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransferGapiTestSuite))
}
