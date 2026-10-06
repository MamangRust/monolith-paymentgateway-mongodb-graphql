package merchant_test

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net"
	"testing"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-merchant/handler/stats"
	stats_cache "github.com/MamangRust/monolith-payment-gateway-merchant/redis/stats"
	apikey_cache "github.com/MamangRust/monolith-payment-gateway-merchant/redis/statsbyapikey"
	merchant_cache "github.com/MamangRust/monolith-payment-gateway-merchant/redis/statsbymerchant"
	stats_repo "github.com/MamangRust/monolith-payment-gateway-merchant/repository/stats"
	apikey_repo "github.com/MamangRust/monolith-payment-gateway-merchant/repository/statsbyapikey"
	merchant_repo "github.com/MamangRust/monolith-payment-gateway-merchant/repository/statsbymerchant"
	stats_service "github.com/MamangRust/monolith-payment-gateway-merchant/service/stats"
	apikey_service "github.com/MamangRust/monolith-payment-gateway-merchant/service/statsbyapikey"
	merchant_service "github.com/MamangRust/monolith-payment-gateway-merchant/service/statsbymerchant"
	pb_merchant "github.com/MamangRust/monolith-payment-gateway-pb/merchant"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/merchant/stats"
	logger_pkg "github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type MerchantStatsHandlerGapiTestSuite struct {
	suite.Suite
	ts         *tests.TestSuite
	lis        *bufconn.Listener
	conn       *grpc.ClientConn
	merchantID int32
	apiKey     string
	testYear   int
}

func (s *MerchantStatsHandlerGapiTestSuite) SetupSuite() {
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
	logger_pkg.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	logObj, _ := logger_pkg.NewLogger("test", lp)
	obs, _ := observability.NewObservability("test", logObj)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(redisClient, logObj, cacheMetrics)

	// Dependency Injection
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

	// Wrap in service.Service if needed, but handler constructor takes interfaces
	// Actually, MerchantStatsHandler constructor takes service.Service which might be too complex to mock entirely here
	// I'll check what service.Service has.
	// Oh, I see NewMerchantStatsHandler(service service.Service)
	// I'll create a mock or a partial service.Service.

	// I'll just use the real handler constructor if possible or create individual handlers.
	amountHandler := merchantstatshandler.NewMerchantStatsAmountHandler(svc, m_svc, a_svc)
	methodHandler := merchantstatshandler.NewMerchantStatsMethodHandler(svc, m_svc, a_svc)
	totalAmountHandler := merchantstatshandler.NewMerchantStatsTotalAmountHandler(svc, m_svc, a_svc)

	// gRPC Setup
	s.lis = bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	pb.RegisterMerchantStatsAmountServiceServer(grpcServer, amountHandler)
	pb.RegisterMerchantStatsMethodServiceServer(grpcServer, methodHandler)
	pb.RegisterMerchantStatsTotalAmountServiceServer(grpcServer, totalAmountHandler)

	go func() {
		if err := grpcServer.Serve(s.lis); err != nil {
			log.Printf("Server exited: %v", err)
		}
	}()

	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return s.lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.testYear = time.Now().Year()

	// Seed Data
	ctx := context.Background()
	_, err = mongoDB.Collection("users").InsertOne(ctx, bson.M{"user_id": 1, "firstname": "Gapi", "lastname": "Stats", "email": "gapi_stats@example.com", "password": "pass", "verification_code": "123", "is_verified": true})
	s.Require().NoError(err)

	s.apiKey = "gapi-key-1"
	s.merchantID = 1
	_, err = mongoDB.Collection("merchants").InsertOne(ctx, bson.M{"merchant_id": s.merchantID, "merchant_no": uuid.New(), "name": "Gapi Merchant", "api_key": s.apiKey, "status": "active"})
	s.Require().NoError(err)

	_, err = mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_id": 1, "card_number": "1212121212121212", "card_type": "debit", "cvv": "123", "card_provider": "visa", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)
	_, err = mongoDB.Collection("transactions").InsertOne(ctx, bson.M{"transaction_id": 1, "transaction_no": uuid.New(), "card_number": "1212121212121212", "amount": 2000, "payment_method": "visa", "merchant_id": s.merchantID, "transaction_time": time.Date(s.testYear, 1, 1, 10, 0, 0, 0, time.UTC), "status": "success"})
	s.Require().NoError(err)
}

func (s *MerchantStatsHandlerGapiTestSuite) TearDownSuite() {
	s.conn.Close()
	s.lis.Close()
	s.ts.Teardown()
}

func (s *MerchantStatsHandlerGapiTestSuite) TestAmountHandlers() {
	client := pb.NewMerchantStatsAmountServiceClient(s.conn)
	ctx := context.Background()

	// 1. Global Monthly
	res1, err := client.FindMonthlyAmountMerchant(ctx, &pb_merchant.FindYearMerchant{Year: int32(s.testYear)})
	s.Require().NoError(err)
	s.Require().Equal("success", res1.Status, "Message: %s", res1.Message)
	s.Require().NotEmpty(res1.Data)
	s.Require().EqualValues(2000, res1.Data[0].TotalAmount)

	// 2. By Merchant
	res2, err := client.FindMonthlyAmountByMerchants(ctx, &pb_merchant.FindYearMerchantById{
		Year:       int32(s.testYear),
		MerchantId: s.merchantID,
	})
	s.Require().NoError(err)
	s.Require().Equal("success", res2.Status, "Message: %s", res2.Message)
	s.Require().NotEmpty(res2.Data)
	s.Require().EqualValues(2000, res2.Data[0].TotalAmount)

	// 3. By Apikey
	res3, err := client.FindMonthlyAmountByApikey(ctx, &pb_merchant.FindYearMerchantByApikey{
		Year:   int32(s.testYear),
		ApiKey: s.apiKey,
	})
	s.Require().NoError(err)
	s.Require().Equal("success", res3.Status, "Message: %s", res3.Message)
	s.Require().NotEmpty(res3.Data)
	s.Require().EqualValues(2000, res3.Data[0].TotalAmount)
}

func (s *MerchantStatsHandlerGapiTestSuite) TestMethodHandlers() {
	client := pb.NewMerchantStatsMethodServiceClient(s.conn)
	ctx := context.Background()

	// 1. Global Monthly
	res1, err := client.FindMonthlyPaymentMethodsMerchant(ctx, &pb_merchant.FindYearMerchant{Year: int32(s.testYear)})
	s.Require().NoError(err)
	s.Require().Equal("success", res1.Status)
	s.Require().NotEmpty(res1.Data)
	// 'visa' should be at least one of the methods
	found := false
	for _, d := range res1.Data {
		if d.PaymentMethod == "visa" {
			found = true
			break
		}
	}
	s.True(found)
}

func (s *MerchantStatsHandlerGapiTestSuite) TestTotalAmountHandlers() {
	client := pb.NewMerchantStatsTotalAmountServiceClient(s.conn)
	ctx := context.Background()

	// 1. Global Monthly
	res1, err := client.FindMonthlyTotalAmountMerchant(ctx, &pb_merchant.FindYearMerchant{Year: int32(s.testYear)})
	s.Require().NoError(err)
	s.Require().Equal("success", res1.Status)
	s.Require().NotEmpty(res1.Data)
	s.Require().EqualValues(2000, res1.Data[0].TotalAmount)
}

func TestMerchantStatsHandlerGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(MerchantStatsHandlerGapiTestSuite))
}
