package transaction_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	card_repo "github.com/MamangRust/monolith-payment-gateway-card/repository"
	merchant_repo "github.com/MamangRust/monolith-payment-gateway-merchant/repository"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/transaction"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	saldo_repo "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/MamangRust/monolith-payment-gateway-transaction/handler"
	"github.com/MamangRust/monolith-payment-gateway-transaction/repository"
	"github.com/MamangRust/monolith-payment-gateway-transaction/service"
	user_repo "github.com/MamangRust/monolith-payment-gateway-user/repository"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TransactionHandlerTestSuite struct {
	suite.Suite
	ts            *tests.TestSuite
	redisClient   *redis.Client
	grpcServer    *grpc.Server
	commandClient pb.TransactionCommandServiceClient
	queryClient   pb.TransactionQueryServiceClient
	conn          *grpc.ClientConn
	gql           http.Handler

	// Repositories for seeding
	userRepo     user_repo.UserCommandRepository
	cardRepo     card_repo.Repositories
	saldoRepo    saldo_repo.Repositories
	merchantRepo merchant_repo.Repositories

	// harness hosts the in-process card/saldo/user/merchant gRPC services the
	// transaction service resolves over adapters.
	harness *tests.DependencyHarness

	customerCardNumber string
	merchantApiKey     string
	merchantID         int32
	merchantCardNumber string
	transactionID      int32
}

func (s *TransactionHandlerTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	mongoDB, err := s.ts.MongoDB()
	s.Require().NoError(err)

	// Repositories for seeding
	s.userRepo = user_repo.NewUserCommandRepository(mongoDB.Collection("users"))
	s.cardRepo = *card_repo.NewRepositories(mongoDB, nil)
	s.saldoRepo = saldo_repo.NewRepositories(mongoDB, nil, nil)
	s.merchantRepo = merchant_repo.NewRepositories(mongoDB, nil)

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	cardRepoWrapper := &transactionCardRepo{
		query:   s.cardRepo.CardQuery,
		command: s.cardRepo.CardCommand,
	}
	saldoWrapper := &realSaldoRepo{repo: s.saldoRepo}
	merchantWrapper := &realMerchantRepo{repo: s.merchantRepo}
	transactionRepos := repository.NewRepositoriesWithAdapters(mongoDB, saldoWrapper, cardRepoWrapper, merchantWrapper)

	harness, err := tests.NewDependencyHarness(mongoDB, cacheStore, log)
	s.Require().NoError(err)
	s.harness = harness

	transactionService := service.NewService(&service.Deps{
		CardAdapter:     harness.CardAdapter,
		SaldoAdapter:    harness.SaldoAdapter,
		MerchantAdapter: harness.MerchantAdapter,
		DB:              mongoDB,
		Kafka:           nil,
		Repositories:    transactionRepos,
		Logger:          log,
		Cache:           cacheStore,
	})

	transactionHandler := handler.NewHandler(transactionService)

	s.grpcServer = grpc.NewServer()
	pb.RegisterTransactionCommandServiceServer(s.grpcServer, transactionHandler)
	pb.RegisterTransactionQueryServiceServer(s.grpcServer, transactionHandler)

	addr, err := tests.RunGRPCServer(s.grpcServer)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn
	s.commandClient = pb.NewTransactionCommandServiceClient(conn)
	s.queryClient = pb.NewTransactionQueryServiceClient(conn)

	s.gql = tests.NewGraphQLHandler(map[string]*grpc.ClientConn{"transaction": conn}, s.redisClient, log)

	// Seed Customer
	customer, err := s.userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Transaction",
		LastName:  "Customer",
		Email:     fmt.Sprintf("customer.%d@transaction.com", time.Now().UnixNano()),
		Password:  "password123",
	})
	s.Require().NoError(err)

	cCard, err := s.cardRepo.CardCommand.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID:       int(customer.UserID),
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(1, 0, 0),
		CVV:          "123",
		CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.customerCardNumber = cCard.CardNumber

	_, err = s.saldoRepo.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{
		CardNumber:   s.customerCardNumber,
		TotalBalance: 1000000,
	})
	s.Require().NoError(err)

	// Seed Merchant owner (distinct user from the customer)
	owner, err := s.userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Merchant",
		LastName:  "Owner",
		Email:     fmt.Sprintf("merchant.owner.%d@transaction.com", time.Now().UnixNano()),
		Password:  "password123",
	})
	s.Require().NoError(err)

	merchant, err := s.merchantRepo.CreateMerchant(context.Background(), &requests.CreateMerchantRequest{
		UserID: int(owner.UserID),
		Name:   "Transaction Merchant",
	})
	s.Require().NoError(err)
	s.merchantID = merchant.MerchantID
	s.merchantApiKey = merchant.ApiKey

	_, err = s.merchantRepo.UpdateMerchantStatus(context.Background(), &requests.UpdateMerchantStatusRequest{
		MerchantID: intPtr(int(merchant.MerchantID)),
		Status:     "active",
	})
	s.Require().NoError(err)

	mCard, err := s.cardRepo.CardCommand.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID:       int(owner.UserID),
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(1, 0, 0),
		CVV:          "321",
		CardProvider: "mastercard",
	})
	s.Require().NoError(err)
	s.merchantCardNumber = mCard.CardNumber

	_, err = s.saldoRepo.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{
		CardNumber:   s.merchantCardNumber,
		TotalBalance: 0,
	})
	s.Require().NoError(err)
}

func (s *TransactionHandlerTestSuite) TearDownSuite() {
	if s.harness != nil {
		s.harness.Close()
	}
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

func (s *TransactionHandlerTestSuite) Test1_CreateTransaction() {
	// createTransaction/updateTransaction are merchant-Kafka-gated at the
	// gateway, so the record is created through the gRPC command service.
	ctx := context.Background()
	res, err := s.commandClient.CreateTransaction(ctx, &pb.CreateTransactionRequest{
		ApiKey:          s.merchantApiKey,
		CardNumber:      s.customerCardNumber,
		Amount:          50000,
		PaymentMethod:   "visa",
		MerchantId:      s.merchantID,
		TransactionTime: timestamppb.New(time.Now()),
	})
	s.Require().NoError(err)
	s.Require().NotNil(res.Data)
	s.transactionID = res.Data.Id
	s.Require().NotZero(s.transactionID)

	// Verify balances
	customerSaldo, _ := s.saldoRepo.FindByCardNumber(ctx, s.customerCardNumber)
	s.Equal(int64(950000), customerSaldo.TotalBalance)

	merchantSaldo, _ := s.saldoRepo.FindByCardNumber(ctx, s.merchantCardNumber)
	s.Equal(int64(50000), merchantSaldo.TotalBalance)
}

func (s *TransactionHandlerTestSuite) Test2_FindTransactionById() {
	s.Require().NotZero(s.transactionID)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`query { findByIdTransaction(input: { transaction_id: %d }) { status message data { id amount } } }`, s.transactionID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data := tests.GraphQLOpEnvelopeData(body, "findByIdTransaction")
	s.Require().NotNil(data)
	s.Equal(float64(s.transactionID), data["id"])
	s.Equal(float64(50000), data["amount"])
}

func (s *TransactionHandlerTestSuite) Test3_FindAllTransactions() {
	body, err := tests.DoGraphQL(s.gql, `query { findAllTransaction(input: { page: 1, page_size: 10 }) { status message data { id amount } } }`)
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	s.Require().NotNil(tests.GraphQLOpData(body, "findAllTransaction"))
}

func (s *TransactionHandlerTestSuite) Test4_UpdateTransaction() {
	s.Require().NotZero(s.transactionID)
	ctx := context.Background()

	_, err := s.commandClient.UpdateTransaction(ctx, &pb.UpdateTransactionRequest{
		ApiKey:          s.merchantApiKey,
		TransactionId:   s.transactionID,
		CardNumber:      s.customerCardNumber,
		Amount:          60000,
		PaymentMethod:   "visa",
		MerchantId:      s.merchantID,
		TransactionTime: timestamppb.New(time.Now()),
	})
	s.Require().NoError(err)

	// Verify adjusted balance (950k - 10k = 940k)
	customerSaldo, _ := s.saldoRepo.FindByCardNumber(ctx, s.customerCardNumber)
	s.Equal(int64(940000), customerSaldo.TotalBalance)
}

func (s *TransactionHandlerTestSuite) Test5_TrashedTransaction() {
	s.Require().NotZero(s.transactionID)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation { trashedTransaction(input: { transaction_id: %d }) { status message data { id } } }`, s.transactionID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))

	// FindTrashed
	body, err = tests.DoGraphQL(s.gql, `query { findByTrashedTransaction(input: { page: 1, page_size: 10 }) { status message data { id } } }`)
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	s.Require().NotNil(tests.GraphQLOpData(body, "findByTrashedTransaction"))
}

func (s *TransactionHandlerTestSuite) Test6_RestoreTransaction() {
	s.Require().NotZero(s.transactionID)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation { restoreTransaction(input: { transaction_id: %d }) { status message data { id } } }`, s.transactionID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
}

func (s *TransactionHandlerTestSuite) Test7_PermanentDeleteTransaction() {
	s.Require().NotZero(s.transactionID)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation { deleteTransactionPermanent(input: { transaction_id: %d }) { status message } }`, s.transactionID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
}

func intPtr(v int) *int { return &v }

func TestTransactionHandlerSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransactionHandlerTestSuite))
}
