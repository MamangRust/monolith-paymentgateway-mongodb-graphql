package transaction_test

import (
	"context"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	tests "github.com/MamangRust/monolith-payment-gateway-test"

	"github.com/MamangRust/monolith-payment-gateway-transaction/repository"
	"github.com/MamangRust/monolith-payment-gateway-transaction/service"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	card_repo "github.com/MamangRust/monolith-payment-gateway-card/repository"
	merchant_repo "github.com/MamangRust/monolith-payment-gateway-merchant/repository"
	saldo_repo "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
)

type TransactionStatsServiceTestSuite struct {
	suite.Suite
	ts         *tests.TestSuite
	svc        service.Service
	userID     int32
	cardNumber string
	merchantID int32
	testYear   int
	testMonth  int
}

func (s *TransactionStatsServiceTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	mongoDB, err := s.ts.MongoDB()
	if err != nil {
		s.Require().NoError(err)
	}

	// Real repository implementations for full integration
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
	s.svc = service.NewService(&service.Deps{
		CardAdapter:     harness.CardAdapter,
		SaldoAdapter:    harness.SaldoAdapter,
		MerchantAdapter: harness.MerchantAdapter,
		DB:              mongoDB,
		Kafka:           nil,
		Repositories:    repos,
		Logger:          myLogger,
		Cache:           cacheStore,
	})

	s.testYear = time.Now().Year()
	s.testMonth = int(time.Now().Month())

	ctx := context.Background()
	_, err = mongoDB.Collection("users").InsertOne(ctx, bson.M{"user_id": 1, "firstname": "TransactionService", "lastname": "Stats", "email": "transaction_svc_stats@example.com", "password": "pass", "verification_code": "123", "is_verified": true})
	s.Require().NoError(err)
	s.userID = 1

	s.cardNumber = "1111222233334444"
	_, err = mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_number": s.cardNumber, "card_type": "debit", "cvv": "123", "card_provider": "visa", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)

	// Merchant needs to be active for transactions
	_, err = mongoDB.Collection("merchants").InsertOne(ctx, bson.M{"merchant_id": 1, "merchant_no": uuid.New(), "name": "Test Merchant", "api_key": "test_key", "status": "active"})
	s.Require().NoError(err)
	s.merchantID = 1

	// Seed Transactions
	_, err = mongoDB.Collection("transactions").InsertOne(ctx, bson.M{"transaction_id": 1, "transaction_no": uuid.New(), "card_number": s.cardNumber, "merchant_id": s.merchantID, "amount": 100000, "payment_method": "bank_transfer", "transaction_time": time.Date(s.testYear, time.Month(s.testMonth), 10, 10, 0, 0, 0, time.UTC), "status": "success"})
	s.Require().NoError(err)

	_, err = mongoDB.Collection("transactions").InsertOne(ctx, bson.M{"transaction_id": 2, "transaction_no": uuid.New(), "card_number": s.cardNumber, "merchant_id": s.merchantID, "amount": 50000, "payment_method": "bank_transfer", "transaction_time": time.Date(s.testYear, time.Month(s.testMonth), 11, 10, 0, 0, 0, time.UTC), "status": "failed"})
	s.Require().NoError(err)
}

func (s *TransactionStatsServiceTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *TransactionStatsServiceTestSuite) TestTransactionStatsService() {
	ctx := context.Background()

	// Global Monthly Success
	reqMonth := &requests.MonthStatusTransaction{Year: s.testYear, Month: s.testMonth}
	resSuccess, err := s.svc.FindMonthTransactionStatusSuccess(ctx, reqMonth)
	s.NoError(err)
	s.NotEmpty(resSuccess)

	// Global Monthly Failed
	resFailed, err := s.svc.FindMonthTransactionStatusFailed(ctx, reqMonth)
	s.NoError(err)
	s.NotEmpty(resFailed)

	// Yearly Success
	resYearSuccess, err := s.svc.FindYearlyTransactionStatusSuccess(ctx, s.testYear)
	s.NoError(err)
	s.NotEmpty(resYearSuccess)

}

func TestTransactionStatsServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransactionStatsServiceTestSuite))
}
