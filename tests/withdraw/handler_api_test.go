package withdraw_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	card_repo "github.com/MamangRust/monolith-payment-gateway-card/repository"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/withdraw"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	saldo_repo "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	user_repo "github.com/MamangRust/monolith-payment-gateway-user/repository"
	"github.com/MamangRust/monolith-payment-gateway-withdraw/handler"
	"github.com/MamangRust/monolith-payment-gateway-withdraw/repository"
	"github.com/MamangRust/monolith-payment-gateway-withdraw/service"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type WithdrawHandlerTestSuite struct {
	suite.Suite
	ts                 *tests.TestSuite
	harness            *tests.DependencyHarness
	grpcServer         *grpc.Server
	conn               *grpc.ClientConn
	gql                http.Handler
	repos              repository.Repositories
	cardRepo           card_repo.CardCommandRepository
	saldoRepo          saldo_repo.Repositories
	userRepo           user_repo.UserCommandRepository
	customerCardNumber string
	withdrawID         int
}

func (s *WithdrawHandlerTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	mongoDB, err := s.ts.MongoDB()
	s.Require().NoError(err)

	// Repositories for seeding and service dependencies
	userRepos := user_repo.NewUserCommandRepository(mongoDB.Collection("users"))
	cardRepos := card_repo.NewRepositories(mongoDB, nil)
	saldoRepos := saldo_repo.NewRepositories(mongoDB, nil, nil)

	s.userRepo = userRepos
	s.cardRepo = cardRepos.CardCommand
	s.saldoRepo = saldoRepos
	s.repos = repository.NewRepositoriesWithAdapters(mongoDB, &realCardRepo{query: cardRepos.CardQuery}, &realSaldoRepo{repo: saldoRepos})

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

	withdrawService := service.NewService(&service.Deps{
		Kafka:        nil,
		Repositories: s.repos,
		CardAdapter:  harness.CardAdapter,
		SaldoAdapter: harness.SaldoAdapter,
		Logger:       log,
		Cache:        cacheStore,
		DB:           mongoDB,
	})

	// Seed Customer
	customer, err := s.userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Withdraw",
		LastName:  "Customer",
		Email:     fmt.Sprintf("withdraw.%d@test.com", time.Now().UnixNano()),
		Password:  "password123",
	})
	s.Require().NoError(err)

	card, err := s.cardRepo.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID:       int(customer.UserID),
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(1, 0, 0),
		CVV:          "999",
		CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.customerCardNumber = card.CardNumber

	_, err = s.saldoRepo.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{
		CardNumber:   s.customerCardNumber,
		TotalBalance: 1000000,
	})
	s.Require().NoError(err)

	// Start the withdraw gRPC service
	withdrawHandler := handler.NewHandler(withdrawService)

	s.grpcServer = grpc.NewServer()
	pb.RegisterWithdrawCommandServiceServer(s.grpcServer, withdrawHandler)
	pb.RegisterWithdrawQueryServiceServer(s.grpcServer, withdrawHandler)

	addr, err := tests.RunGRPCServer(s.grpcServer)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.gql = tests.NewGraphQLHandler(map[string]*grpc.ClientConn{"withdraw": conn}, redisClient, log)
}

func (s *WithdrawHandlerTestSuite) TearDownSuite() {
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

func (s *WithdrawHandlerTestSuite) Test1_CreateWithdraw() {
	today := time.Now().Format("2006-01-02")

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation {
		createWithdraw(input: { card_number: "%s", withdraw_amount: 100000, withdraw_time: "%s" }) {
			status message
			data { withdraw_id card_number withdraw_amount }
		}
	}`, s.customerCardNumber, today))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data := tests.GraphQLOpEnvelopeData(body, "createWithdraw")
	s.Require().NotNil(data)
	s.Equal(float64(100000), data["withdraw_amount"])
	s.withdrawID = int(data["withdraw_id"].(float64))
	s.Require().NotZero(s.withdrawID)

	// Verify balance
	customerSaldo, _ := s.saldoRepo.FindByCardNumber(context.Background(), s.customerCardNumber)
	s.Equal(int64(900000), customerSaldo.TotalBalance)
}

func (s *WithdrawHandlerTestSuite) Test2_FindWithdrawById() {
	s.Require().NotZero(s.withdrawID)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`	query { findByIdWithdraw(input: { withdraw_id: %d }) { status message data { withdraw_id withdraw_amount } } }`, s.withdrawID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data := tests.GraphQLOpEnvelopeData(body, "findByIdWithdraw")
	s.Require().NotNil(data)
	s.Equal(float64(s.withdrawID), data["withdraw_id"])
}

func (s *WithdrawHandlerTestSuite) Test3_FindAllWithdraws() {
	body, err := tests.DoGraphQL(s.gql, `	query { findAllWithdraw(input: { page: 1, page_size: 10 }) { status message data { withdraw_id withdraw_amount } } }`)
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	s.Require().NotNil(tests.GraphQLOpData(body, "findAllWithdraw"))
}

func (s *WithdrawHandlerTestSuite) Test4_UpdateWithdraw() {
	s.Require().NotZero(s.withdrawID)
	today := time.Now().Format("2006-01-02")

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation {
		updateWithdraw(input: { withdraw_id: %d, card_number: "%s", withdraw_amount: 150000, withdraw_time: "%s" }) {
			status message
			data { withdraw_id withdraw_amount }
		}
	}`, s.withdrawID, s.customerCardNumber, today))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data := tests.GraphQLOpEnvelopeData(body, "updateWithdraw")
	s.Require().NotNil(data)
	s.Equal(float64(150000), data["withdraw_amount"])

	// Verify adjusted balance (900k - 50k = 850k)
	customerSaldo, _ := s.saldoRepo.FindByCardNumber(context.Background(), s.customerCardNumber)
	s.Equal(int64(850000), customerSaldo.TotalBalance)
}

func (s *WithdrawHandlerTestSuite) Test5_TrashedWithdraw() {
	s.Require().NotZero(s.withdrawID)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`	mutation { trashedWithdraw(input: { withdraw_id: %d }) { status message data { withdraw_id } } }`, s.withdrawID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))

	// FindTrashed
	body, err = tests.DoGraphQL(s.gql, `	query { findByTrashedWithdraw(input: { page: 1, page_size: 10 }) { status message data { withdraw_id } } }`)
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	s.Require().NotNil(tests.GraphQLOpData(body, "findByTrashedWithdraw"))
}

func (s *WithdrawHandlerTestSuite) Test6_RestoreWithdraw() {
	s.Require().NotZero(s.withdrawID)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`	mutation { restoreWithdraw(input: { withdraw_id: %d }) { status message data { withdraw_id } } }`, s.withdrawID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
}

func (s *WithdrawHandlerTestSuite) Test7_PermanentDeleteWithdraw() {
	s.Require().NotZero(s.withdrawID)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation { deleteWithdrawPermanent(input: { withdraw_id: %d }) { status message } }`, s.withdrawID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
}

func TestWithdrawHandlerSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(WithdrawHandlerTestSuite))
}
