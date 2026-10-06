package transaction_test

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	repository "github.com/MamangRust/monolith-payment-gateway-transaction/repository/stats"
	statsbycard_repository "github.com/MamangRust/monolith-payment-gateway-transaction/repository/statsbycard"
	"github.com/stretchr/testify/suite"
)

type TransactionStatsRepositoryTestSuite struct {
	suite.Suite
	ts         *tests.TestSuite
	repo       repository.TransactionStatsStatusRepository
	repoByCard statsbycard_repository.TransactionStatsByCardStatusRepository
	mongoDB    *mongo.Database
	testYear   int
	testMonth  int
}

func (s *TransactionStatsRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	mongoDB, err := s.ts.MongoDB()
	if err != nil {
		s.Require().NoError(err)
	}
	s.mongoDB = mongoDB
	s.repo = repository.NewTransactionStatsStatusRepository(mongoDB)
	s.repoByCard = statsbycard_repository.NewTransactionStatsByCardStatusRepository(mongoDB)
	s.testYear = time.Now().Year()
	s.testMonth = int(time.Now().Month())
}

func (s *TransactionStatsRepositoryTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *TransactionStatsRepositoryTestSuite) TestTransactionStatusStats() {
	ctx := context.Background()

	// Seed data
	_, err := s.mongoDB.Collection("users").InsertOne(ctx, bson.M{"firstname": "Transaction", "lastname": "Stats", "email": "transaction_stats@example.com", "password": "pass", "verification_code": "123", "is_verified": true})
	s.Require().NoError(err)

	cardNumber := "1111222233334444"
	_, err = s.mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_number": cardNumber, "card_type": "debit", "cvv": "123", "card_provider": "visa", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)

	merchantID := int32(1)
	_, err = s.mongoDB.Collection("merchants").InsertOne(ctx, bson.M{"merchant_id": merchantID, "merchant_no": uuid.New(), "name": "Merchant Stats", "api_key": "test_api_key", "status": "active"})
	s.Require().NoError(err)

	// Successful transaction
	_, err = s.mongoDB.Collection("transactions").InsertOne(ctx, bson.M{"transaction_id": 1, "transaction_no": uuid.New(), "card_number": cardNumber, "amount": 100000, "payment_method": "credit_card", "merchant_id": merchantID, "transaction_time": time.Date(s.testYear, time.Month(s.testMonth), 10, 10, 0, 0, 0, time.UTC), "status": "success"})
	s.Require().NoError(err)

	// Failed transaction
	_, err = s.mongoDB.Collection("transactions").InsertOne(ctx, bson.M{"transaction_id": 2, "transaction_no": uuid.New(), "card_number": cardNumber, "amount": 50000, "payment_method": "credit_card", "merchant_id": merchantID, "transaction_time": time.Date(s.testYear, time.Month(s.testMonth), 11, 10, 0, 0, 0, time.UTC), "status": "failed"})
	s.Require().NoError(err)

	// Global Monthly
	reqMonth := &requests.MonthStatusTransaction{Year: s.testYear, Month: s.testMonth}
	resSuccess, err := s.repo.GetMonthTransactionStatusSuccess(ctx, reqMonth)
	s.NoError(err)
	s.NotEmpty(resSuccess)

	resFailed, err := s.repo.GetMonthTransactionStatusFailed(ctx, reqMonth)
	s.NoError(err)
	s.NotEmpty(resFailed)

	// Global Yearly
	resYearSuccess, err := s.repo.GetYearlyTransactionStatusSuccess(ctx, s.testYear)
	s.NoError(err)
	s.NotEmpty(resYearSuccess)

	resYearFailed, err := s.repo.GetYearlyTransactionStatusFailed(ctx, s.testYear)
	s.NoError(err)
	s.NotEmpty(resYearFailed)

	// Card Monthly
	reqCard := &requests.MonthStatusTransactionCardNumber{
		Year:       s.testYear,
		Month:      s.testMonth,
		CardNumber: cardNumber,
	}
	resCardSuccess, err := s.repoByCard.GetMonthTransactionStatusSuccessByCardNumber(ctx, reqCard)
	s.NoError(err)
	s.NotEmpty(resCardSuccess)

	resCardFailed, err := s.repoByCard.GetMonthTransactionStatusFailedByCardNumber(ctx, reqCard)
	s.NoError(err)
	s.NotEmpty(resCardFailed)

}

func TestTransactionStatsRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransactionStatsRepositoryTestSuite))
}
