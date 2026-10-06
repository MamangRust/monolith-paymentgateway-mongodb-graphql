package transfer_test

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	repository "github.com/MamangRust/monolith-payment-gateway-transfer/repository/stats"
	statsbycard_repository "github.com/MamangRust/monolith-payment-gateway-transfer/repository/statsbycard"
	"github.com/stretchr/testify/suite"
)

type TransferStatsRepositoryTestSuite struct {
	suite.Suite
	ts         *tests.TestSuite
	repo       repository.TransferStatsStatusRepository
	repoByCard statsbycard_repository.TransferStatsByCardStatusRepository
	mongoDB    *mongo.Database
	testYear   int
	testMonth  int
}

func (s *TransferStatsRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	mongoDB, err := s.ts.MongoDB()
	if err != nil {
		s.Require().NoError(err)
	}
	s.mongoDB = mongoDB
	s.repo = repository.NewTransferStatsStatusRepository(mongoDB)
	s.repoByCard = statsbycard_repository.NewTransferStatsByCardStatusRepository(mongoDB)
	s.testYear = time.Now().Year()
	s.testMonth = int(time.Now().Month())
}

func (s *TransferStatsRepositoryTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *TransferStatsRepositoryTestSuite) TestTransferStatusStats() {
	ctx := context.Background()

	// Seed data
	_, err := s.mongoDB.Collection("users").InsertOne(ctx, bson.M{"user_id": 1, "firstname": "Transfer", "lastname": "Stats", "email": "transfer_stats@example.com", "password": "pass", "verification_code": "123", "is_verified": true})
	s.Require().NoError(err)

	cardFrom := "1111222233334444"
	cardTo := "5555666677778888"
	_, err = s.mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_id": 1, "card_number": cardFrom, "card_type": "debit", "cvv": "123", "card_provider": "visa", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)
	_, err = s.mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_id": 2, "card_number": cardTo, "card_type": "debit", "cvv": "123", "card_provider": "visa", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)

	// Successful transfer
	_, err = s.mongoDB.Collection("transfers").InsertOne(ctx, bson.M{"transfer_id": 1, "transfer_no": uuid.New(), "transfer_from": cardFrom, "transfer_to": cardTo, "transfer_amount": 100000, "transfer_time": time.Date(s.testYear, time.Month(s.testMonth), 10, 10, 0, 0, 0, time.UTC), "status": "success"})
	s.Require().NoError(err)

	// Failed transfer
	_, err = s.mongoDB.Collection("transfers").InsertOne(ctx, bson.M{"transfer_id": 2, "transfer_no": uuid.New(), "transfer_from": cardFrom, "transfer_to": cardTo, "transfer_amount": 50000, "transfer_time": time.Date(s.testYear, time.Month(s.testMonth), 11, 10, 0, 0, 0, time.UTC), "status": "failed"})
	s.Require().NoError(err)

	// Global Monthly
	reqMonth := &requests.MonthStatusTransfer{Year: s.testYear, Month: s.testMonth}
	resSuccess, err := s.repo.GetMonthTransferStatusSuccess(ctx, reqMonth)
	s.NoError(err)
	s.NotEmpty(resSuccess)

	resFailed, err := s.repo.GetMonthTransferStatusFailed(ctx, reqMonth)
	s.NoError(err)
	s.NotEmpty(resFailed)

	// Global Yearly
	resYearSuccess, err := s.repo.GetYearlyTransferStatusSuccess(ctx, s.testYear)
	s.NoError(err)
	s.NotEmpty(resYearSuccess)

	resYearFailed, err := s.repo.GetYearlyTransferStatusFailed(ctx, s.testYear)
	s.NoError(err)
	s.NotEmpty(resYearFailed)

	// Card Monthly
	reqCard := &requests.MonthStatusTransferCardNumber{
		Year:       s.testYear,
		Month:      s.testMonth,
		CardNumber: cardFrom,
	}
	resCardSuccess, err := s.repoByCard.GetMonthTransferStatusSuccessByCardNumber(ctx, reqCard)
	s.NoError(err)
	s.NotEmpty(resCardSuccess)

	resCardFailed, err := s.repoByCard.GetMonthTransferStatusFailedByCardNumber(ctx, reqCard)
	s.NoError(err)
	s.NotEmpty(resCardFailed)
}

func TestTransferStatsRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransferStatsRepositoryTestSuite))
}
