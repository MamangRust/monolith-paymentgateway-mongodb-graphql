package card_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/MamangRust/monolith-payment-gateway-card/repository"
	"github.com/MamangRust/monolith-payment-gateway-card/service"
	pbcard "github.com/MamangRust/monolith-payment-gateway-pb/card"
	pbuser "github.com/MamangRust/monolith-payment-gateway-pb/user"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CardAuthServiceTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	cardService service.Service
	mongoDB     *mongo.Database
	cardNumber  string
	userID      int
	txnID       string
	harness     *tests.DependencyHarness
}

func (s *CardAuthServiceTestSuite) SetupSuite() {
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

	harness, err := tests.NewDependencyHarness(mongoDB, cacheStore, log)
	s.Require().NoError(err)
	s.harness = harness

	s.cardService = service.NewService(&service.Deps{
		Repositories: repos,
		UserAdapter:  harness.UserAdapter,
		Logger:       log,
		Cache:        cacheStore,
		Kafka:        nil,
	})

	// Seed User via the harness card service (which has UserAdapter)
	cardClient := pbcard.NewCardCommandServiceClient(harness.Conn)
	userClient := pbuser.NewUserCommandServiceClient(harness.Conn)

	userRes, err := userClient.Create(context.Background(), &pbuser.CreateUserRequest{
		Firstname:       "Auth",
		Lastname:        "Tester",
		Email:           fmt.Sprintf("auth.tester.%d@example.com", time.Now().UnixNano()),
		Password:        "password123",
		ConfirmPassword: "password123",
	})
	s.Require().NoError(err)
	s.userID = int(userRes.Data.Id)

	// Seed Card
	cardRes, err := cardClient.CreateCard(context.Background(), &pbcard.CreateCardRequest{
		UserId:       int32(s.userID),
		CardType:     "debit",
		ExpireDate:   timestamppb.New(time.Now().AddDate(2, 0, 0)),
		Cvv:          "123",
		CardProvider: "VISA",
	})
	s.Require().NoError(err)
	s.cardNumber = cardRes.Data.CardNumber
}

func (s *CardAuthServiceTestSuite) TearDownSuite() {
	if s.harness != nil {
		s.harness.Close()
	}
	s.ts.Teardown()
}

func (s *CardAuthServiceTestSuite) Test1_Authorize() {
	s.Require().NotEmpty(s.cardNumber)
	ctx := context.Background()

	req := &requests.AuthorizeCardRequest{
		CardNumber:     s.cardNumber,
		MerchantID:     1,
		Amount:         50000,
		Currency:       "USD",
		PosEntryMode:   "05",
		Mcc:            "5812",
		IdempotencyKey: fmt.Sprintf("auth-idem-%d", time.Now().UnixNano()),
	}

	res, err := s.cardService.Authorize(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(s.cardNumber, res.CardNumber)
	s.Equal("pending", res.Status)
	s.NotEmpty(res.TxnID)
	s.txnID = res.TxnID
}

func (s *CardAuthServiceTestSuite) Test2_GetAuthTransaction() {
	s.Require().NotEmpty(s.txnID)
	ctx := context.Background()

	res, err := s.cardService.GetAuthTransaction(ctx, s.txnID)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(s.txnID, res.TxnID)
	s.Equal(s.cardNumber, res.CardNumber)
}

func (s *CardAuthServiceTestSuite) Test3_Reverse() {
	s.Require().NotEmpty(s.txnID)
	s.Require().NotEmpty(s.cardNumber)
	ctx := context.Background()

	// The auth state machine only permits reversing approved transactions:
	// pending -> approved -> reversed. Authorize leaves the transaction
	// "pending", so approve it before attempting the reversal.
	if _, err := s.mongoDB.Collection("card_auth_transactions").UpdateOne(ctx, bson.M{"txn_id": s.txnID}, bson.M{"$set": bson.M{"status": "approved", "updated_at": time.Now()}}); err != nil {
		s.Require().NoError(err)
	}

	req := &requests.ReverseTransactionRequest{
		TxnID:          s.txnID,
		CardNumber:     s.cardNumber,
		Amount:         50000,
		IdempotencyKey: fmt.Sprintf("rev-idem-%d", time.Now().UnixNano()),
	}

	res, err := s.cardService.Reverse(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(s.txnID, res.TxnID)
	s.Equal("reversed", res.Status)
}

func (s *CardAuthServiceTestSuite) Test4_Authorize_Idempotency() {
	s.Require().NotEmpty(s.cardNumber)
	ctx := context.Background()

	idempotencyKey := fmt.Sprintf("idem-test-%d", time.Now().UnixNano())
	req := &requests.AuthorizeCardRequest{
		CardNumber:     s.cardNumber,
		MerchantID:     2,
		Amount:         25000,
		Currency:       "USD",
		PosEntryMode:   "07",
		Mcc:            "5411",
		IdempotencyKey: idempotencyKey,
	}

	first, err := s.cardService.Authorize(ctx, req)
	s.NoError(err)
	s.NotNil(first)

	second, err := s.cardService.Authorize(ctx, req)
	s.NoError(err)
	s.NotNil(second)

	s.Equal(first.TxnID, second.TxnID)
	s.Equal(first.Amount, second.Amount)
}

func TestCardAuthServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(CardAuthServiceTestSuite))
}
