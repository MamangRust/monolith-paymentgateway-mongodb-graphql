package withdraw_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	withdrawstatsrepository "github.com/MamangRust/monolith-payment-gateway-withdraw/repository/stats"
	withdrawstatsbycardrepository "github.com/MamangRust/monolith-payment-gateway-withdraw/repository/statsbycard"
	"github.com/stretchr/testify/suite"
)

type WithdrawStatsRepositoryTestSuite struct {
	suite.Suite
	ts               *tests.TestSuite
	repo             withdrawstatsrepository.WithdrawStatsStatusRepository
	repoAmount       withdrawstatsrepository.WithdrawStatsAmountRepository
	repoByCard       withdrawstatsbycardrepository.WithdrawStatsByCardStatusRepository
	repoAmountByCard withdrawstatsbycardrepository.WithdrawStatsByCardAmountRepository
	testYear         int
}

func (s *WithdrawStatsRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	mongoDB, err := s.ts.MongoDB()
	if err != nil {
		s.Require().NoError(err)
	}
	s.repo = withdrawstatsrepository.NewWithdrawStatsStatusRepository(mongoDB)
	s.repoAmount = withdrawstatsrepository.NewWithdrawStatsAmountRepository(mongoDB)
	s.repoByCard = withdrawstatsbycardrepository.NewWithdrawStatsByCardStatusRepository(mongoDB)
	s.repoAmountByCard = withdrawstatsbycardrepository.NewWithdrawStatsByCardAmountRepository(mongoDB)
	s.testYear = time.Now().Year()

	ctx := context.Background()

	_, err = mongoDB.Collection("users").InsertOne(ctx, bson.M{"firstname": "WithdrawRepo", "lastname": "Stats", "email": "withdraw_repo_stats@example.com", "password": "pass", "verification_code": "123", "is_verified": true})
	s.Require().NoError(err)

	cardNumber1 := "1111222233334444"
	_, err = mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_number": cardNumber1, "card_type": "debit", "cvv": "123", "card_provider": "visa", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)

	// Jan Success (1000)
	_, err = mongoDB.Collection("withdraws").InsertOne(ctx, bson.M{"withdraw_id": 1, "withdraw_no": uuid.New(), "card_number": cardNumber1, "withdraw_amount": 1000, "withdraw_time": time.Date(s.testYear, 1, 10, 10, 0, 0, 0, time.UTC), "status": "success"})
	s.Require().NoError(err)
	// Feb Failed (500)
	_, err = mongoDB.Collection("withdraws").InsertOne(ctx, bson.M{"withdraw_id": 2, "withdraw_no": uuid.New(), "card_number": cardNumber1, "withdraw_amount": 500, "withdraw_time": time.Date(s.testYear, 2, 15, 10, 0, 0, 0, time.UTC), "status": "failed"})
	s.Require().NoError(err)
}

func (s *WithdrawStatsRepositoryTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *WithdrawStatsRepositoryTestSuite) TestGlobalStats() {
	ctx := context.Background()

	reqMonth := &requests.MonthStatusWithdraw{Year: s.testYear, Month: 1}
	resS, err := s.repo.GetMonthWithdrawStatusSuccess(ctx, reqMonth)
	s.NoError(err)
	s.NotEmpty(resS)

	resF, err := s.repo.GetMonthWithdrawStatusFailed(ctx, reqMonth)
	s.NoError(err)
	s.NotEmpty(resF)

	resA, err := s.repoAmount.GetMonthlyWithdraws(ctx, s.testYear)
	s.NoError(err)
	s.NotEmpty(resA)

}

func (s *WithdrawStatsRepositoryTestSuite) TestByCardStats() {
	ctx := context.Background()
	cardNumber1 := "1111222233334444"

	reqMonth := &requests.MonthStatusWithdrawCardNumber{Year: s.testYear, Month: 1, CardNumber: cardNumber1}
	reqYearMonth := &requests.YearMonthCardNumber{Year: s.testYear, CardNumber: cardNumber1}

	resS, err := s.repoByCard.GetMonthWithdrawStatusSuccessByCardNumber(ctx, reqMonth)
	s.NoError(err)
	s.NotEmpty(resS)

	resF, err := s.repoByCard.GetMonthWithdrawStatusFailedByCardNumber(ctx, reqMonth)
	s.NoError(err)
	s.NotEmpty(resF)

	resA, err := s.repoAmountByCard.GetMonthlyWithdrawsByCardNumber(ctx, reqYearMonth)
	s.NoError(err)
	s.NotEmpty(resA)

	resY, err := s.repoAmountByCard.GetYearlyWithdrawsByCardNumber(ctx, reqYearMonth)
	s.NoError(err)
	s.NotEmpty(resY)

}

func TestWithdrawStatsRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(WithdrawStatsRepositoryTestSuite))
}
