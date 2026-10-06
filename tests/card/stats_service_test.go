package card_test

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-card/repository"
	"github.com/MamangRust/monolith-payment-gateway-card/service"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

type CardStatsServiceTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	mongoDB     *mongo.Database
	cardService service.Service
	cardNumber1 string
	cardNumber2 string
	testYear    int
}

func (s *CardStatsServiceTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	redisClient := redis.NewClient(opts)

	mongoDB, err := s.ts.MongoDB()
	if err != nil {
		s.Require().NoError(err)
	}
	s.mongoDB = mongoDB
	repos := repository.NewRepositories(mongoDB, nil)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(redisClient, log, cacheMetrics)

	s.cardService = service.NewService(&service.Deps{
		Repositories: repos,
		Logger:       log,
		Cache:        cacheStore,
		Kafka:        nil,
	})

	s.testYear = time.Now().Year()

	// Seed Users and Cards using direct SQL (same as repository test)
	ctx := context.Background()
	s.cardNumber1 = "2222333344445555"
	s.cardNumber2 = "6666777788889999"

	// Create a user first
	_, err = s.mongoDB.Collection("users").InsertOne(ctx, bson.M{"firstname": "Service", "lastname": "Stats", "email": "service_stats@example.com", "password": "pass", "verification_code": "123456", "is_verified": true})
	s.Require().NoError(err)

	// Create Cards
	s.Require().NoError(s.ts.InsertFixture(ctx, "cards", "card_id", "", bson.M{"card_number": s.cardNumber1, "card_type": "debit", "cvv": "123", "card_provider": "visa", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)}))
	s.Require().NoError(s.ts.InsertFixture(ctx, "cards", "card_id", "", bson.M{"card_number": s.cardNumber2, "card_type": "credit", "cvv": "456", "card_provider": "mastercard", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)}))

	s.seedHistoricalData()
}

func (s *CardStatsServiceTestSuite) seedHistoricalData() {
	// Seed Saldos
	s.insertSaldo(s.cardNumber1, 1000, time.Date(s.testYear, 1, 15, 10, 0, 0, 0, time.UTC))
	s.insertSaldo(s.cardNumber1, 2000, time.Date(s.testYear, 2, 15, 10, 0, 0, 0, time.UTC))
	s.insertSaldo(s.cardNumber2, 3000, time.Date(s.testYear, 2, 20, 10, 0, 0, 0, time.UTC))

	// Seed Topups
	s.insertTopup(s.cardNumber1, 500, time.Date(s.testYear, 1, 10, 10, 0, 0, 0, time.UTC))
	s.insertTopup(s.cardNumber1, 1000, time.Date(s.testYear, 2, 10, 10, 0, 0, 0, time.UTC))
	s.insertTopup(s.cardNumber2, 1500, time.Date(s.testYear, 2, 12, 10, 0, 0, 0, time.UTC))

	// Seed Withdraws
	s.insertWithdraw(s.cardNumber1, 100, time.Date(s.testYear, 1, 20, 10, 0, 0, 0, time.UTC))
	s.insertWithdraw(s.cardNumber1, 200, time.Date(s.testYear, 2, 20, 10, 0, 0, 0, time.UTC))
	s.insertWithdraw(s.cardNumber2, 300, time.Date(s.testYear, 2, 25, 10, 0, 0, 0, time.UTC))

	// Seed Transfers
	s.insertTransfer(s.cardNumber1, s.cardNumber2, 50, time.Date(s.testYear, 1, 25, 10, 0, 0, 0, time.UTC))
	s.insertTransfer(s.cardNumber1, s.cardNumber2, 100, time.Date(s.testYear, 2, 28, 10, 0, 0, 0, time.UTC))
}

func (s *CardStatsServiceTestSuite) insertSaldo(cardNumber string, amount int, t time.Time) {
	_, err := s.mongoDB.Collection("saldos").UpdateMany(context.Background(), bson.M{"card_number": cardNumber}, bson.M{"$set": bson.M{"deleted_at": time.Now()}})
	s.Require().NoError(err)
	s.Require().NoError(s.ts.InsertFixture(context.Background(), "saldos", "saldo_id", "", bson.M{"card_number": cardNumber, "total_balance": amount, "created_at": t, "updated_at": t}))
}

func (s *CardStatsServiceTestSuite) insertTopup(cardNumber string, amount int, t time.Time) {
	s.Require().NoError(s.ts.InsertFixture(context.Background(), "topups", "topup_id", "topup_no", bson.M{"card_number": cardNumber, "topup_amount": amount, "topup_time": t, "topup_method": "bank_transfer", "status": "success"}))
}

func (s *CardStatsServiceTestSuite) insertWithdraw(cardNumber string, amount int, t time.Time) {
	s.Require().NoError(s.ts.InsertFixture(context.Background(), "withdraws", "withdraw_id", "withdraw_no", bson.M{"card_number": cardNumber, "withdraw_amount": amount, "withdraw_time": t, "status": "success"}))
}

func (s *CardStatsServiceTestSuite) insertTransfer(from, to string, amount int, t time.Time) {
	s.Require().NoError(s.ts.InsertFixture(context.Background(), "transfers", "transfer_id", "transfer_no", bson.M{"transfer_from": from, "transfer_to": to, "transfer_amount": amount, "transfer_time": t, "status": "success"}))
}

func (s *CardStatsServiceTestSuite) TearDownSuite() {
	if s.ts != nil {
		s.ts.Teardown()
	}
}

// --- Balance Service Tests ---

func (s *CardStatsServiceTestSuite) TestBalanceService() {
	ctx := context.Background()

	// Global Monthly
	res, err := s.cardService.FindMonthlyBalance(ctx, s.testYear)
	s.NoError(err)
	s.NotEmpty(res)
	s.Equal(int32(1000), res[0].TotalBalance) // Jan
	s.Equal(int32(5000), res[1].TotalBalance) // Feb

	// By Card Monthly
	req1 := &requests.MonthYearCardNumberCard{CardNumber: s.cardNumber1, Year: s.testYear}
	res1, err := s.cardService.FindMonthlyBalancesByCardNumber(ctx, req1)
	s.NoError(err)
	s.Equal(int32(1000), res1[0].TotalBalance) // Jan Card 1
	s.Equal(int32(2000), res1[1].TotalBalance) // Feb Card 1

	// Yearly
	yRes, err := s.cardService.FindYearlyBalance(ctx, s.testYear)
	s.NoError(err)
	s.NotEmpty(yRes)
}

// --- Topup Service Tests ---

func (s *CardStatsServiceTestSuite) TestTopupService() {
	ctx := context.Background()

	// Global Monthly
	res, err := s.cardService.FindMonthlyTopupAmount(ctx, s.testYear)
	s.NoError(err)
	s.Equal(int32(500), res[0].TotalTopupAmount)
	s.Equal(int32(2500), res[1].TotalTopupAmount)

	// By Card Monthly
	req1 := &requests.MonthYearCardNumberCard{CardNumber: s.cardNumber1, Year: s.testYear}
	res1, err := s.cardService.FindMonthlyTopupAmountByCardNumber(ctx, req1)
	s.NoError(err)
	s.Equal(int32(500), res1[0].TotalTopupAmount)
	s.Equal(int32(1000), res1[1].TotalTopupAmount)
}

// --- Withdraw Service Tests ---

func (s *CardStatsServiceTestSuite) TestWithdrawService() {
	ctx := context.Background()

	// Global Monthly
	res, err := s.cardService.FindMonthlyWithdrawAmount(ctx, s.testYear)
	s.NoError(err)
	s.Equal(int32(100), res[0].TotalWithdrawAmount)
	s.Equal(int32(500), res[1].TotalWithdrawAmount)

	// By Card Monthly
	req1 := &requests.MonthYearCardNumberCard{CardNumber: s.cardNumber1, Year: s.testYear}
	res1, err := s.cardService.FindMonthlyWithdrawAmountByCardNumber(ctx, req1)
	s.NoError(err)
	s.Equal(int32(100), res1[0].TotalWithdrawAmount)
	s.Equal(int32(200), res1[1].TotalWithdrawAmount)
}

// --- Transfer Service Tests ---

func (s *CardStatsServiceTestSuite) TestTransferService() {
	ctx := context.Background()

	// Global Monthly Sender
	resS, err := s.cardService.FindMonthlyTransferAmountSender(ctx, s.testYear)
	s.NoError(err)
	s.Equal(int32(50), resS[0].TotalSentAmount)
	s.Equal(int32(100), resS[1].TotalSentAmount)

	// By Card Monthly Sender
	req1 := &requests.MonthYearCardNumberCard{CardNumber: s.cardNumber1, Year: s.testYear}
	res1S, err := s.cardService.FindMonthlyTransferAmountBySender(ctx, req1)
	s.NoError(err)
	s.Equal(int32(50), res1S[0].TotalSentAmount)
	s.Equal(int32(100), res1S[1].TotalSentAmount)

	// By Card Monthly Receiver
	req2 := &requests.MonthYearCardNumberCard{CardNumber: s.cardNumber2, Year: s.testYear}
	res2R, err := s.cardService.FindMonthlyTransferAmountByReceiver(ctx, req2)
	s.NoError(err)
	s.Equal(int32(50), res2R[0].TotalReceivedAmount)
	s.Equal(int32(100), res2R[1].TotalReceivedAmount)
}

func TestCardStatsServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(CardStatsServiceTestSuite))
}
