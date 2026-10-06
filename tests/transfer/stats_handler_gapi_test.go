package transfer_test

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"net"
	"testing"
	"time"

	pbtransfer "github.com/MamangRust/monolith-payment-gateway-pb/transfer"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/transfer/stats"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/MamangRust/monolith-payment-gateway-transfer/handler"
	"github.com/MamangRust/monolith-payment-gateway-transfer/repository"
	"github.com/MamangRust/monolith-payment-gateway-transfer/service"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	card_repo "github.com/MamangRust/monolith-payment-gateway-card/repository"
	saldo_repo "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
)

type TransferStatsHandlerGapiTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	lis         *bufconn.Listener
	conn        *grpc.ClientConn
	client      pb.TransferStatsStatusServiceClient
	userID      int32
	cardNumber1 string
	cardNumber2 string
	testYear    int
	testMonth   int
}

func (s *TransferStatsHandlerGapiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts
	mongoDB, err := s.ts.MongoDB()
	if err != nil {
		s.Require().NoError(err)
	}

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

	s.lis = bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	pb.RegisterTransferStatsStatusServiceServer(server, h)

	go func() {
		if err := server.Serve(s.lis); err != nil {
		}
	}()

	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return s.lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn
	s.client = pb.NewTransferStatsStatusServiceClient(conn)

	s.testYear = time.Now().Year()
	s.testMonth = int(time.Now().Month())

	ctx := context.Background()
	_, err = mongoDB.Collection("users").InsertOne(ctx, bson.M{"user_id": 1, "firstname": "TransferGapi", "lastname": "Stats", "email": "transfer_gapi_stats@example.com", "password": "pass", "verification_code": "123", "is_verified": true})
	s.Require().NoError(err)

	s.cardNumber1 = "1111222233334444"
	s.cardNumber2 = "5555666677778888"
	_, err = mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_id": 1, "card_number": s.cardNumber1, "card_type": "debit", "cvv": "123", "card_provider": "visa", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)
	_, err = mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_id": 2, "card_number": s.cardNumber2, "card_type": "debit", "cvv": "123", "card_provider": "mastercard", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)

	_, err = mongoDB.Collection("transfers").InsertOne(ctx, bson.M{"transfer_id": 1, "transfer_no": uuid.New(), "transfer_from": s.cardNumber1, "transfer_to": s.cardNumber2, "transfer_amount": 150000, "transfer_time": time.Date(s.testYear, time.Month(s.testMonth), 10, 10, 0, 0, 0, time.UTC), "status": "success"})
	s.Require().NoError(err)
}

func (s *TransferStatsHandlerGapiTestSuite) TearDownSuite() {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.lis != nil {
		s.lis.Close()
	}
	s.ts.Teardown()
}

func (s *TransferStatsHandlerGapiTestSuite) TestFindMonthlyTransferStatusSuccess() {
	ctx := context.Background()
	res, err := s.client.FindMonthlyTransferStatusSuccess(ctx, &pbtransfer.FindMonthlyTransferStatus{
		Year:  int32(s.testYear),
		Month: int32(s.testMonth),
	})
	s.NoError(err)
	s.Equal("success", res.Status)
	s.NotEmpty(res.Data)
}

func TestTransferStatsHandlerGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransferStatsHandlerGapiTestSuite))
}
