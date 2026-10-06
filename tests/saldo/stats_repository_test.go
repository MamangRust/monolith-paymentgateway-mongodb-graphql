package saldo_test

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
	"time"

	repository "github.com/MamangRust/monolith-payment-gateway-saldo/repository/stats"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/stretchr/testify/suite"
)

type SaldoStatsRepositoryTestSuite struct {
	suite.Suite
	ts       *tests.TestSuite
	mongoDB  *mongo.Database
	repo     repository.SaldoStatsRepository
	testYear int
}

func (s *SaldoStatsRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	mongoDB, err := s.ts.MongoDB()
	if err != nil {
		s.Require().NoError(err)
	}
	s.mongoDB = mongoDB
	s.repo = repository.NewSaldoStatsRepository(mongoDB)
	s.testYear = time.Now().Year()
}

func (s *SaldoStatsRepositoryTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *SaldoStatsRepositoryTestSuite) TestBalanceStats() {
	ctx := context.Background()

	// Seed data
	_, err := s.mongoDB.Collection("users").InsertOne(ctx, bson.M{"user_id": 9101, "firstname": "Saldo", "lastname": "Stats", "email": "saldo_stats_balance@example.com", "password": "pass", "verification_code": "123", "is_verified": true})
	s.Require().NoError(err)

	cardNumber := "1111222233334444"
	_, err = s.mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_id": 9101, "card_number": cardNumber, "card_type": "debit", "cvv": "123", "card_provider": "visa", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)

	_, err = s.mongoDB.Collection("saldos").InsertOne(ctx, bson.M{"saldo_id": 9101, "card_number": cardNumber, "total_balance": 50000, "created_at": time.Date(s.testYear, 1, 15, 10, 0, 0, 0, time.UTC), "updated_at": time.Date(s.testYear, 1, 15, 10, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)

	// Monthly Global
	res, err := s.repo.GetMonthlySaldoBalances(ctx, s.testYear)
	s.NoError(err)
	s.NotEmpty(res)

	// Yearly Global (5 years)
	yearlyRes, err := s.repo.GetYearlySaldoBalances(ctx, s.testYear)
	s.NoError(err)
	s.Len(yearlyRes, 5)

}

func (s *SaldoStatsRepositoryTestSuite) TestTotalStats() {
	ctx := context.Background()

	// Seed data for current month
	_, err := s.mongoDB.Collection("users").InsertOne(ctx, bson.M{"user_id": 9102, "firstname": "Saldo", "lastname": "Total", "email": "saldo_total_stats@example.com", "password": "pass", "verification_code": "123", "is_verified": true})
	s.Require().NoError(err)
	cardNumber := "2222333344445555"
	_, err = s.mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_id": 9102, "card_number": cardNumber, "card_type": "debit", "cvv": "123", "card_provider": "visa", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)
	now := time.Now()
	_, err = s.mongoDB.Collection("saldos").InsertOne(ctx, bson.M{"saldo_id": 9102, "card_number": cardNumber, "total_balance": 50000, "created_at": now, "updated_at": now})
	s.Require().NoError(err)

	// Monthly Total Balance (at least current month)
	req := &requests.MonthTotalSaldoBalance{
		Year:  s.testYear,
		Month: int(time.Now().Month()),
	}
	res, err := s.repo.GetMonthlyTotalSaldoBalance(ctx, req)
	s.NoError(err)
	s.NotEmpty(res)

	// Yearly Total Balance
	yearlyRes, err := s.repo.GetYearTotalSaldoBalance(ctx, s.testYear)
	s.NoError(err)
	s.NotEmpty(yearlyRes)
}

func TestSaldoStatsRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(SaldoStatsRepositoryTestSuite))
}
