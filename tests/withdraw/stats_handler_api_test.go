package withdraw_test

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	pb "github.com/MamangRust/monolith-payment-gateway-pb/withdraw/stats"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/MamangRust/monolith-payment-gateway-withdraw/handler"
	"github.com/MamangRust/monolith-payment-gateway-withdraw/repository"
	"github.com/MamangRust/monolith-payment-gateway-withdraw/service"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	card_repo "github.com/MamangRust/monolith-payment-gateway-card/repository"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	saldo_repo "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	"go.mongodb.org/mongo-driver/bson"
)

type WithdrawStatsHandlerApiTestSuite struct {
	suite.Suite
	ts         *tests.TestSuite
	grpcServer *grpc.Server
	conn       *grpc.ClientConn
	gql        http.Handler
	userID     int32
	cardNumber string
	testYear   int
	testMonth  int
}

func (s *WithdrawStatsHandlerApiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	mongoDB, err := s.ts.MongoDB()
	s.Require().NoError(err)

	realSaldo := &realSaldoRepo{repo: saldo_repo.NewRepositories(mongoDB, nil, nil)}
	realCard := &realCardRepo{
		query: card_repo.NewCardQueryRepository(mongoDB.Collection("cards")),
	}

	repos := repository.NewRepositoriesWithAdapters(mongoDB, realCard, realSaldo)

	zapLog := zap.NewNop()
	myLogger := &logger.Logger{Log: zapLog}

	redisOption, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	redisClient := redis.NewClient(redisOption)
	cacheStore := cache.NewCacheStore(redisClient, myLogger, &dummyCacheMetrics{})

	svc := service.NewService(&service.Deps{
		Kafka:        nil,
		Repositories: repos,
		Logger:       myLogger,
		Cache:        cacheStore,
		DB:           mongoDB,
	})

	h := handler.NewHandler(svc)

	s.grpcServer = grpc.NewServer()
	pb.RegisterWithdrawStatsStatusServiceServer(s.grpcServer, h)
	pb.RegisterWithdrawStatsAmountServiceServer(s.grpcServer, h)

	addr, err := tests.RunGRPCServer(s.grpcServer)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.gql = tests.NewGraphQLHandler(map[string]*grpc.ClientConn{"withdraw": conn}, redisClient, myLogger)

	s.testYear = time.Now().Year()
	s.testMonth = int(time.Now().Month())

	ctx := context.Background()
	_, err = mongoDB.Collection("users").InsertOne(ctx, bson.M{"user_id": 1, "firstname": "WithdrawApi", "lastname": "Stats", "email": "withdraw_api_stats@example.com", "password": "pass", "verification_code": "123", "is_verified": true})
	s.Require().NoError(err)
	s.userID = 1

	s.cardNumber = "9999000011112222"
	_, err = mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_id": 1, "card_number": s.cardNumber, "card_type": "debit", "cvv": "123", "card_provider": "visa", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)

	_, err = mongoDB.Collection("withdraws").InsertOne(ctx, bson.M{"card_number": s.cardNumber, "withdraw_amount": 300000, "withdraw_time": time.Date(s.testYear, time.Month(s.testMonth), 10, 10, 0, 0, 0, time.UTC), "status": "success"})
	s.Require().NoError(err)
}

func (s *WithdrawStatsHandlerApiTestSuite) TearDownSuite() {
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

func (s *WithdrawStatsHandlerApiTestSuite) TestFindMonthlyWithdrawStatusSuccess() {
	month := strconv.Itoa(s.testMonth)
	year := strconv.Itoa(s.testYear)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`query {
		findMonthlyWithdrawStatusSuccess(input: { year: %s, month: %s }) {
			status message
			data { month }
		}
	}`, year, month))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	opRes := tests.GraphQLOpData(body, "findMonthlyWithdrawStatusSuccess")
	s.Require().NotNil(opRes)
	s.Equal("success", opRes["status"])
	payload, ok := opRes["data"].([]any)
	s.Require().True(ok, "expected monthly status data array, got %+v", opRes["data"])
	s.Require().NotEmpty(payload)
}

func (s *WithdrawStatsHandlerApiTestSuite) TestFindYearlyWithdrawStatusSuccess() {
	year := strconv.Itoa(s.testYear)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`query {
		findYearlyWithdrawStatusSuccess(input: { year: %s }) {
			status message
			data { year }
		}
	}`, year))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	opRes := tests.GraphQLOpData(body, "findYearlyWithdrawStatusSuccess")
	s.Require().NotNil(opRes)
	s.Equal("success", opRes["status"])
	payload, ok := opRes["data"].([]any)
	s.Require().True(ok, "expected yearly status data array, got %+v", opRes["data"])
	s.Require().NotEmpty(payload)
}

func (s *WithdrawStatsHandlerApiTestSuite) TestFindMonthlyWithdrawAmount() {
	year := strconv.Itoa(s.testYear)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`query {
		findMonthlyWithdraws(input: { year: %s }) {
			status message
			data { month total_amount }
		}
	}`, year))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	opRes := tests.GraphQLOpData(body, "findMonthlyWithdraws")
	s.Require().NotNil(opRes)
	s.Equal("success", opRes["status"])
	payload, ok := opRes["data"].([]any)
	s.Require().True(ok, "expected monthly withdraw amount data array, got %+v", opRes["data"])
	s.Require().NotEmpty(payload)
}

func (s *WithdrawStatsHandlerApiTestSuite) TestFindMonthlyWithdrawAmountByCardNumber() {
	year := strconv.Itoa(s.testYear)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`query {
		findMonthlyWithdrawsByCardNumber(input: { year: %s, card_number: "%s" }) {
			status message
			data { month total_amount }
		}
	}`, year, s.cardNumber))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	opRes := tests.GraphQLOpData(body, "findMonthlyWithdrawsByCardNumber")
	s.Require().NotNil(opRes)
	s.Equal("success", opRes["status"])
	payload, ok := opRes["data"].([]any)
	s.Require().True(ok, "expected monthly withdraw amount-by-card data array, got %+v", opRes["data"])
	s.Require().NotEmpty(payload)
}

func TestWithdrawStatsHandlerApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(WithdrawStatsHandlerApiTestSuite))
}
