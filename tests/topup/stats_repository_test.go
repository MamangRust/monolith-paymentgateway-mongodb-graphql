package topup_test

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	stats_repo "github.com/MamangRust/monolith-payment-gateway-topup/repository/stats"
	statsbycard_repo "github.com/MamangRust/monolith-payment-gateway-topup/repository/statsbycard"
	"github.com/stretchr/testify/suite"
)

type TopupStatsRepositoryTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	repo        stats_repo.TopupStatsRepository
	statsbycard statsbycard_repo.TopupStatsByCardRepository
	userID      int32
	cardNumber1 string
	cardNumber2 string
	testYear    int
}

func (s *TopupStatsRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	mongoDB, err := s.ts.MongoDB()
	if err != nil {
		s.Require().NoError(err)
	}
	s.repo = stats_repo.NewTopupStatsRepository(mongoDB)
	s.statsbycard = statsbycard_repo.NewTopupStatsByCardRepository(mongoDB)

	s.testYear = time.Now().Year()

	// Seed Data
	ctx := context.Background()
	_, err = mongoDB.Collection("users").InsertOne(ctx, bson.M{"user_id": 9200, "firstname": "Topup", "lastname": "Stats", "email": "topup_stats@example.com", "password": "pass", "verification_code": "123", "is_verified": true})
	s.Require().NoError(err)

	s.cardNumber1 = "1111222233334444"
	s.cardNumber2 = "5555666677778888"

	_, err = mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_id": 9200, "card_number": s.cardNumber1, "card_type": "debit", "cvv": "123", "card_provider": "visa", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)
	_, err = mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_id": 9201, "card_number": s.cardNumber2, "card_type": "debit", "cvv": "123", "card_provider": "mastercard", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)

	// Seed Topups for trailing 12 months and 5 years
	now := time.Now().UTC()

	// Topups in current month
	_, err = mongoDB.Collection("topups").InsertOne(ctx, bson.M{"topup_id": 9301, "topup_no": "seed-topup-9301", "card_number": s.cardNumber1, "topup_amount": 500, "topup_method": "bank_transfer", "topup_time": now.AddDate(0, 0, -1), "status": "success"})
	s.Require().NoError(err)

	// Topups last month
	_, err = mongoDB.Collection("topups").InsertOne(ctx, bson.M{"topup_id": 9302, "topup_no": "seed-topup-9302", "card_number": s.cardNumber1, "topup_amount": 300, "topup_method": "e-wallet", "topup_time": now.AddDate(0, -1, -5), "status": "success"})
	s.Require().NoError(err)

	// Topups 13 months ago (should not show in trailing 12)
	_, err = mongoDB.Collection("topups").InsertOne(ctx, bson.M{"topup_id": 9303, "topup_no": "seed-topup-9303", "card_number": s.cardNumber1, "topup_amount": 1000, "topup_method": "bank_transfer", "topup_time": now.AddDate(0, -13, 0), "status": "success"})
	s.Require().NoError(err)

	// Topups for yearly stats (5 years rolling)
	_, err = mongoDB.Collection("topups").InsertOne(ctx, bson.M{"topup_id": 9304, "topup_no": "seed-topup-9304", "card_number": s.cardNumber2, "topup_amount": 1000, "topup_method": "bank_transfer", "topup_time": now, "status": "success"})
	s.Require().NoError(err)
	_, err = mongoDB.Collection("topups").InsertOne(ctx, bson.M{"topup_id": 9305, "topup_no": "seed-topup-9305", "card_number": s.cardNumber2, "topup_amount": 2000, "topup_method": "bank_transfer", "topup_time": now.AddDate(-1, 0, 0), "status": "success"})
	s.Require().NoError(err)
	_, err = mongoDB.Collection("topups").InsertOne(ctx, bson.M{"topup_id": 9306, "topup_no": "seed-topup-9306", "card_number": s.cardNumber2, "topup_amount": 3000, "topup_method": "bank_transfer", "topup_time": now.AddDate(-6, 0, 0), "status": "success"})
	s.Require().NoError(err)
}

func (s *TopupStatsRepositoryTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *TopupStatsRepositoryTestSuite) TestGlobalStats() {
	ctx := context.Background()

	// Monthly Amount
	res, err := s.repo.GetMonthlyTopupAmounts(ctx, s.testYear)
	s.NoError(err)
	s.NotEmpty(res)

	// Yearly Amount
	yearlyRes, err := s.repo.GetYearlyTopupAmounts(ctx, s.testYear)
	s.NoError(err)
	s.NotEmpty(yearlyRes)

	// Monthly Methods (Trailing 12 months)
	methRes, err := s.repo.GetMonthlyTopupMethods(ctx, s.testYear)
	s.NoError(err)
	s.NotEmpty(methRes)

	// Status Stats
	// Global Monthly
	reqMonth := &requests.MonthTopupStatus{Year: s.testYear, Month: 1}
	resS, err := s.repo.GetMonthTopupStatusSuccess(ctx, reqMonth)
	s.NoError(err)
	s.NotEmpty(resS)

	resF, err := s.repo.GetMonthTopupStatusFailed(ctx, reqMonth)
	s.NoError(err)
	s.NotEmpty(resF)

}

func (s *TopupStatsRepositoryTestSuite) TestCardStats() {
	ctx := context.Background()

	reqMethod := &requests.YearMonthMethod{
		CardNumber: s.cardNumber1,
		Year:       s.testYear,
	}

	// Card 1 Monthly Amounts
	res1, err := s.statsbycard.GetMonthlyTopupAmountsByCardNumber(ctx, reqMethod)
	s.NoError(err)
	s.NotEmpty(res1)

	// Card 1 Monthly Methods
	methRes1, err := s.statsbycard.GetMonthlyTopupMethodsByCardNumber(ctx, reqMethod)
	s.NoError(err)
	s.NotEmpty(methRes1)

	// Card 1 Status success
	now := time.Now().UTC()
	reqStatus := &requests.MonthTopupStatusCardNumber{
		CardNumber: s.cardNumber1,
		Year:       s.testYear,
		Month:      int(now.Month()),
	}
	statusRes1, err := s.statsbycard.GetMonthTopupStatusSuccessByCardNumber(ctx, reqStatus)
	s.NoError(err)
	s.NotEmpty(statusRes1)

}

func TestTopupStatsRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TopupStatsRepositoryTestSuite))
}
