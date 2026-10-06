package merchant_test

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	merchantstatshandler "github.com/MamangRust/monolith-payment-gateway-merchant/handler/stats"
	stats_cache "github.com/MamangRust/monolith-payment-gateway-merchant/redis/stats"
	apikey_cache "github.com/MamangRust/monolith-payment-gateway-merchant/redis/statsbyapikey"
	merchant_cache "github.com/MamangRust/monolith-payment-gateway-merchant/redis/statsbymerchant"
	stats_repo "github.com/MamangRust/monolith-payment-gateway-merchant/repository/stats"
	apikey_repo "github.com/MamangRust/monolith-payment-gateway-merchant/repository/statsbyapikey"
	merchant_repo "github.com/MamangRust/monolith-payment-gateway-merchant/repository/statsbymerchant"
	stats_service "github.com/MamangRust/monolith-payment-gateway-merchant/service/stats"
	apikey_service "github.com/MamangRust/monolith-payment-gateway-merchant/service/statsbyapikey"
	merchant_service "github.com/MamangRust/monolith-payment-gateway-merchant/service/statsbymerchant"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/merchant/stats"
	logger_pkg "github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MerchantStatsHandlerApiTestSuite struct {
	suite.Suite
	ts         *tests.TestSuite
	grpcServer *grpc.Server
	conn       *grpc.ClientConn
	gql        http.Handler
	merchantID int32
	apiKey     string
	testYear   int
}

func (s *MerchantStatsHandlerApiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	redisClient := redis.NewClient(opts)

	mongoDB, err := s.ts.MongoDB()
	s.Require().NoError(err)

	// Logger & Observability
	logger_pkg.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	logObj, _ := logger_pkg.NewLogger("test", lp)
	obs, _ := observability.NewObservability("test", logObj)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(redisClient, logObj, cacheMetrics)

	// Dependency Injection (Merchant Service side)
	s_cache := stats_cache.NewMerchantStatsCache(cacheStore)
	a_cache := apikey_cache.NewMerchantStatsByApiKeyCache(cacheStore)
	m_cache := merchant_cache.NewMerchantStatsByMerchantCache(cacheStore)

	s_repo := stats_repo.NewMerchantStatsRepository(mongoDB)
	a_repo := apikey_repo.NewMerchantStatsByApiKeyRepository(mongoDB)
	m_repo := merchant_repo.NewMerchantStatsByMerchantRepository(mongoDB)

	svc := stats_service.NewMerchantStatsService(&stats_service.DepsStats{
		Mencache:      s_cache,
		Logger:        logObj,
		Repository:    s_repo,
		Observability: obs,
	})
	a_svc := apikey_service.NewMerchantStatsByApiKeyService(&apikey_service.DepsStatsByApiKey{
		Mencache:      a_cache,
		Logger:        logObj,
		Repository:    a_repo,
		Observability: obs,
	})
	m_svc := merchant_service.NewMerchantStatsByMerchantService(&merchant_service.DepsStatsByMerchant{
		Mencache:      m_cache,
		Logger:        logObj,
		Repository:    m_repo,
		Observability: obs,
	})

	amountHandler := merchantstatshandler.NewMerchantStatsAmountHandler(svc, m_svc, a_svc)
	methodHandler := merchantstatshandler.NewMerchantStatsMethodHandler(svc, m_svc, a_svc)
	totalAmountHandler := merchantstatshandler.NewMerchantStatsTotalAmountHandler(svc, m_svc, a_svc)

	// gRPC Setup
	s.grpcServer = grpc.NewServer()
	pb.RegisterMerchantStatsAmountServiceServer(s.grpcServer, amountHandler)
	pb.RegisterMerchantStatsMethodServiceServer(s.grpcServer, methodHandler)
	pb.RegisterMerchantStatsTotalAmountServiceServer(s.grpcServer, totalAmountHandler)

	addr, err := tests.RunGRPCServer(s.grpcServer)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.gql = tests.NewGraphQLHandler(map[string]*grpc.ClientConn{"merchant": conn}, redisClient, logObj)

	s.testYear = time.Now().Year()

	// Seed Data
	ctx := context.Background()
	_, err = mongoDB.Collection("users").InsertOne(ctx, bson.M{"user_id": 1, "firstname": "Api", "lastname": "Stats", "email": fmt.Sprintf("api.stats.%d@example.com", time.Now().UnixNano()), "password": "pass", "verification_code": "123", "is_verified": true})
	s.Require().NoError(err)

	s.apiKey = "api-key-stats"
	s.merchantID = 1
	_, err = mongoDB.Collection("merchants").InsertOne(ctx, bson.M{"merchant_id": s.merchantID, "merchant_no": uuid.New(), "name": "Api Stats Merchant", "api_key": s.apiKey, "status": "active"})
	s.Require().NoError(err)

	_, err = mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_id": 1, "card_number": "9999999999999999", "card_type": "debit", "cvv": "123", "card_provider": "visa", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)
	_, err = mongoDB.Collection("transactions").InsertOne(ctx, bson.M{"transaction_id": 1, "transaction_no": uuid.New(), "card_number": "9999999999999999", "amount": 3000, "payment_method": "mastercard", "merchant_id": s.merchantID, "transaction_time": time.Date(s.testYear, 2, 1, 10, 0, 0, 0, time.UTC), "status": "success"})
	s.Require().NoError(err)
}

func (s *MerchantStatsHandlerApiTestSuite) TearDownSuite() {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
	if s.ts != nil {
		s.ts.Teardown()
	}
}

func (s *MerchantStatsHandlerApiTestSuite) TestAmountApi() {
	year := strconv.Itoa(s.testYear)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`query {
		findMonthlyAmountMerchant(input: { year: %s }) {
			status message
			data { month total_amount }
		}
	}`, year))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	opRes := tests.GraphQLOpData(body, "findMonthlyAmountMerchant")
	s.Require().NotNil(opRes)
	s.Equal("success", opRes["status"])
	payload, ok := opRes["data"].([]any)
	s.Require().True(ok, "expected monthly amount data array, got %+v", opRes["data"])
	s.Require().NotEmpty(payload)
	// Feb is index 1
	febData := payload[1].(map[string]any)
	s.EqualValues(3000, febData["total_amount"])
}

func (s *MerchantStatsHandlerApiTestSuite) TestMethodApi() {
	year := strconv.Itoa(s.testYear)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`query {
		findMonthlyPaymentMethodsMerchant(input: { year: %s }) {
			status message
			data { month payment_method total_amount }
		}
	}`, year))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	opRes := tests.GraphQLOpData(body, "findMonthlyPaymentMethodsMerchant")
	s.Require().NotNil(opRes)
	s.Equal("success", opRes["status"])
	payload, ok := opRes["data"].([]any)
	s.Require().True(ok, "expected monthly payment methods data array, got %+v", opRes["data"])
	s.Require().NotEmpty(payload)

	// Check for 'mastercard' in Feb
	found := false
	for _, d := range payload {
		m := d.(map[string]any)
		if m["payment_method"] == "mastercard" && m["month"] == "Feb" {
			found = true
			break
		}
	}
	s.True(found)
}

func (s *MerchantStatsHandlerApiTestSuite) TestTotalAmountApi() {
	year := strconv.Itoa(s.testYear)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`query {
		findMonthlyTotalAmountMerchant(input: { year: %s }) {
			status message
			data { month year total_amount }
		}
	}`, year))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	opRes := tests.GraphQLOpData(body, "findMonthlyTotalAmountMerchant")
	s.Require().NotNil(opRes)
	s.Equal("success", opRes["status"])
	payload, ok := opRes["data"].([]any)
	s.Require().True(ok, "expected monthly total amount data array, got %+v", opRes["data"])
	s.Require().NotEmpty(payload)
	febData := payload[1].(map[string]any)
	s.EqualValues(3000, febData["total_amount"])
}

func TestMerchantStatsHandlerApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(MerchantStatsHandlerApiTestSuite))
}
