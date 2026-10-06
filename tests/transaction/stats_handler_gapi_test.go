package transaction_test

import (
	"context"
	"net"
	"testing"
	"time"

	pbtransaction "github.com/MamangRust/monolith-payment-gateway-pb/transaction"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/transaction/stats"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/MamangRust/monolith-payment-gateway-transaction/handler"
	"github.com/MamangRust/monolith-payment-gateway-transaction/repository"
	"github.com/MamangRust/monolith-payment-gateway-transaction/service"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	card_repo "github.com/MamangRust/monolith-payment-gateway-card/repository"
	merchant_repo "github.com/MamangRust/monolith-payment-gateway-merchant/repository"
	saldo_repo "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type TransactionStatsHandlerGapiTestSuite struct {
	suite.Suite
	ts         *tests.TestSuite
	lis        *bufconn.Listener
	conn       *grpc.ClientConn
	client     pb.TransactionStatsStatusServiceClient
	userID     int32
	cardNumber string
	merchantID int32
	testYear   int
	testMonth  int
}

func (s *TransactionStatsHandlerGapiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	mongoDB, err := s.ts.MongoDB()
	if err != nil {
		s.Require().NoError(err)
	}

	realSaldo := &realSaldoRepo{repo: saldo_repo.NewRepositories(mongoDB, nil, nil)}
	realCard := &realCardRepo{
		query:   card_repo.NewCardQueryRepository(mongoDB.Collection("cards")),
		command: card_repo.NewCardCommandRepository(mongoDB.Collection("cards")),
	}
	realMerchant := &realMerchantRepo{repo: merchant_repo.NewMerchantQueryRepository(mongoDB.Collection("merchants"))}

	repos := repository.NewRepositoriesWithAdapters(mongoDB, realSaldo, realCard, realMerchant)

	zapLog := zap.NewNop()
	myLogger := &logger.Logger{Log: zapLog}

	redisOption, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	redisClient := redis.NewClient(redisOption)
	cacheStore := cache.NewCacheStore(redisClient, myLogger, &dummyCacheMetrics{})

	harness, hErr := tests.NewDependencyHarness(mongoDB, cacheStore, myLogger)
	s.Require().NoError(hErr)
	s.T().Cleanup(harness.Close)
	svc := service.NewService(&service.Deps{
		CardAdapter:     harness.CardAdapter,
		SaldoAdapter:    harness.SaldoAdapter,
		MerchantAdapter: harness.MerchantAdapter,
		DB:              mongoDB,
		Kafka:           nil,
		Repositories:    repos,
		Logger:          myLogger,
		Cache:           cacheStore,
	})

	h := handler.NewHandler(svc)

	s.lis = bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	pb.RegisterTransactionStatsStatusServiceServer(server, h)

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
	s.client = pb.NewTransactionStatsStatusServiceClient(conn)

	s.testYear = time.Now().Year()
	s.testMonth = int(time.Now().Month())

	ctx := context.Background()
	_, err = mongoDB.Collection("users").InsertOne(ctx, bson.M{"firstname": "TransactionGapi", "lastname": "Stats", "email": "transaction_gapi_stats@example.com", "password": "pass", "verification_code": "123", "is_verified": true})
	s.Require().NoError(err)

	s.cardNumber = "2222333344445555"
	_, err = mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_number": s.cardNumber, "card_type": "debit", "cvv": "123", "card_provider": "visa", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)

	_, _ = mongoDB.Collection("merchants").InsertOne(ctx, bson.M{"merchant_id": 1, "merchant_no": uuid.New(), "name": "Merchant Gapi", "api_key": "test_key_gapi", "status": "active"})
	s.merchantID = 1
	s.Require().NoError(err)

	_, err = mongoDB.Collection("transactions").InsertOne(ctx, bson.M{"card_number": s.cardNumber, "merchant_id": s.merchantID, "amount": 200000, "payment_method": "credit_card", "transaction_time": time.Date(s.testYear, time.Month(s.testMonth), 10, 10, 0, 0, 0, time.UTC), "status": "success"})
	s.Require().NoError(err)
}

func (s *TransactionStatsHandlerGapiTestSuite) TearDownSuite() {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.lis != nil {
		s.lis.Close()
	}
	s.ts.Teardown()
}

func (s *TransactionStatsHandlerGapiTestSuite) TestFindMonthlyTransactionStatusSuccess() {
	ctx := context.Background()
	res, err := s.client.FindMonthlyTransactionStatusSuccess(ctx, &pbtransaction.FindMonthlyTransactionStatus{
		Year:  int32(s.testYear),
		Month: int32(s.testMonth),
	})
	s.NoError(err)
	s.Equal("success", res.Status)
	s.NotEmpty(res.Data)
}

func TestTransactionStatsHandlerGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransactionStatsHandlerGapiTestSuite))
}
