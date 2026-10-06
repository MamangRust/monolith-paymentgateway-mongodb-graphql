package saldo_test

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	pbstats "github.com/MamangRust/monolith-payment-gateway-pb/saldo/stats"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	saldo_handler "github.com/MamangRust/monolith-payment-gateway-saldo/handler"
	saldo_repository "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	saldo_service "github.com/MamangRust/monolith-payment-gateway-saldo/service"
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

type SaldoStatsHandlerApiTestSuite struct {
	suite.Suite
	ts         *tests.TestSuite
	grpcServer *grpc.Server
	conn       *grpc.ClientConn
	gql        http.Handler
	userID     int32
	cardNumber string
	testYear   int
}

func (s *SaldoStatsHandlerApiTestSuite) SetupSuite() {
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

	repos := saldo_repository.NewRepositories(mongoDB, nil, nil)
	svc := saldo_service.NewService(&saldo_service.Deps{
		Repositories: repos,
		Logger:       log,
		Cache:        cacheStore,
		DB:           mongoDB,
	})
	h := saldo_handler.NewHandler(svc)

	s.grpcServer = grpc.NewServer()
	pbstats.RegisterSaldoStatsBalanceServiceServer(s.grpcServer, h)
	pbstats.RegisterSaldoStatsTotalBalanceServer(s.grpcServer, h)

	addr, err := tests.RunGRPCServer(s.grpcServer)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.gql = tests.NewGraphQLHandler(map[string]*grpc.ClientConn{"saldo": conn}, redisClient, log)

	s.testYear = time.Now().Year()

	ctx := context.Background()
	_, err = mongoDB.Collection("users").InsertOne(ctx, bson.M{
		"firstname": "SaldoApi", "lastname": "Stats", "email": "saldo_api_stats@example.com",
		"password": "pass", "verification_code": "123", "is_verified": true,
	})
	s.Require().NoError(err)

	s.cardNumber = "0000111122223333"
	s.Require().NoError(s.ts.InsertFixture(ctx, "cards", "card_id", "", bson.M{
		"card_number": s.cardNumber, "card_type": "debit", "cvv": "123",
		"card_provider": "visa", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
	}))

	now := time.Now().UTC()
	s.Require().NoError(s.ts.InsertFixture(ctx, "saldos", "saldo_id", "", bson.M{
		"card_number": s.cardNumber, "total_balance": 3000000,
		"created_at": now, "updated_at": now,
	}))
}

func (s *SaldoStatsHandlerApiTestSuite) TearDownSuite() {
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

func (s *SaldoStatsHandlerApiTestSuite) TestFindMonthlySaldoBalances() {
	year := strconv.Itoa(s.testYear)
	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`query {
		findMonthlySaldoBalances(input: { year: %s }) {
			status message
			data { month total_balance }
		}
	}`, year))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	s.Require().NotNil(tests.GraphQLOpData(body, "findMonthlySaldoBalances"))
}

func TestSaldoStatsHandlerApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(SaldoStatsHandlerApiTestSuite))
}
