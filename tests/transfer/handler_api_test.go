package transfer_test

import (
	"context"
	"fmt"
	"net/http"
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

type TransferHandlerTestSuite struct {
	suite.Suite
	ts                 *tests.TestSuite
	harness            *tests.DependencyHarness
	grpcServer         *grpc.Server
	conn               *grpc.ClientConn
	gql                http.Handler
	repos              repository.Repositories
	userRepo           user_repo.UserCommandRepository
	cardRepo           card_repo.Repositories
	saldoRepo          saldo_repo.Repositories
	senderCardNumber   string
	receiverCardNumber string
	transferID         int
}

func (s *TransferHandlerTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	mongoDB, err := s.ts.MongoDB()
	s.Require().NoError(err)

	// Repositories for seeding
	s.userRepo = user_repo.NewUserCommandRepository(mongoDB.Collection("users"))
	s.cardRepo = *card_repo.NewRepositories(mongoDB, nil)
	s.saldoRepo = saldo_repo.NewRepositories(mongoDB, nil, nil)

	// Transfer repos
	s.repos = repository.NewRepositoriesWithAdapters(mongoDB, &realSaldoRepo{repo: s.saldoRepo}, &realCardRepo{query: s.cardRepo.CardQuery})

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	redisClient := redis.NewClient(opts)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(redisClient, log, cacheMetrics)

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
		LastName:  "User",
		Email:     fmt.Sprintf("sender.%d@transfer.com", time.Now().UnixNano()),
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
		LastName:  "User",
		Email:     fmt.Sprintf("receiver.%d@transfer.com", time.Now().UnixNano()),
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

	// Start the transfer gRPC service
	transferHandlerGapi := handler.NewHandler(transferService)

	s.grpcServer = grpc.NewServer()
	pb.RegisterTransferCommandServiceServer(s.grpcServer, transferHandlerGapi)
	pb.RegisterTransferQueryServiceServer(s.grpcServer, transferHandlerGapi)

	addr, err := tests.RunGRPCServer(s.grpcServer)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.gql = tests.NewGraphQLHandler(map[string]*grpc.ClientConn{"transfer": conn}, redisClient, log)
}

func (s *TransferHandlerTestSuite) TearDownSuite() {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.harness != nil {
		s.harness.Close()
	}
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
	if s.ts != nil {
		s.ts.Teardown()
	}
}

func (s *TransferHandlerTestSuite) Test1_CreateTransfer() {
	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation {
		createTransfer(input: { transfer_from: "%s", transfer_to: "%s", transfer_amount: 100000 }) {
			status message
			data { id transfer_from transfer_to transfer_amount }
		}
	}`, s.senderCardNumber, s.receiverCardNumber))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data := tests.GraphQLOpEnvelopeData(body, "createTransfer")
	s.Require().NotNil(data)
	s.Equal(float64(100000), data["transfer_amount"])
	s.transferID = int(data["id"].(float64))
	s.Require().NotZero(s.transferID)

	// Verify balances
	senderSaldo, _ := s.saldoRepo.FindByCardNumber(context.Background(), s.senderCardNumber)
	s.Equal(int64(900000), senderSaldo.TotalBalance)

	receiverSaldo, _ := s.saldoRepo.FindByCardNumber(context.Background(), s.receiverCardNumber)
	s.Equal(int64(100000), receiverSaldo.TotalBalance)
}

func (s *TransferHandlerTestSuite) Test2_FindTransferById() {
	s.Require().NotZero(s.transferID)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`query { findByIdTransfer(input: { transfer_id: %d }) { status message data { id transfer_amount } } }`, s.transferID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data := tests.GraphQLOpEnvelopeData(body, "findByIdTransfer")
	s.Require().NotNil(data)
	s.Equal(float64(s.transferID), data["id"])
}

func (s *TransferHandlerTestSuite) Test3_FindAllTransfers() {
	body, err := tests.DoGraphQL(s.gql, `query { findAllTransfer(input: { page: 1, page_size: 10 }) { status message data { id transfer_amount } } }`)
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	s.Require().NotNil(tests.GraphQLOpData(body, "findAllTransfer"))
}

func (s *TransferHandlerTestSuite) Test4_UpdateTransfer() {
	s.Require().NotZero(s.transferID)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation {
		updateTransfer(input: { transfer_id: %d, transfer_from: "%s", transfer_to: "%s", transfer_amount: 150000 }) {
			status message
			data { id transfer_amount }
		}
	}`, s.transferID, s.senderCardNumber, s.receiverCardNumber))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data := tests.GraphQLOpEnvelopeData(body, "updateTransfer")
	s.Require().NotNil(data)
	s.Equal(float64(150000), data["transfer_amount"])

	// Verify adjusted balances (Sender 900k - 50k = 850k, Receiver 100k + 50k = 150k)
	senderSaldo, _ := s.saldoRepo.FindByCardNumber(context.Background(), s.senderCardNumber)
	s.Equal(int64(850000), senderSaldo.TotalBalance)

	receiverSaldo, _ := s.saldoRepo.FindByCardNumber(context.Background(), s.receiverCardNumber)
	s.Equal(int64(150000), receiverSaldo.TotalBalance)
}

func (s *TransferHandlerTestSuite) Test5_TrashedTransfer() {
	s.Require().NotZero(s.transferID)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation { trashedTransfer(input: { transfer_id: %d }) { status message data { id } } }`, s.transferID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))

	// FindTrashed
	body, err = tests.DoGraphQL(s.gql, `query { findByTrashedTransfer(input: { page: 1, page_size: 10 }) { status message data { id } } }`)
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	s.Require().NotNil(tests.GraphQLOpData(body, "findByTrashedTransfer"))
}

func (s *TransferHandlerTestSuite) Test6_RestoreTransfer() {
	s.Require().NotZero(s.transferID)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation { restoreTransfer(input: { transfer_id: %d }) { status message data { id } } }`, s.transferID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
}

func (s *TransferHandlerTestSuite) Test7_PermanentDeleteTransfer() {
	s.Require().NotZero(s.transferID)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation { deleteTransferPermanent(input: { transfer_id: %d }) { status message } }`, s.transferID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
}

func TestTransferHandlerSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransferHandlerTestSuite))
}
