package topup_test

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	card_repo "github.com/MamangRust/monolith-payment-gateway-card/repository"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/topup/stats"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	saldo_repo "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/MamangRust/monolith-payment-gateway-topup/handler"
	"github.com/MamangRust/monolith-payment-gateway-topup/repository"
	"github.com/MamangRust/monolith-payment-gateway-topup/service"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TopupStatsHandlerApiTestSuite struct {
	suite.Suite
	ts         *tests.TestSuite
	grpcServer *grpc.Server
	conn       *grpc.ClientConn
	gql        http.Handler
	userID     int32
	cardNumber string
	testYear   int
}

func (s *TopupStatsHandlerApiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	mongoDB, err := s.ts.MongoDB()
	s.Require().NoError(err)

	realSaldo := &realSaldoRepo{repo: saldo_repo.NewRepositories(mongoDB, nil, nil)}
	realCard := &realCardRepo{
		query: card_repo.NewCardQueryRepository(mongoDB.Collection("cards")),
		cmd:   card_repo.NewCardCommandRepository(mongoDB.Collection("cards")),
	}

	repos := repository.NewRepositoriesWithAdapters(mongoDB, realCard, realSaldo)

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	redisClient := redis.NewClient(opts)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheStore := cache.NewCacheStore(redisClient, log, &dummyCacheMetrics{})

	svc := service.NewService(&service.Deps{
		Kafka:        nil,
		Repositories: repos,
		Logger:       log,
		Cache:        cacheStore,
		DB:           mongoDB,
	})

	h := handler.NewHandler(svc)

	s.grpcServer = grpc.NewServer()
	pb.RegisterTopupStatsAmountServiceServer(s.grpcServer, h)
	pb.RegisterTopupStatsMethodServiceServer(s.grpcServer, h)
	pb.RegisterTopupStatsStatusServiceServer(s.grpcServer, h)

	addr, err := tests.RunGRPCServer(s.grpcServer)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.gql = tests.NewGraphQLHandler(map[string]*grpc.ClientConn{"topup": conn}, redisClient, log)

	s.testYear = time.Now().Year()

	// Seed Data
	ctx := context.Background()
	s.userID = 9120
	_, err = mongoDB.Collection("users").InsertOne(ctx, bson.M{
		"user_id":           s.userID,
		"firstname":         "TopupApi",
		"lastname":          "Stats",
		"email":             "topup_api_stats@example.com",
		"password":          "pass",
		"verification_code": "123",
		"is_verified":       true,
	})
	s.Require().NoError(err)

	s.cardNumber = "4444555566667777"
	_, err = mongoDB.Collection("cards").InsertOne(ctx, bson.M{
		"card_id":       s.userID,
		"card_number":   s.cardNumber,
		"card_type":     "debit",
		"cvv":           "123",
		"card_provider": "visa",
		"expire_date":   time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	s.Require().NoError(err)

	_, err = mongoDB.Collection("topups").InsertOne(ctx, bson.M{
		"topup_id":     9131,
		"topup_no":     "seed-topup-9131",
		"card_number":  s.cardNumber,
		"topup_amount": 10000,
		"topup_method": "bank_transfer",
		"topup_time":   time.Date(s.testYear, 1, 10, 10, 0, 0, 0, time.UTC),
		"status":       "success",
	})
	s.Require().NoError(err)
}

func (s *TopupStatsHandlerApiTestSuite) TearDownSuite() {
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

func (s *TopupStatsHandlerApiTestSuite) TestFindMonthlyTopupAmounts() {
	year := strconv.Itoa(s.testYear)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`query {
		findMonthlyTopupAmounts(input: { year: %s }) {
			status message
			data { month total_amount }
		}
	}`, year))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	opRes := tests.GraphQLOpData(body, "findMonthlyTopupAmounts")
	s.Require().NotNil(opRes)
	s.Equal("success", opRes["status"])
	payload, ok := opRes["data"].([]any)
	s.Require().True(ok, "expected monthly topup amounts data array, got %+v", opRes["data"])
	s.Require().NotEmpty(payload)
}

func (s *TopupStatsHandlerApiTestSuite) TestFindMonthlyTopupAmountsByCard() {
	year := strconv.Itoa(s.testYear)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`query {
		findMonthlyTopupAmountsByCardNumber(input: { year: %s, card_number: "%s" }) {
			status message
			data { month total_amount }
		}
	}`, year, s.cardNumber))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	opRes := tests.GraphQLOpData(body, "findMonthlyTopupAmountsByCardNumber")
	s.Require().NotNil(opRes)
	s.Equal("success", opRes["status"])
}

func TestTopupStatsHandlerApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TopupStatsHandlerApiTestSuite))
}
