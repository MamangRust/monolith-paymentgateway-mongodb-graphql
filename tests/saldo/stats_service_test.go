package saldo_test

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	tests "github.com/MamangRust/monolith-payment-gateway-test"

	"github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	"github.com/MamangRust/monolith-payment-gateway-saldo/service"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type SaldoStatsServiceTestSuite struct {
	suite.Suite
	ts         *tests.TestSuite
	svc        service.Service
	userID     int32
	cardNumber string
	testYear   int
}

func (s *SaldoStatsServiceTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	mongoDB, err := s.ts.MongoDB()
	if err != nil {
		s.Require().NoError(err)
	}
	repos := repository.NewRepositories(mongoDB, nil, nil)

	zapLog := zap.NewNop()
	myLogger := &logger.Logger{Log: zapLog}

	redisOption, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	redisClient := redis.NewClient(redisOption)
	cacheStore := cache.NewCacheStore(redisClient, myLogger, &dummyCacheMetrics{})

	s.svc = service.NewService(&service.Deps{
		Repositories: repos,
		Logger:       myLogger,
		Cache:        cacheStore,
		DB:           mongoDB,
	})

	s.testYear = time.Now().Year()

	ctx := context.Background()
	_, err = mongoDB.Collection("users").InsertOne(ctx, bson.M{"firstname": "SaldoService", "lastname": "Stats", "email": "saldo_svc_stats@example.com", "password": "pass", "verification_code": "123", "is_verified": true})
	s.Require().NoError(err)

	s.cardNumber = "4444555566667777"
	_, err = mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_number": s.cardNumber, "card_type": "debit", "cvv": "123", "card_provider": "visa", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)

	// Seed Saldo with balance history
	_, err = mongoDB.Collection("saldos").InsertOne(ctx, bson.M{"card_number": s.cardNumber, "total_balance": 1000000})
	s.Require().NoError(err)
}

func (s *SaldoStatsServiceTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *SaldoStatsServiceTestSuite) TestSaldoStatsService() {
	ctx := context.Background()

	// Monthly balances
	resMonthly, err := s.svc.FindMonthlySaldoBalances(ctx, s.testYear)
	s.NoError(err)
	s.NotEmpty(resMonthly)

	// Yearly balances
	resYearly, err := s.svc.FindYearlySaldoBalances(ctx, s.testYear)
	s.NoError(err)
	s.NotEmpty(resYearly)
}

func TestSaldoStatsServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(SaldoStatsServiceTestSuite))
}
