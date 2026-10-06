package topup_test

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net"
	"testing"
	"time"

	pbtopup "github.com/MamangRust/monolith-payment-gateway-pb/topup"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/topup/stats"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/MamangRust/monolith-payment-gateway-topup/handler"
	"github.com/MamangRust/monolith-payment-gateway-topup/repository"
	"github.com/MamangRust/monolith-payment-gateway-topup/service"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	card_repo "github.com/MamangRust/monolith-payment-gateway-card/repository"
	saldo_repo "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
)

type TopupStatsHandlerGapiTestSuite struct {
	suite.Suite
	ts           *tests.TestSuite
	lis          *bufconn.Listener
	conn         *grpc.ClientConn
	amountClient pb.TopupStatsAmountServiceClient
	statusClient pb.TopupStatsStatusServiceClient
	userID       int32
	cardNumber1  string
	testYear     int
}

func (s *TopupStatsHandlerGapiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts
	mongoDB, err := s.ts.MongoDB()
	if err != nil {
		s.Require().NoError(err)
	}

	realSaldo := &realSaldoRepo{repo: saldo_repo.NewRepositories(mongoDB, nil, nil)}
	realCard := &realCardRepo{
		query: card_repo.NewCardQueryRepository(mongoDB.Collection("cards")),
		cmd:   card_repo.NewCardCommandRepository(mongoDB.Collection("cards")),
	}

	repos := repository.NewRepositoriesWithAdapters(mongoDB, realCard, realSaldo)

	zapLog := zap.NewNop()
	myLogger := &logger.Logger{Log: zapLog}

	redisOption, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	redisClient := redis.NewClient(redisOption)
	cacheStore := cache.NewCacheStore(redisClient, myLogger, &dummyCacheMetrics{})

	svc := service.NewService(&service.Deps{
		Kafka:        nil,
		Repositories: repos,
		Logger:       myLogger,
		Cache:        cacheStore,
		DB:           mongoDB,
	})

	h := handler.NewHandler(svc)

	s.lis = bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	pb.RegisterTopupStatsAmountServiceServer(server, h)
	pb.RegisterTopupStatsStatusServiceServer(server, h)

	go func() {
		if err := server.Serve(s.lis); err != nil {
			log.Printf("Server exited with error: %v", err)
		}
	}()

	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return s.lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.amountClient = pb.NewTopupStatsAmountServiceClient(conn)
	s.statusClient = pb.NewTopupStatsStatusServiceClient(conn)

	s.testYear = time.Now().Year()

	// Seed Data
	ctx := context.Background()
	_, err = mongoDB.Collection("users").InsertOne(ctx, bson.M{"user_id": 9230, "firstname": "TopupGapi", "lastname": "Stats", "email": "topup_gapi_stats@example.com", "password": "pass", "verification_code": "123", "is_verified": true})
	s.Require().NoError(err)

	s.cardNumber1 = "3333444455556666"
	_, err = mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_id": 9230, "card_number": s.cardNumber1, "card_type": "debit", "cvv": "123", "card_provider": "visa", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)

	_, err = mongoDB.Collection("topups").InsertOne(ctx, bson.M{"topup_id": 9331, "topup_no": "seed-topup-9331", "card_number": s.cardNumber1, "topup_amount": 5000, "topup_method": "bank_transfer", "topup_time": time.Date(s.testYear, 1, 10, 10, 0, 0, 0, time.UTC), "status": "success"})
	s.Require().NoError(err)
}

func (s *TopupStatsHandlerGapiTestSuite) TearDownSuite() {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.lis != nil {
		s.lis.Close()
	}
	s.ts.Teardown()
}

func (s *TopupStatsHandlerGapiTestSuite) TestFindMonthlyTopupAmounts() {
	ctx := context.Background()
	resp, err := s.amountClient.FindMonthlyTopupAmounts(ctx, &pbtopup.FindYearTopupStatus{
		Year: int32(s.testYear),
	})
	s.NoError(err)
	s.Equal("success", resp.Status)
	s.NotEmpty(resp.Data)
	s.Equal(int32(5000), resp.Data[0].TotalAmount)
}

func (s *TopupStatsHandlerGapiTestSuite) TestFindMonthlyTopupAmountsByCardNumber() {
	ctx := context.Background()
	resp, err := s.amountClient.FindMonthlyTopupAmountsByCardNumber(ctx, &pbtopup.FindYearTopupCardNumber{
		Year:       int32(s.testYear),
		CardNumber: s.cardNumber1,
	})
	s.NoError(err)
	s.Equal("success", resp.Status)
	s.NotEmpty(resp.Data)
	s.Equal(int32(5000), resp.Data[0].TotalAmount)
}

func TestTopupStatsHandlerGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TopupStatsHandlerGapiTestSuite))
}
