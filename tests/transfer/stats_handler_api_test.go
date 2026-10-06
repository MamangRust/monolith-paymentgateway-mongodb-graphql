package transfer_test

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	pb "github.com/MamangRust/monolith-payment-gateway-pb/transfer/stats"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/MamangRust/monolith-payment-gateway-transfer/handler"
	"github.com/MamangRust/monolith-payment-gateway-transfer/repository"
	"github.com/MamangRust/monolith-payment-gateway-transfer/service"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	card_repo "github.com/MamangRust/monolith-payment-gateway-card/repository"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	saldo_repo "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type TransferStatsHandlerApiTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	grpcServer  *grpc.Server
	conn        *grpc.ClientConn
	gql         http.Handler
	userID      int32
	cardNumber1 string
	cardNumber2 string
	testYear    int
	testMonth   int
}

func (s *TransferStatsHandlerApiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	mongoDB, err := s.ts.MongoDB()
	s.Require().NoError(err)

	realSaldo := &realSaldoRepo{repo: saldo_repo.NewRepositories(mongoDB, nil, nil)}
	realCard := &realCardRepo{
		query: card_repo.NewCardQueryRepository(mongoDB.Collection("cards")),
	}

	repos := repository.NewRepositoriesWithAdapters(mongoDB, realSaldo, realCard)

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
	pb.RegisterTransferStatsStatusServiceServer(s.grpcServer, h)

	addr, err := tests.RunGRPCServer(s.grpcServer)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.gql = tests.NewGraphQLHandler(map[string]*grpc.ClientConn{"transfer": conn}, redisClient, myLogger)

	s.testYear = time.Now().Year()
	s.testMonth = int(time.Now().Month())

	ctx := context.Background()
	_, err = mongoDB.Collection("users").InsertOne(ctx, bson.M{"user_id": 1, "firstname": "TransferApi", "lastname": "Stats", "email": "transfer_api_stats@example.com", "password": "pass", "verification_code": "123", "is_verified": true})
	s.Require().NoError(err)
	s.userID = 1

	s.cardNumber1 = "2222111100009999"
	s.cardNumber2 = "3333444455556666"
	_, err = mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_id": 1, "card_number": s.cardNumber1, "card_type": "debit", "cvv": "123", "card_provider": "visa", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)
	_, err = mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_id": 2, "card_number": s.cardNumber2, "card_type": "debit", "cvv": "123", "card_provider": "mastercard", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)

	_, err = mongoDB.Collection("transfers").InsertOne(ctx, bson.M{"transfer_id": 1, "transfer_no": uuid.New(), "transfer_from": s.cardNumber1, "transfer_to": s.cardNumber2, "transfer_amount": 250000, "transfer_time": time.Date(s.testYear, time.Month(s.testMonth), 10, 10, 0, 0, 0, time.UTC), "status": "success"})
	s.Require().NoError(err)
}

func (s *TransferStatsHandlerApiTestSuite) TearDownSuite() {
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

func (s *TransferStatsHandlerApiTestSuite) TestFindMonthlyTransferStatusSuccess() {
	month := strconv.Itoa(s.testMonth)
	year := strconv.Itoa(s.testYear)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`query {
		findMonthlyTransferStatusSuccess(input: { year: %s, month: %s }) {
			status message
			data { month }
		}
	}`, year, month))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	opRes := tests.GraphQLOpData(body, "findMonthlyTransferStatusSuccess")
	s.Require().NotNil(opRes)
	s.Equal("success", opRes["status"])
}

func TestTransferStatsHandlerApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransferStatsHandlerApiTestSuite))
}
