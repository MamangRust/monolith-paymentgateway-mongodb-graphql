package transaction_test

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	card_repo "github.com/MamangRust/monolith-payment-gateway-card/repository"
	merchant_repo "github.com/MamangRust/monolith-payment-gateway-merchant/repository"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/transaction/stats"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	saldo_repo "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/MamangRust/monolith-payment-gateway-transaction/handler"
	"github.com/MamangRust/monolith-payment-gateway-transaction/repository"
	"github.com/MamangRust/monolith-payment-gateway-transaction/service"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TransactionStatsHandlerApiTestSuite struct {
	suite.Suite
	ts         *tests.TestSuite
	grpcServer *grpc.Server
	conn       *grpc.ClientConn
	gql        http.Handler
	harness    *tests.DependencyHarness
	userID     int32
	cardNumber string
	merchantID int32
	testYear   int
	testMonth  int
}

func (s *TransactionStatsHandlerApiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	mongoDB, err := s.ts.MongoDB()
	s.Require().NoError(err)

	realSaldo := &realSaldoRepo{repo: saldo_repo.NewRepositories(mongoDB, nil, nil)}
	realCard := &realCardRepo{
		query:   card_repo.NewCardQueryRepository(mongoDB.Collection("cards")),
		command: card_repo.NewCardCommandRepository(mongoDB.Collection("cards")),
	}
	realMerchant := &realMerchantRepo{repo: merchant_repo.NewMerchantQueryRepository(mongoDB.Collection("merchants"))}

	repos := repository.NewRepositoriesWithAdapters(mongoDB, realSaldo, realCard, realMerchant)

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	redisClient := redis.NewClient(opts)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheStore := cache.NewCacheStore(redisClient, log, &dummyCacheMetrics{})

	harness, err := tests.NewDependencyHarness(mongoDB, cacheStore, log)
	s.Require().NoError(err)
	s.harness = harness

	svc := service.NewService(&service.Deps{
		CardAdapter:     harness.CardAdapter,
		SaldoAdapter:    harness.SaldoAdapter,
		MerchantAdapter: harness.MerchantAdapter,
		DB:              mongoDB,
		Kafka:           nil,
		Repositories:    repos,
		Logger:          log,
		Cache:           cacheStore,
	})

	h := handler.NewHandler(svc)

	s.grpcServer = grpc.NewServer()
	pb.RegisterTransactionStatsStatusServiceServer(s.grpcServer, h)

	addr, err := tests.RunGRPCServer(s.grpcServer)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.gql = tests.NewGraphQLHandler(map[string]*grpc.ClientConn{"transaction": conn}, redisClient, log)

	s.testYear = time.Now().Year()
	s.testMonth = int(time.Now().Month())

	// Seed Data
	ctx := context.Background()
	s.userID = 9110
	_, err = mongoDB.Collection("users").InsertOne(ctx, bson.M{
		"user_id":           s.userID,
		"firstname":         "TransactionApi",
		"lastname":          "Stats",
		"email":             "transaction_api_stats@example.com",
		"password":          "pass",
		"verification_code": "123",
		"is_verified":       true,
	})
	s.Require().NoError(err)

	s.cardNumber = "5555444433332222"
	_, err = mongoDB.Collection("cards").InsertOne(ctx, bson.M{
		"card_id":       s.userID,
		"card_number":   s.cardNumber,
		"card_type":     "debit",
		"cvv":           "123",
		"card_provider": "visa",
		"expire_date":   time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	s.Require().NoError(err)

	s.merchantID = 9111
	_, err = mongoDB.Collection("merchants").InsertOne(ctx, bson.M{
		"merchant_id": s.merchantID,
		"merchant_no": uuid.New(),
		"name":        "Merchant Api",
		"api_key":     "test_key_api",
		"status":      "active",
	})
	s.Require().NoError(err)

	_, err = mongoDB.Collection("transactions").InsertOne(ctx, bson.M{
		"card_number":      s.cardNumber,
		"merchant_id":      s.merchantID,
		"amount":           300000,
		"payment_method":   "ewallet",
		"transaction_time": time.Date(s.testYear, time.Month(s.testMonth), 10, 10, 0, 0, 0, time.UTC),
		"status":           "success",
	})
	s.Require().NoError(err)
}

func (s *TransactionStatsHandlerApiTestSuite) TearDownSuite() {
	if s.harness != nil {
		s.harness.Close()
	}
	if s.conn != nil {
		s.conn.Close()
	}
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
	if s.ts != nil {
		s.ts.Teardown()
	}
}

func (s *TransactionStatsHandlerApiTestSuite) TestFindMonthlyTransactionStatusSuccess() {
	month := strconv.Itoa(s.testMonth)
	year := strconv.Itoa(s.testYear)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`query {
		findMonthlyTransactionStatusSuccess(input: { year: %s, month: %s }) {
			status message
			data { month }
		}
	}`, year, month))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	opRes := tests.GraphQLOpData(body, "findMonthlyTransactionStatusSuccess")
	s.Require().NotNil(opRes)
	s.Equal("success", opRes["status"])
}

func TestTransactionStatsHandlerApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransactionStatsHandlerApiTestSuite))
}
