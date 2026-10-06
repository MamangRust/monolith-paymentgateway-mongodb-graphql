package saldo_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	pbcard "github.com/MamangRust/monolith-payment-gateway-pb/card"
	pbsaldo "github.com/MamangRust/monolith-payment-gateway-pb/saldo"
	pbuser "github.com/MamangRust/monolith-payment-gateway-pb/user"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	saldo_handler "github.com/MamangRust/monolith-payment-gateway-saldo/handler"
	saldo_repository "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	saldo_service "github.com/MamangRust/monolith-payment-gateway-saldo/service"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SaldoApiTestSuite struct {
	suite.Suite
	ts         *tests.TestSuite
	harness    *tests.DependencyHarness
	grpcServer *grpc.Server
	conn       *grpc.ClientConn
	gql        http.Handler
	userID     int
	cardID     int
	cardNumber string
	saldoID    int
}

func (s *SaldoApiTestSuite) SetupSuite() {
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

	// The saldo repositories resolve cards over gRPC, so run the shared
	// dependency harness in-process and reuse its card adapter.
	harness, err := tests.NewDependencyHarness(mongoDB, cacheStore, log)
	s.Require().NoError(err)
	s.harness = harness

	repos := saldo_repository.NewRepositories(mongoDB, nil, nil)
	saldoSvc := saldo_service.NewService(&saldo_service.Deps{
		Cache:        cacheStore,
		Repositories: repos,
		CardAdapter:  harness.CardAdapter,
		Logger:       log,
		DB:           mongoDB,
	})
	saldoH := saldo_handler.NewHandler(saldoSvc)

	// Create user + card directly (the saldo gateway ops only need the card number)
	ctx := context.Background()
	userRes, err := pbuser.NewUserCommandServiceClient(harness.Conn).Create(ctx, &pbuser.CreateUserRequest{
		Firstname:       "Saldo",
		Lastname:        "Api",
		Email:           fmt.Sprintf("saldo.api.%d@example.com", time.Now().UnixNano()),
		Password:        "Password123!",
		ConfirmPassword: "Password123!",
	})
	s.Require().NoError(err)
	s.userID = int(userRes.Data.Id)

	cardRes, err := pbcard.NewCardCommandServiceClient(harness.Conn).CreateCard(ctx, &pbcard.CreateCardRequest{
		UserId:       userRes.Data.Id,
		CardType:     "debit",
		ExpireDate:   timestamppb.New(time.Now().AddDate(5, 0, 0)),
		Cvv:          "123",
		CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.cardID = int(cardRes.Data.Id)
	s.cardNumber = cardRes.Data.CardNumber

	// Start the saldo gRPC service
	s.grpcServer = grpc.NewServer()
	pbsaldo.RegisterSaldoQueryServiceServer(s.grpcServer, saldoH)
	pbsaldo.RegisterSaldoCommandServiceServer(s.grpcServer, saldoH)

	addr, err := tests.RunGRPCServer(s.grpcServer)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.gql = tests.NewGraphQLHandler(map[string]*grpc.ClientConn{"saldo": conn}, redisClient, log)
}

func (s *SaldoApiTestSuite) TearDownSuite() {
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

func (s *SaldoApiTestSuite) Test1_SaldoLifecycle() {
	// Create
	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation {
		createSaldo(input: { card_number: "%s", total_balance: 100000 }) {
			status message
			data { saldo_id card_number total_balance }
		}
	}`, s.cardNumber))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data := tests.GraphQLOpEnvelopeData(body, "createSaldo")
	s.Require().NotNil(data)
	s.Equal(float64(100000), data["total_balance"])
	s.saldoID = int(data["saldo_id"].(float64))
	s.Require().NotZero(s.saldoID)

	// FindById
	body, err = tests.DoGraphQL(s.gql, fmt.Sprintf(`	query { findByIdSaldo(input: { saldo_id: %d }) { status message data { saldo_id total_balance } } }`, s.saldoID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data = tests.GraphQLOpEnvelopeData(body, "findByIdSaldo")
	s.Require().NotNil(data)
	s.Equal(float64(100000), data["total_balance"])

	// Update
	body, err = tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation {
		updateSaldo(input: { saldo_id: %d, card_number: "%s", total_balance: 200000 }) {
			status message
			data { saldo_id total_balance }
		}
	}`, s.saldoID, s.cardNumber))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data = tests.GraphQLOpEnvelopeData(body, "updateSaldo")
	s.Require().NotNil(data)
	s.Equal(float64(200000), data["total_balance"])
}

func (s *SaldoApiTestSuite) Test2_QueryOperations() {
	s.Require().NotZero(s.saldoID)
	s.Require().NotEmpty(s.cardNumber)

	// FindAll
	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`	query { findAllSaldo(input: { page: 1, page_size: 10, search: "%s" }) { status message data { saldo_id total_balance } } }`, s.cardNumber))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	s.Require().NotNil(tests.GraphQLOpData(body, "findAllSaldo"))

	// FindByActive
	body, err = tests.DoGraphQL(s.gql, fmt.Sprintf(`	query { findByActiveSaldo(input: { page: 1, page_size: 10, search: "%s" }) { status message data { saldo_id total_balance } } }`, s.cardNumber))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	s.Require().NotNil(tests.GraphQLOpData(body, "findByActiveSaldo"))

	// FindByCardNumber
	body, err = tests.DoGraphQL(s.gql, fmt.Sprintf(`	query { findByCardNumberSaldo(card_number: "%s") { status message data { saldo_id card_number total_balance } } }`, s.cardNumber))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data := tests.GraphQLOpEnvelopeData(body, "findByCardNumberSaldo")
	s.Require().NotNil(data)
	s.Equal(s.cardNumber, data["card_number"])
}

func (s *SaldoApiTestSuite) Test3_TrashAndRestore() {
	s.Require().NotZero(s.saldoID)

	// Trash
	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`	mutation { trashedSaldo(input: { saldo_id: %d }) { status message data { saldo_id } } }`, s.saldoID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))

	// FindTrashed
	body, err = tests.DoGraphQL(s.gql, fmt.Sprintf(`	query { findByTrashedSaldo(input: { page: 1, page_size: 10, search: "%s" }) { status message data { saldo_id } } }`, s.cardNumber))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	s.Require().NotNil(tests.GraphQLOpData(body, "findByTrashedSaldo"))

	// Restore
	body, err = tests.DoGraphQL(s.gql, fmt.Sprintf(`	mutation { restoreSaldo(input: { saldo_id: %d }) { status message data { saldo_id } } }`, s.saldoID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
}

func TestSaldoApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(SaldoApiTestSuite))
}
