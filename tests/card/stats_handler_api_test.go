package card_test

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	card_handler "github.com/MamangRust/monolith-payment-gateway-card/handler"
	card_repository "github.com/MamangRust/monolith-payment-gateway-card/repository"
	card_service "github.com/MamangRust/monolith-payment-gateway-card/service"
	pbcard "github.com/MamangRust/monolith-payment-gateway-pb/card"
	pbstats "github.com/MamangRust/monolith-payment-gateway-pb/card/stats"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CardStatsApiTestSuite struct {
	suite.Suite
	ts         *tests.TestSuite
	grpcServer *grpc.Server
	conn       *grpc.ClientConn
	gql        http.Handler
	cardNumber string
	testYear   int
}

func (s *CardStatsApiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	redisClient := redis.NewClient(opts)

	mongoDB, err := s.ts.MongoDB()
	s.Require().NoError(err)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(redisClient, log, cacheMetrics)

	repos := card_repository.NewRepositories(mongoDB, nil)
	cardSvc := card_service.NewService(&card_service.Deps{
		Repositories: repos,
		Logger:       log,
		Cache:        cacheStore,
		Kafka:        nil,
	})
	cardH := card_handler.NewHandler(cardSvc)

	// Setup gRPC server
	s.grpcServer = grpc.NewServer()
	pbcard.RegisterCardQueryServiceServer(s.grpcServer, cardH)
	pbcard.RegisterCardCommandServiceServer(s.grpcServer, cardH)
	pbstats.RegisterCardStatsBalanceServiceServer(s.grpcServer, cardH)
	pbstats.RegisterCardStatsTopupServiceServer(s.grpcServer, cardH)
	pbstats.RegisterCardStatsTransferServiceServer(s.grpcServer, cardH)
	pbstats.RegisterCardStatsWithdrawServiceServer(s.grpcServer, cardH)
	pbstats.RegisterCardStatsTransactionServiceServer(s.grpcServer, cardH)

	addr, err := tests.RunGRPCServer(s.grpcServer)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.gql = tests.NewGraphQLHandler(map[string]*grpc.ClientConn{"card": conn}, redisClient, log)

	s.testYear = time.Now().Year()

	// Seed data
	ctx := context.Background()
	s.cardNumber = "4444555566667777"

	_, err = mongoDB.Collection("users").InsertOne(ctx, bson.M{
		"firstname": "Api", "lastname": "Stats", "email": "api_stats_card@example.com",
		"password": "pass", "verification_code": "123", "is_verified": true,
	})
	s.Require().NoError(err)

	s.Require().NoError(s.ts.InsertFixture(ctx, "cards", "card_id", "", bson.M{
		"card_number": s.cardNumber, "card_type": "debit", "cvv": "123",
		"card_provider": "visa", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
	}))

	s.seedHistoricalData()
}

func (s *CardStatsApiTestSuite) seedHistoricalData() {
	ctx := context.Background()
	janBalance := time.Date(s.testYear, 1, 15, 10, 0, 0, 0, time.UTC)
	s.Require().NoError(s.ts.InsertFixture(ctx, "saldos", "saldo_id", "", bson.M{
		"card_number": s.cardNumber, "total_balance": 1000,
		"created_at": janBalance, "updated_at": janBalance,
	}))
	s.Require().NoError(s.ts.InsertFixture(ctx, "topups", "topup_id", "topup_no", bson.M{
		"card_number": s.cardNumber, "topup_amount": 500,
		"topup_time": time.Date(s.testYear, 1, 10, 10, 0, 0, 0, time.UTC),
		"topup_method": "api", "status": "success",
	}))
}

func (s *CardStatsApiTestSuite) TearDownSuite() {
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

func (s *CardStatsApiTestSuite) TestBalanceApi() {
	year := strconv.Itoa(s.testYear)

	// Global monthly balance
	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`query {
		findMonthlyBalance(input: { year: %s }) {
			status message
			data { month total_balance }
		}
	}`, year))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	opRes := tests.GraphQLOpData(body, "findMonthlyBalance")
	s.Require().NotNil(opRes)
	s.Equal("success", opRes["status"])
	payload, ok := opRes["data"].([]any)
	s.Require().True(ok, "expected monthly balance data array, got %+v", opRes["data"])
	s.Require().NotEmpty(payload)
	firstMonth := payload[0].(map[string]any)
	s.Equal(float64(1000), firstMonth["total_balance"])

	// Monthly balance by card
	body, err = tests.DoGraphQL(s.gql, fmt.Sprintf(`query {
		findMonthlyBalanceByCardNumber(input: { year: %s, card_number: "%s" }) {
			status message
			data { month total_balance }
		}
	}`, year, s.cardNumber))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	opRes = tests.GraphQLOpData(body, "findMonthlyBalanceByCardNumber")
	s.Require().NotNil(opRes)
	s.Equal("success", opRes["status"])
}

func (s *CardStatsApiTestSuite) TestTopupApi() {
	year := strconv.Itoa(s.testYear)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`query {
		findMonthlyTopupAmount(input: { year: %s }) {
			status message
			data { month total_amount }
		}
	}`, year))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	opRes := tests.GraphQLOpData(body, "findMonthlyTopupAmount")
	s.Require().NotNil(opRes)
	s.Equal("success", opRes["status"])
	payload, ok := opRes["data"].([]any)
	s.Require().True(ok, "expected monthly topup data array, got %+v", opRes["data"])
	s.Require().NotEmpty(payload)
	firstMonth := payload[0].(map[string]any)
	s.Equal(float64(500), firstMonth["total_amount"])
}

func TestCardStatsApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(CardStatsApiTestSuite))
}
