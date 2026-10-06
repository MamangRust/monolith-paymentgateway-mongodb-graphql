package topup_test

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/MamangRust/monolith-payment-gateway-topup/repository"
	"github.com/MamangRust/monolith-payment-gateway-topup/service"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	card_repo "github.com/MamangRust/monolith-payment-gateway-card/repository"
	saldo_repo "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
)

type TopupStatsServiceTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	svc         service.Service
	userID      int32
	cardNumber1 string
	cardNumber2 string
	testYear    int
}

func (s *TopupStatsServiceTestSuite) SetupSuite() {
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
		cmd:   card_repo.NewCardCommandRepository(mongoDB.Collection("cards")),
	}

	repos := repository.NewRepositoriesWithAdapters(mongoDB, realCard, realSaldo)

	// Setup Logger
	zapLog := zap.NewNop()
	myLogger := &logger.Logger{Log: zapLog}

	// Setup Redis & Cache
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

	// Seed Data
	ctx := context.Background()
	_, err = mongoDB.Collection("users").InsertOne(ctx, bson.M{"user_id": 9210, "firstname": "TopupService", "lastname": "Stats", "email": "topup_svc_stats@example.com", "password": "pass", "verification_code": "123", "is_verified": true})
	s.Require().NoError(err)

	s.cardNumber1 = "2222333344445555"
	s.cardNumber2 = "6666777788889999"

	_, err = mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_id": 9210, "card_number": s.cardNumber1, "card_type": "debit", "cvv": "123", "card_provider": "visa", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)
	_, err = mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_id": 9211, "card_number": s.cardNumber2, "card_type": "debit", "cvv": "123", "card_provider": "mastercard", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)

	// Seed Topups
	_, err = mongoDB.Collection("topups").InsertOne(ctx, bson.M{"topup_id": 9311, "topup_no": "seed-topup-9311", "card_number": s.cardNumber1, "topup_amount": 1000, "topup_method": "bank_transfer", "topup_time": time.Date(s.testYear, 1, 10, 10, 0, 0, 0, time.UTC), "status": "success"})
	s.Require().NoError(err)
	_, err = mongoDB.Collection("topups").InsertOne(ctx, bson.M{"topup_id": 9312, "topup_no": "seed-topup-9312", "card_number": s.cardNumber2, "topup_amount": 2000, "topup_method": "bank_transfer", "topup_time": time.Date(s.testYear, 1, 15, 10, 0, 0, 0, time.UTC), "status": "success"})
	s.Require().NoError(err)
}

func (s *TopupStatsServiceTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *TopupStatsServiceTestSuite) TestTopupStatsService() {
	ctx := context.Background()
	// Global Monthly
	res, err := s.svc.FindMonthlyTopupAmounts(ctx, s.testYear)
	s.NoError(err)
	s.NotEmpty(res)

	// Card Specific Monthly
	reqMethod := &requests.YearMonthMethod{
		Year:       s.testYear,
		CardNumber: s.cardNumber1,
	}
	resCard, err := s.svc.FindMonthlyTopupAmountsByCardNumber(ctx, reqMethod)
	s.NoError(err)
	s.NotEmpty(resCard)

}

func TestTopupStatsServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TopupStatsServiceTestSuite))
}
