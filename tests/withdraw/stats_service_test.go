package withdraw_test

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	withdraw_cache "github.com/MamangRust/monolith-payment-gateway-withdraw/redis/stats"
	withdraw_cache_bycard "github.com/MamangRust/monolith-payment-gateway-withdraw/redis/statsbycard"
	withdraw_repo "github.com/MamangRust/monolith-payment-gateway-withdraw/repository/stats"
	withdraw_repo_bycard "github.com/MamangRust/monolith-payment-gateway-withdraw/repository/statsbycard"
	withdraw_service "github.com/MamangRust/monolith-payment-gateway-withdraw/service/stats"
	withdraw_service_bycard "github.com/MamangRust/monolith-payment-gateway-withdraw/service/statsbycard"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

type WithdrawStatsServiceTestSuite struct {
	suite.Suite
	ts        *tests.TestSuite
	svc       withdraw_service_bycard.WithdrawStatsByCardStatusService
	svcAmount withdraw_service_bycard.WithdrawStatsByCardAmountService
	testYear  int
}

func (s *WithdrawStatsServiceTestSuite) SetupSuite() {
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

	// Logger & Observability
	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	obs, _ := observability.NewObservability("test", log)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(redisClient, log, cacheMetrics)

	// Caches
	s_cache := withdraw_cache.NewWithdrawStatsCache(cacheStore)
	c_cache := withdraw_cache_bycard.NewWithdrawStatsByCardCache(cacheStore)

	// Repositories
	s_repo := withdraw_repo.NewWithdrawStatsStatusRepository(mongoDB)
	sa_repo := withdraw_repo.NewWithdrawStatsAmountRepository(mongoDB)
	c_repo := withdraw_repo_bycard.NewWithdrawStatsByCardStatusRepository(mongoDB)
	ca_repo := withdraw_repo_bycard.NewWithdrawStatsByCardAmountRepository(mongoDB)

	// Services
	s.svc = withdraw_service_bycard.NewWithdrawStatsByCardStatusService(&withdraw_service_bycard.WithdrawStatsByCardStatusDeps{
		Cache:         c_cache,
		Logger:        log,
		Repository:    c_repo,
		Observability: obs,
	})
	s.svcAmount = withdraw_service_bycard.NewWithdrawStatsByCardAmountService(&withdraw_service_bycard.WithdrawStatsByCardAmountDeps{
		Cache:         c_cache,
		Logger:        log,
		Repository:    ca_repo,
		Observability: obs,
	})

	_ = withdraw_service.NewWithdrawStatsStatusService(&withdraw_service.WithdrawStatsStatusDeps{
		Cache:         s_cache,
		Logger:        log,
		Repository:    s_repo,
		Observability: obs,
	})
	_ = withdraw_service.NewWithdrawStatsAmountService(&withdraw_service.WithdrawStatsAmountDeps{
		Cache:         s_cache,
		Logger:        log,
		Repository:    sa_repo,
		Observability: obs,
	})

	s.testYear = time.Now().Year()

	// Seed Data
	ctx := context.Background()
	_, err = mongoDB.Collection("users").InsertOne(ctx, bson.M{"firstname": "WithdrawSvc", "lastname": "Stats", "email": "withdraw_svc_stats@example.com", "password": "pass", "verification_code": "123", "is_verified": true})
	s.Require().NoError(err)

	cardNumber1 := "5555444433332222"
	_, err = mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_number": cardNumber1, "card_type": "debit", "cvv": "123", "card_provider": "visa", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)

	// Jan Success (2000)
	_, err = mongoDB.Collection("withdraws").InsertOne(ctx, bson.M{"card_number": cardNumber1, "withdraw_amount": 2000, "withdraw_time": time.Date(s.testYear, 1, 10, 10, 0, 0, 0, time.UTC), "status": "success"})
	s.Require().NoError(err)
}

func (s *WithdrawStatsServiceTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *WithdrawStatsServiceTestSuite) TestByCardService() {
	ctx := context.Background()
	cardNumber1 := "5555444433332222"
	reqStatus := &requests.MonthStatusWithdrawCardNumber{Year: s.testYear, Month: 1, CardNumber: cardNumber1}
	reqYearMonth := &requests.YearMonthCardNumber{Year: s.testYear, CardNumber: cardNumber1}

	resS, err := s.svc.FindMonthWithdrawStatusSuccessByCardNumber(ctx, reqStatus)
	s.NoError(err)
	s.NotEmpty(resS)

	resA, err := s.svcAmount.FindMonthlyWithdrawsByCardNumber(ctx, reqYearMonth)
	s.NoError(err)
	s.NotEmpty(resA)

	resY, err := s.svcAmount.FindYearlyWithdrawsByCardNumber(ctx, reqYearMonth)
	s.NoError(err)
	s.NotEmpty(resY)

}

func TestWithdrawStatsServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(WithdrawStatsServiceTestSuite))
}
