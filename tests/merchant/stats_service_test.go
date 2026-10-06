package merchant_test

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"time"

	stats_cache "github.com/MamangRust/monolith-payment-gateway-merchant/redis/stats"
	apikey_cache "github.com/MamangRust/monolith-payment-gateway-merchant/redis/statsbyapikey"
	merchant_cache "github.com/MamangRust/monolith-payment-gateway-merchant/redis/statsbymerchant"
	stats_repo "github.com/MamangRust/monolith-payment-gateway-merchant/repository/stats"
	apikey_repo "github.com/MamangRust/monolith-payment-gateway-merchant/repository/statsbyapikey"
	merchant_repo "github.com/MamangRust/monolith-payment-gateway-merchant/repository/statsbymerchant"
	stats_service "github.com/MamangRust/monolith-payment-gateway-merchant/service/stats"
	apikey_service "github.com/MamangRust/monolith-payment-gateway-merchant/service/statsbyapikey"
	merchant_service "github.com/MamangRust/monolith-payment-gateway-merchant/service/statsbymerchant"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

type MerchantStatsServiceTestSuite struct {
	suite.Suite
	ts              *tests.TestSuite
	service         stats_service.MerchantStatsAmountService
	apikeyService   apikey_service.MerchantStatsByApiKeyAmountService
	merchantService merchant_service.MerchantStatsByMerchantAmountService
	merchantID1     int32
	apiKey1         string
	testYear        int
}

func (s *MerchantStatsServiceTestSuite) SetupSuite() {
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
	s_cache := stats_cache.NewMerchantStatsCache(cacheStore)
	a_cache := apikey_cache.NewMerchantStatsByApiKeyCache(cacheStore)
	m_cache := merchant_cache.NewMerchantStatsByMerchantCache(cacheStore)

	// Repositories
	s_repo := stats_repo.NewMerchantStatsRepository(mongoDB)
	a_repo := apikey_repo.NewMerchantStatsByApiKeyRepository(mongoDB)
	m_repo := merchant_repo.NewMerchantStatsByMerchantRepository(mongoDB)

	// Services
	s.service = stats_service.NewMerchantStatsAmountService(&stats_service.MerchantStatsAmountDeps{
		Cache:         s_cache,
		Logger:        log,
		Repository:    s_repo,
		Observability: obs,
	})
	s.apikeyService = apikey_service.NewMerchantStatsAmountByApiKeyService(&apikey_service.MerchantStatsAmountByApiKeyDeps{
		Cache:         a_cache,
		Logger:        log,
		Repository:    a_repo,
		Observability: obs,
	})
	s.merchantService = merchant_service.NewMerchantStatsByMerchantAmountService(&merchant_service.MerchantStatsByMerchantAmountServiceDeps{
		Cache:         m_cache,
		Logger:        log,
		Repository:    m_repo,
		Observability: obs,
	})

	s.testYear = time.Now().Year()

	// Seed Data
	ctx := context.Background()
	_, err = mongoDB.Collection("users").InsertOne(ctx, bson.M{"user_id": 1, "firstname": "MerchantSvc", "lastname": "Stats", "email": "merchant_svc_stats@example.com", "password": "pass", "verification_code": "123", "is_verified": true})
	s.Require().NoError(err)

	s.apiKey1 = "svc-merchant-key-1"
	s.merchantID1 = 1
	_, err = mongoDB.Collection("merchants").InsertOne(ctx, bson.M{"merchant_id": s.merchantID1, "merchant_no": uuid.New(), "name": "Svc Merchant 1", "api_key": s.apiKey1, "status": "active"})
	s.Require().NoError(err)

	cardNumber1 := "9999888877776666"
	_, err = mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_id": 1, "card_number": cardNumber1, "card_type": "debit", "cvv": "123", "card_provider": "visa", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)

	// Transaction: Jan (1000)
	_, err = mongoDB.Collection("transactions").InsertOne(ctx, bson.M{"transaction_id": 1, "transaction_no": uuid.New(), "card_number": cardNumber1, "amount": 1000, "payment_method": "visa", "merchant_id": s.merchantID1, "transaction_time": time.Date(s.testYear, 1, 10, 10, 0, 0, 0, time.UTC), "status": "success"})
	s.Require().NoError(err)
}

func (s *MerchantStatsServiceTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *MerchantStatsServiceTestSuite) TestGlobalService() {
	ctx := context.Background()
	res, err := s.service.FindMonthlyAmountMerchant(ctx, s.testYear)
	s.NoError(err)
	s.NotEmpty(res)

}

func (s *MerchantStatsServiceTestSuite) TestMerchantService() {
	ctx := context.Background()
	res, err := s.merchantService.FindMonthlyAmountByMerchants(ctx, &requests.MonthYearAmountMerchant{Year: s.testYear, MerchantID: int(s.merchantID1)})
	s.NoError(err)
	s.NotEmpty(res)

}

func (s *MerchantStatsServiceTestSuite) TestApiKeyService() {
	ctx := context.Background()
	res, err := s.apikeyService.FindMonthlyAmountByApikey(ctx, &requests.MonthYearAmountApiKey{Year: s.testYear, Apikey: s.apiKey1})
	s.NoError(err)
	s.NotEmpty(res)

}

func TestMerchantStatsServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(MerchantStatsServiceTestSuite))
}
