package transfer_test

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/MamangRust/monolith-payment-gateway-transfer/repository"
	"github.com/MamangRust/monolith-payment-gateway-transfer/service"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	card_repo "github.com/MamangRust/monolith-payment-gateway-card/repository"
	saldo_repo "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
)

type TransferStatsServiceTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	svc         service.Service
	userID      int32
	cardNumber1 string
	cardNumber2 string
	testYear    int
	testMonth   int
}

func (s *TransferStatsServiceTestSuite) SetupSuite() {
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

	s.svc = service.NewService(&service.Deps{
		Kafka:        nil,
		Repositories: repos,
		Logger:       myLogger,
		Cache:        cacheStore,
		DB:           mongoDB,
	})

	s.testYear = time.Now().Year()
	s.testMonth = int(time.Now().Month())

	ctx := context.Background()
	_, err = mongoDB.Collection("users").InsertOne(ctx, bson.M{"user_id": 1, "firstname": "TransferService", "lastname": "Stats", "email": "transfer_svc_stats@example.com", "password": "pass", "verification_code": "123", "is_verified": true})
	s.Require().NoError(err)

	s.cardNumber1 = "1234567890123456"
	s.cardNumber2 = "6543210987654321"
	_, err = mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_id": 1, "card_number": s.cardNumber1, "card_type": "debit", "cvv": "123", "card_provider": "visa", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)
	_, err = mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_id": 2, "card_number": s.cardNumber2, "card_type": "debit", "cvv": "123", "card_provider": "mastercard", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)

	// Seed Transfers
	_, err = mongoDB.Collection("transfers").InsertOne(ctx, bson.M{"transfer_id": 1, "transfer_no": uuid.New(), "transfer_from": s.cardNumber1, "transfer_to": s.cardNumber2, "transfer_amount": 100000, "transfer_time": time.Date(s.testYear, time.Month(s.testMonth), 10, 10, 0, 0, 0, time.UTC), "status": "success"})
	s.Require().NoError(err)

	_, err = mongoDB.Collection("transfers").InsertOne(ctx, bson.M{"transfer_id": 2, "transfer_no": uuid.New(), "transfer_from": s.cardNumber1, "transfer_to": s.cardNumber2, "transfer_amount": 50000, "transfer_time": time.Date(s.testYear, time.Month(s.testMonth), 11, 10, 0, 0, 0, time.UTC), "status": "failed"})
	s.Require().NoError(err)
}

func (s *TransferStatsServiceTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *TransferStatsServiceTestSuite) TestTransferStatsService() {
	ctx := context.Background()

	// Global Monthly Success
	reqMonth := &requests.MonthStatusTransfer{Year: s.testYear, Month: s.testMonth}
	resSuccess, err := s.svc.FindMonthTransferStatusSuccess(ctx, reqMonth)
	s.NoError(err)
	s.NotEmpty(resSuccess)

	// Global Monthly Failed
	resFailed, err := s.svc.FindMonthTransferStatusFailed(ctx, reqMonth)
	s.NoError(err)
	s.NotEmpty(resFailed)

	// Yearly Success
	resYearSuccess, err := s.svc.FindYearlyTransferStatusSuccess(ctx, s.testYear)
	s.NoError(err)
	s.NotEmpty(resYearSuccess)

	// Card Specific Monthly Success
	reqCard := &requests.MonthStatusTransferCardNumber{
		Year:       s.testYear,
		Month:      s.testMonth,
		CardNumber: s.cardNumber1,
	}
	resCardSuccess, err := s.svc.FindMonthTransferStatusSuccessByCardNumber(ctx, reqCard)
	s.NoError(err)
	s.NotEmpty(resCardSuccess)

}

func TestTransferStatsServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransferStatsServiceTestSuite))
}
