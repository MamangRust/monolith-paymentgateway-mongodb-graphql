package card_test

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
	"time"

	sharederrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"

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

type CardPaymentServiceTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	cardService service.Service
	mongoDB     *mongo.Database
	userID      int
	cardNumber  string
}

func (s *CardPaymentServiceTestSuite) SetupSuite() {
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
		FirstName: "Payment",
		LastName:  "Tester",
		Email:     fmt.Sprintf("payment.tester.%d.%d@example.com", time.Now().UnixNano(), time.Now().UnixNano()%10000),
		Password:  "password123",
	})
	s.Require().NoError(err)
	s.userID = int(user.UserID)

	// Seed Card
	cardReq := &requests.CreateCardRequest{
		UserID:       s.userID,
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(2, 0, 0),
		CVV:          "123",
		CardProvider: "VISA",
	}
	card, err := s.cardService.CreateCard(context.Background(), cardReq)
	s.Require().NoError(err)

	// Get the full card details to extract the card number
	cardDetails, err := s.cardService.FindById(context.Background(), int(card.CardID))
	s.Require().NoError(err)
	s.cardNumber = cardDetails.CardNumber
}

func (s *CardPaymentServiceTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *CardPaymentServiceTestSuite) Test1_PostPayment() {
	ctx := context.Background()
	req := &requests.PostPaymentRequest{
		ReferenceID:    fmt.Sprintf("ref-%d", time.Now().UnixNano()),
		CardNumber:     s.cardNumber,
		Amount:         100000,
		PaymentChannel: "bank_transfer",
		BillingID:      nil,
	}

	res, err := s.cardService.PostPayment(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(s.cardNumber, res.CardNumber)
	s.Equal(int64(100000), res.Amount)
}

func (s *CardPaymentServiceTestSuite) Test2_IdempotentReplayAndPayloadConflict() {
	ctx := context.Background()
	referenceID := fmt.Sprintf("payment-replay-%d", time.Now().UnixNano())
	firstRequest := &requests.PostPaymentRequest{
		ReferenceID:    referenceID,
		CardNumber:     s.cardNumber,
		Amount:         25000,
		PaymentChannel: "bank_transfer",
	}

	first, err := s.cardService.PostPayment(ctx, firstRequest)
	s.Require().NoError(err)

	second, err := s.cardService.PostPayment(ctx, firstRequest)
	s.Require().NoError(err)
	s.Equal(first.PaymentID, second.PaymentID)

	conflictingRequest := *firstRequest
	conflictingRequest.Amount = firstRequest.Amount + 1
	_, err = s.cardService.PostPayment(ctx, &conflictingRequest)
	s.Require().Error(err)
	appErr, ok := err.(*sharederrors.AppError)
	s.Require().True(ok)
	s.Equal(sharederrors.ErrorTypeConflict, appErr.Type)

	paymentCount, err := s.mongoDB.Collection("card_payments").CountDocuments(ctx, bson.M{"reference_id": referenceID})
	s.Require().NoError(err)
	s.Equal(int64(1), paymentCount)
}

func (s *CardPaymentServiceTestSuite) Test3_GetPaymentHistory() {
	ctx := context.Background()

	res, err := s.cardService.GetPaymentHistory(ctx, s.cardNumber, 1, 10)
	s.NoError(err)
	s.NotNil(res)
	s.GreaterOrEqual(len(res), 1)
}

func (s *CardPaymentServiceTestSuite) Test4_CountPayments() {
	ctx := context.Background()

	total, err := s.cardService.CountPayments(ctx, s.cardNumber)
	s.NoError(err)
	s.GreaterOrEqual(total, 1)
}

func TestCardPaymentServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(CardPaymentServiceTestSuite))
}
