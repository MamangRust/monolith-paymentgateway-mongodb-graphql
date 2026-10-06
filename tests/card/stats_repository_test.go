package card_test

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"strconv"
	"testing"
	"time"

	repositorystats "github.com/MamangRust/monolith-payment-gateway-card/repository/stats"
	repositorystatsbycard "github.com/MamangRust/monolith-payment-gateway-card/repository/statsbycard"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/stretchr/testify/suite"
)

type CardStatsRepositoryTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	mongoDB     *mongo.Database
	repo        repositorystats.CardStatsRepository
	repoByCard  repositorystatsbycard.CardStatsByCardRepository
	cardNumber1 string
	cardNumber2 string
	testYear    int
}

func (s *CardStatsRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	mongoDB, err := s.ts.MongoDB()
	if err != nil {
		s.Require().NoError(err)
	}
	s.mongoDB = mongoDB
	s.repo = repositorystats.NewCardStatsRepository(mongoDB)
	s.repoByCard = repositorystatsbycard.NewCardStatsByCardRepository(mongoDB)
	s.testYear = time.Now().Year()

	// Seed Users and Cards
	ctx := context.Background()
	s.cardNumber1 = "1111222233334444"
	s.cardNumber2 = "5555666677778888"

	// Create a user first
	_, err = s.mongoDB.Collection("users").InsertOne(ctx, bson.M{"firstname": "Stats", "lastname": "User", "email": "stats@example.com", "password": "pass", "verification_code": "123456", "is_verified": true})
	s.Require().NoError(err)

	// Create Cards. The cards collection indexes card_id as unique, so the
	// fixtures must carry one.
	s.Require().NoError(s.ts.InsertFixture(ctx, "cards", "card_id", "", bson.M{"card_number": s.cardNumber1, "card_type": "debit", "cvv": "123", "card_provider": "visa", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)}))
	s.Require().NoError(s.ts.InsertFixture(ctx, "cards", "card_id", "", bson.M{"card_number": s.cardNumber2, "card_type": "credit", "cvv": "456", "card_provider": "mastercard", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)}))

	s.seedHistoricalData()
}

func (s *CardStatsRepositoryTestSuite) seedHistoricalData() {
	// Seed Saldos (Monthly Balance)
	// Jan: 1000 (Card 1)
	// Feb: 2000 (Card 1), 3000 (Card 2)
	s.insertSaldo(s.cardNumber1, 1000, time.Date(s.testYear, 1, 15, 10, 0, 0, 0, time.UTC))
	s.insertSaldo(s.cardNumber1, 2000, time.Date(s.testYear, 2, 15, 10, 0, 0, 0, time.UTC))
	s.insertSaldo(s.cardNumber2, 3000, time.Date(s.testYear, 2, 20, 10, 0, 0, 0, time.UTC))

	// Seed Topups
	// Jan: 500 (Card 1)
	// Feb: 1000 (Card 1), 1500 (Card 2)
	s.insertTopup(s.cardNumber1, 500, time.Date(s.testYear, 1, 10, 10, 0, 0, 0, time.UTC))
	s.insertTopup(s.cardNumber1, 1000, time.Date(s.testYear, 2, 10, 10, 0, 0, 0, time.UTC))
	s.insertTopup(s.cardNumber2, 1500, time.Date(s.testYear, 2, 12, 10, 0, 0, 0, time.UTC))

	// Seed Withdraws
	// Jan: 100 (Card 1)
	// Feb: 200 (Card 1), 300 (Card 2)
	s.insertWithdraw(s.cardNumber1, 100, time.Date(s.testYear, 1, 20, 10, 0, 0, 0, time.UTC))
	s.insertWithdraw(s.cardNumber1, 200, time.Date(s.testYear, 2, 20, 10, 0, 0, 0, time.UTC))
	s.insertWithdraw(s.cardNumber2, 300, time.Date(s.testYear, 2, 25, 10, 0, 0, 0, time.UTC))

	// Seed Transfers
	// Jan: 50 from Card 1 to Card 2
	// Feb: 100 from Card 1 to Card 2
	s.insertTransfer(s.cardNumber1, s.cardNumber2, 50, time.Date(s.testYear, 1, 25, 10, 0, 0, 0, time.UTC))
	s.insertTransfer(s.cardNumber1, s.cardNumber2, 100, time.Date(s.testYear, 2, 28, 10, 0, 0, 0, time.UTC))
}

func (s *CardStatsRepositoryTestSuite) insertSaldo(cardNumber string, amount int, t time.Time) {
	_, err := s.mongoDB.Collection("saldos").UpdateMany(context.Background(), bson.M{"card_number": cardNumber}, bson.M{"$set": bson.M{"deleted_at": time.Now()}})
	s.Require().NoError(err)
	s.Require().NoError(s.ts.InsertFixture(context.Background(), "saldos", "saldo_id", "", bson.M{"card_number": cardNumber, "total_balance": amount, "created_at": t, "updated_at": t}))
}

func (s *CardStatsRepositoryTestSuite) insertTopup(cardNumber string, amount int, t time.Time) {
	s.Require().NoError(s.ts.InsertFixture(context.Background(), "topups", "topup_id", "topup_no", bson.M{"card_number": cardNumber, "topup_amount": amount, "topup_time": t, "topup_method": "bank_transfer", "status": "success"}))
}

func (s *CardStatsRepositoryTestSuite) insertWithdraw(cardNumber string, amount int, t time.Time) {
	s.Require().NoError(s.ts.InsertFixture(context.Background(), "withdraws", "withdraw_id", "withdraw_no", bson.M{"card_number": cardNumber, "withdraw_amount": amount, "withdraw_time": t, "status": "success"}))
}

func (s *CardStatsRepositoryTestSuite) insertTransfer(from, to string, amount int, t time.Time) {
	s.Require().NoError(s.ts.InsertFixture(context.Background(), "transfers", "transfer_id", "transfer_no", bson.M{"transfer_from": from, "transfer_to": to, "transfer_amount": amount, "transfer_time": t, "status": "success"}))
}

func (s *CardStatsRepositoryTestSuite) TearDownSuite() {
	if s.ts != nil {
		s.ts.Teardown()
	}
}

// --- Balance Tests ---

func (s *CardStatsRepositoryTestSuite) TestBalanceStats() {
	ctx := context.Background()

	// Global Monthly
	res, err := s.repo.GetMonthlyBalance(ctx, s.testYear)
	s.NoError(err)
	s.NotEmpty(res)
	s.Equal(int32(1000), res[0].TotalBalance) // Jan
	s.Equal(int32(5000), res[1].TotalBalance) // Feb (2000 + 3000)

	// By Card Monthly
	req1 := &requests.MonthYearCardNumberCard{CardNumber: s.cardNumber1, Year: s.testYear}
	res1, err := s.repoByCard.GetMonthlyBalancesByCardNumber(ctx, req1)
	s.NoError(err)
	s.Equal(int32(1000), res1[0].TotalBalance) // Jan Card 1
	s.Equal(int32(2000), res1[1].TotalBalance) // Feb Card 1

	// Yearly
	yRes, err := s.repo.GetYearlyBalance(ctx, s.testYear)
	s.NoError(err)
	s.NotEmpty(yRes)
	found := false
	yearStr := strconv.Itoa(s.testYear)
	for _, r := range yRes {
		// Normalize the numeric value for comparison
		val := fmt.Sprintf("%v", r.Year)
		if fmt.Sprintf("%v", val) == yearStr {
			s.Equal(int32(6000), int32(r.TotalBalance))
			found = true
		}
	}
	s.True(found)
}

// --- Topup Tests ---

func (s *CardStatsRepositoryTestSuite) TestTopupStats() {
	ctx := context.Background()

	// Global Monthly
	res, err := s.repo.GetMonthlyTopupAmount(ctx, s.testYear)
	s.NoError(err)
	s.Equal(int32(500), res[0].TotalTopupAmount)  // Jan
	s.Equal(int32(2500), res[1].TotalTopupAmount) // Feb (1000 + 1500)

	// By Card Monthly
	req1 := &requests.MonthYearCardNumberCard{CardNumber: s.cardNumber1, Year: s.testYear}
	res1, err := s.repoByCard.GetMonthlyTopupAmountByCardNumber(ctx, req1)
	s.NoError(err)
	s.Equal(int32(500), res1[0].TotalTopupAmount)  // Jan Card 1
	s.Equal(int32(1000), res1[1].TotalTopupAmount) // Feb Card 1
}

// --- Withdraw Tests ---

func (s *CardStatsRepositoryTestSuite) TestWithdrawStats() {
	ctx := context.Background()

	// Global Monthly
	res, err := s.repo.GetMonthlyWithdrawAmount(ctx, s.testYear)
	s.NoError(err)
	s.Equal(int32(100), res[0].TotalWithdrawAmount) // Jan
	s.Equal(int32(500), res[1].TotalWithdrawAmount) // Feb (200 + 300)

	// By Card Monthly
	req1 := &requests.MonthYearCardNumberCard{CardNumber: s.cardNumber1, Year: s.testYear}
	res1, err := s.repoByCard.GetMonthlyWithdrawAmountByCardNumber(ctx, req1)
	s.NoError(err)
	s.Equal(int32(100), res1[0].TotalWithdrawAmount) // Jan Card 1
	s.Equal(int32(200), res1[1].TotalWithdrawAmount) // Feb Card 1
}

// --- Transfer Tests ---

func (s *CardStatsRepositoryTestSuite) TestTransferStats() {
	ctx := context.Background()

	// Global Monthly Sender
	resS, err := s.repo.GetMonthlyTransferAmountSender(ctx, s.testYear)
	s.NoError(err)
	s.Equal(int32(50), resS[0].TotalSentAmount)  // Jan
	s.Equal(int32(100), resS[1].TotalSentAmount) // Feb

	// By Card Monthly Sender
	req1 := &requests.MonthYearCardNumberCard{CardNumber: s.cardNumber1, Year: s.testYear}
	res1S, err := s.repoByCard.GetMonthlyTransferAmountBySender(ctx, req1)
	s.NoError(err)
	s.Equal(int32(50), res1S[0].TotalSentAmount)  // Jan Card 1
	s.Equal(int32(100), res1S[1].TotalSentAmount) // Feb Card 1

	// By Card Monthly Receiver (Card 2)
	req2 := &requests.MonthYearCardNumberCard{CardNumber: s.cardNumber2, Year: s.testYear}
	res2R, err := s.repoByCard.GetMonthlyTransferAmountByReceiver(ctx, req2)
	s.NoError(err)
	s.Equal(int32(50), res2R[0].TotalReceivedAmount)  // Jan Card 2
	s.Equal(int32(100), res2R[1].TotalReceivedAmount) // Feb Card 2
}

func TestCardStatsRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(CardStatsRepositoryTestSuite))
}
