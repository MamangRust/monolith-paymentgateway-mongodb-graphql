package card_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	card_handler "github.com/MamangRust/monolith-payment-gateway-card/handler"
	card_repository "github.com/MamangRust/monolith-payment-gateway-card/repository"
	card_service "github.com/MamangRust/monolith-payment-gateway-card/service"
	pbcard "github.com/MamangRust/monolith-payment-gateway-pb/card"
	pbuser "github.com/MamangRust/monolith-payment-gateway-pb/user"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CardApiTestSuite struct {
	suite.Suite
	ts         *tests.TestSuite
	harness    *tests.DependencyHarness
	grpcServer *grpc.Server
	conn       *grpc.ClientConn
	gql        http.Handler
	userID     int
	cardID     int
	cardNumber string
}

func (s *CardApiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	redisClient := redis.NewClient(opts)

	mongoDB, err := s.ts.MongoDB()
	s.Require().NoError(err)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(redisClient, log, cacheMetrics)

	// The card repositories resolve users over gRPC, so run the shared
	// dependency harness (user + role + card + saldo + merchant) in-process
	// and reuse its user adapter.
	harness, err := tests.NewDependencyHarness(mongoDB, cacheStore, log)
	s.Require().NoError(err)
	s.harness = harness

	// Create a user to own the card.
	userRes, err := pbuser.NewUserCommandServiceClient(harness.Conn).Create(context.Background(), &pbuser.CreateUserRequest{
		Firstname:       "Card",
		Lastname:        "Api",
		Email:           fmt.Sprintf("card.api.%d.%d@example.com", time.Now().UnixNano(), time.Now().UnixNano()%1000),
		Password:        "Password123!",
		ConfirmPassword: "Password123!",
	})
	s.Require().NoError(err)
	s.userID = int(userRes.Data.Id)

	repos := card_repository.NewRepositories(mongoDB, nil)
	cardSvc := card_service.NewService(&card_service.Deps{
		Cache:        cacheStore,
		Repositories: repos,
		UserAdapter:  harness.UserAdapter,
		Logger:       log,
		Kafka:        nil,
	})
	cardH := card_handler.NewHandler(cardSvc)

	// Start the card gRPC service
	s.grpcServer = grpc.NewServer()
	pbcard.RegisterCardQueryServiceServer(s.grpcServer, cardH)
	pbcard.RegisterCardCommandServiceServer(s.grpcServer, cardH)

	addr, err := tests.RunGRPCServer(s.grpcServer)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.gql = tests.NewGraphQLHandler(map[string]*grpc.ClientConn{"card": conn}, redisClient, log)
}

func (s *CardApiTestSuite) TearDownSuite() {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
	if s.harness != nil {
		s.harness.Close()
	}
	if s.ts != nil {
		s.ts.Teardown()
	}
}

func (s *CardApiTestSuite) Test1_CardLifecycle() {
	expireDate := time.Now().AddDate(5, 0, 0).Format("2006-01-02")

	// Create
	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation {
		createCard(input: { user_id: %d, card_type: "debit", expire_date: "%s", cvv: "123", card_provider: "visa" }) {
			status message
			data { id user_id card_number card_type card_provider }
		}
	}`, s.userID, expireDate))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data := tests.GraphQLOpEnvelopeData(body, "createCard")
	s.Require().NotNil(data)
	s.Equal("debit", data["card_type"])
	s.Equal(float64(s.userID), data["user_id"])
	s.cardID = int(data["id"].(float64))
	s.cardNumber = data["card_number"].(string)
	s.Require().NotZero(s.cardID)
	s.Require().NotEmpty(s.cardNumber)

	// FindById
	body, err = tests.DoGraphQL(s.gql, fmt.Sprintf(`query { findByIdCard(input: { card_id: %d }) { status message data { id card_number } } }`, s.cardID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data = tests.GraphQLOpEnvelopeData(body, "findByIdCard")
	s.Require().NotNil(data)
	s.Equal(float64(s.cardID), data["id"])

	// Update
	body, err = tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation {
		updateCard(input: { card_id: %d, user_id: %d, card_type: "credit", expire_date: "%s", cvv: "456", card_provider: "mastercard" }) {
			status message
			data { id card_type card_provider }
		}
	}`, s.cardID, s.userID, expireDate))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data = tests.GraphQLOpEnvelopeData(body, "updateCard")
	s.Require().NotNil(data)
	s.Equal("credit", data["card_type"])
}

func (s *CardApiTestSuite) Test2_QueryOperations() {
	s.Require().NotZero(s.cardID)
	s.Require().NotEmpty(s.cardNumber)

	// FindAll
	body, err := tests.DoGraphQL(s.gql, `query { findAllCard(input: { page: 1, page_size: 10 }) { status message data { id card_number } } }`)
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	s.Require().NotNil(tests.GraphQLOpData(body, "findAllCard"))

	// FindByActive
	body, err = tests.DoGraphQL(s.gql, `query { findByActiveCard(input: { page: 1, page_size: 10 }) { status message data { id card_number } } }`)
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	s.Require().NotNil(tests.GraphQLOpData(body, "findByActiveCard"))

	// FindByUserId
	body, err = tests.DoGraphQL(s.gql, fmt.Sprintf(`query { findByUserIdCard(input: { user_id: %d }) { status message data { id card_number } } }`, s.userID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data := tests.GraphQLOpEnvelopeData(body, "findByUserIdCard")
	s.Require().NotNil(data)
	s.Equal(s.cardNumber, data["card_number"])

	// FindByCardNumber
	body, err = tests.DoGraphQL(s.gql, fmt.Sprintf(`query { findByCardNumber(input: { card_number: "%s" }) { status message data { id card_number } } }`, s.cardNumber))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data = tests.GraphQLOpEnvelopeData(body, "findByCardNumber")
	s.Require().NotNil(data)
	s.Equal(s.cardNumber, data["card_number"])
}

func (s *CardApiTestSuite) Test3_TrashAndRestore() {
	s.Require().NotZero(s.cardID)

	// Trash
	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation { trashedCard(input: { card_id: %d }) { status message data { id } } }`, s.cardID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))

	// FindTrashed
	body, err = tests.DoGraphQL(s.gql, `query { findByTrashedCard(input: { page: 1, page_size: 10 }) { status message data { id } } }`)
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	s.Require().NotNil(tests.GraphQLOpData(body, "findByTrashedCard"))

	// Restore
	body, err = tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation { restoreCard(input: { card_id: %d }) { status message data { id } } }`, s.cardID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
}

func TestCardApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(CardApiTestSuite))
}
