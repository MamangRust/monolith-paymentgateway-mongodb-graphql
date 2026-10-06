package card_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-card/repository"
	"github.com/MamangRust/monolith-payment-gateway-card/service"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	user_repo "github.com/MamangRust/monolith-payment-gateway-user/repository"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

type CardRewardServiceTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	cardService service.Service
	cardNumber  string
	userID      int
}

func (s *CardRewardServiceTestSuite) SetupSuite() {
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
	repos := repository.NewRepositories(mongoDB, nil)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(redisClient, log, cacheMetrics)

	// The card service resolves users through its UserAdapter, so bring up the
	// in-process dependency services and wire the adapter to them.
	harness, err := tests.NewDependencyHarness(mongoDB, cacheStore, log)
	s.Require().NoError(err)
	s.T().Cleanup(harness.Close)

	s.cardService = service.NewService(&service.Deps{
		Repositories: repos,
		UserAdapter:  harness.UserAdapter,
		Logger:       log,
		Cache:        cacheStore,
		Kafka:        nil,
	})

	// Seed User
	userRepo := user_repo.NewUserCommandRepository(mongoDB.Collection("users"))
	user, err := userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Reward",
		LastName:  "Tester",
		Email:     fmt.Sprintf("reward.tester.%d.%d@example.com", time.Now().UnixNano(), time.Now().UnixNano()%10000),
		Password:  "password123",
	})
	s.Require().NoError(err)
	s.userID = int(user.UserID)

	// Seed Card
	card, err := s.cardService.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID:       s.userID,
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(2, 0, 0),
		CVV:          "123",
		CardProvider: "VISA",
	})
	s.Require().NoError(err)
	s.NotNil(card)
	s.cardNumber = card.CardNumber
}

func (s *CardRewardServiceTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *CardRewardServiceTestSuite) Test1_EarnRewards() {
	s.Require().NotEmpty(s.cardNumber)
	ctx := context.Background()
	req := &requests.EarnRewardsRequest{
		CardNumber: s.cardNumber,
		TxnID:      fmt.Sprintf("txn-%d", time.Now().UnixNano()),
		Amount:     50000,
		Mcc:        "5411",
	}

	res, err := s.cardService.EarnRewards(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(s.cardNumber, res.CardNumber)
}

func (s *CardRewardServiceTestSuite) Test2_GetBalance() {
	s.Require().NotEmpty(s.cardNumber)
	ctx := context.Background()

	balance, err := s.cardService.GetBalance(ctx, s.cardNumber)
	s.NoError(err)
	s.Positive(balance)
}

func (s *CardRewardServiceTestSuite) Test3_GetHistory() {
	s.Require().NotEmpty(s.cardNumber)
	ctx := context.Background()

	history, err := s.cardService.GetHistory(ctx, s.cardNumber)
	s.NoError(err)
	s.NotNil(history)
	s.GreaterOrEqual(len(history), 1)
}

func (s *CardRewardServiceTestSuite) Test4_RedeemRewards() {
	s.Require().NotEmpty(s.cardNumber)
	ctx := context.Background()

	// Get current balance first
	balance, err := s.cardService.GetBalance(ctx, s.cardNumber)
	s.Require().NoError(err)
	s.Require().Positive(balance)

	// Redeem half the points. Redemption consumes whole reward ledger entries,
	// so RedeemRewards returns the redeemed points (>= requested), not the
	// remaining balance. Verify both the redeemed amount and the new balance.
	pointsToRedeem := balance / 2
	redeemed, err := s.cardService.RedeemRewards(ctx, s.cardNumber, pointsToRedeem)
	s.NoError(err)
	s.GreaterOrEqual(redeemed, pointsToRedeem)

	newBalance, err := s.cardService.GetBalance(ctx, s.cardNumber)
	s.NoError(err)
	s.Equal(balance-redeemed, newBalance)
}

func TestCardRewardServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(CardRewardServiceTestSuite))
}
