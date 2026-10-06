package topup_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	card_handler "github.com/MamangRust/monolith-payment-gateway-card/handler"
	card_repository "github.com/MamangRust/monolith-payment-gateway-card/repository"
	card_service "github.com/MamangRust/monolith-payment-gateway-card/service"
	pbcard "github.com/MamangRust/monolith-payment-gateway-pb/card"
	pbsaldo "github.com/MamangRust/monolith-payment-gateway-pb/saldo"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/topup"
	pbuser "github.com/MamangRust/monolith-payment-gateway-pb/user"
	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	"github.com/MamangRust/monolith-payment-gateway-pkg/hash"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	saldo_handler "github.com/MamangRust/monolith-payment-gateway-saldo/handler"
	saldo_repository "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	saldo_service "github.com/MamangRust/monolith-payment-gateway-saldo/service"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/MamangRust/monolith-payment-gateway-topup/handler"
	"github.com/MamangRust/monolith-payment-gateway-topup/repository"
	"github.com/MamangRust/monolith-payment-gateway-topup/service"
	user_handler "github.com/MamangRust/monolith-payment-gateway-user/handler"
	user_repository "github.com/MamangRust/monolith-payment-gateway-user/repository"
	user_service "github.com/MamangRust/monolith-payment-gateway-user/service"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TopupApiTestSuite struct {
	suite.Suite
	ts         *tests.TestSuite
	grpcServer *grpc.Server
	lis        *bufconn.Listener
	conn       *grpc.ClientConn
	gql        http.Handler
	userID     int
	cardID     int
	cardNumber string
	topupID    int
}

func (s *TopupApiTestSuite) SetupSuite() {
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
	hasher := hash.NewHashingPassword()
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(redisClient, log, cacheMetrics)

	// One gRPC server for card, saldo, user and topup so the topup adapters and
	// the GraphQL gateway can dial it. grpc.NewClient dials lazily, so all
	// services must be registered before the server starts serving.
	s.lis = bufconn.Listen(1024 * 1024)
	s.grpcServer = grpc.NewServer()

	s.conn, err = grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return s.lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)

	cardAdapter := adapter.NewCardAdapter(pbcard.NewCardQueryServiceClient(s.conn), pbcard.NewCardCommandServiceClient(s.conn))
	saldoAdapter := adapter.NewSaldoAdapter(pbsaldo.NewSaldoQueryServiceClient(s.conn), pbsaldo.NewSaldoCommandServiceClient(s.conn))
	userAdapter := adapter.NewUserAdapter(pbuser.NewUserQueryServiceClient(s.conn))

	cardSvc := card_service.NewService(&card_service.Deps{
		Cache:        cacheStore,
		Repositories: card_repository.NewRepositories(mongoDB, nil),
		UserAdapter:  userAdapter,
		Logger:       log,
		Kafka:        nil,
	})
	cardH := card_handler.NewHandler(cardSvc)

	saldoSvc := saldo_service.NewService(&saldo_service.Deps{
		Cache:        cacheStore,
		Repositories: saldo_repository.NewRepositories(mongoDB, nil, nil),
		CardAdapter:  cardAdapter,
		Logger:       log,
		DB:           mongoDB,
	})
	saldoH := saldo_handler.NewHandler(saldoSvc)

	// CreateUser assigns the default ROLE_ADMIN through the role service.
	roleQueryClient, roleCommandClient, err := s.ts.UserRoleAdapters(log)
	s.Require().NoError(err)

	userSvc := user_service.NewService(&user_service.Deps{
		Cache:        cacheStore,
		Repositories: user_repository.NewRepositories(mongoDB, roleQueryClient, roleCommandClient),
		Hash:         hasher,
		Logger:       log,
	})
	userH := user_handler.NewHandler(userSvc)

	// Topup repositories reach card and saldo over the shared gRPC server.
	repos := repository.NewRepositoriesWithAdapters(mongoDB, cardAdapter, saldoAdapter)
	topupSvc := service.NewService(&service.Deps{
		Kafka:        nil,
		Cache:        cacheStore,
		Repositories: repos,
		CardAdapter:  cardAdapter,
		SaldoAdapter: saldoAdapter,
		Logger:       log,
		DB:           mongoDB,
	})
	topupH := handler.NewHandler(topupSvc)

	pbcard.RegisterCardQueryServiceServer(s.grpcServer, cardH)
	pbcard.RegisterCardCommandServiceServer(s.grpcServer, cardH)
	pbsaldo.RegisterSaldoQueryServiceServer(s.grpcServer, saldoH)
	pbsaldo.RegisterSaldoCommandServiceServer(s.grpcServer, saldoH)
	pbuser.RegisterUserQueryServiceServer(s.grpcServer, userH)
	pb.RegisterTopupQueryServiceServer(s.grpcServer, topupH)
	pb.RegisterTopupCommandServiceServer(s.grpcServer, topupH)

	go func() {
		if err := s.grpcServer.Serve(s.lis); err != nil {
		}
	}()

	s.gql = tests.NewGraphQLHandler(map[string]*grpc.ClientConn{"topup": s.conn}, redisClient, log)

	// Create test environment (user + card + saldo)
	ctx := context.Background()
	userRes, err := userH.Create(ctx, &pbuser.CreateUserRequest{
		Firstname:       "Topup",
		Lastname:        "Api",
		Email:           fmt.Sprintf("topup.api.%d@example.com", time.Now().UnixNano()),
		Password:        "Password123!",
		ConfirmPassword: "Password123!",
	})
	s.Require().NoError(err)
	s.userID = int(userRes.Data.Id)

	cardRes, err := cardH.CreateCard(ctx, &pbcard.CreateCardRequest{
		UserId:       userRes.Data.Id,
		CardType:     "debit",
		ExpireDate:   timestamppb.New(time.Now().AddDate(5, 0, 0)),
		Cvv:          "123",
		CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.cardID = int(cardRes.Data.Id)
	s.cardNumber = cardRes.Data.CardNumber

	_, err = saldoH.CreateSaldo(ctx, &pbsaldo.CreateSaldoRequest{
		CardNumber:   s.cardNumber,
		TotalBalance: 100000,
	})
	s.Require().NoError(err)
}

func (s *TopupApiTestSuite) TearDownSuite() {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.lis != nil {
		s.lis.Close()
	}
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
	if s.ts != nil {
		s.ts.Teardown()
	}
}

func (s *TopupApiTestSuite) Test1_CreateTopup() {
	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation {
		createTopup(input: { card_number: "%s", topup_no: "", topup_amount: 50000, topup_method: "bca" }) {
			status message
			data { id card_number topup_amount topup_method }
		}
	}`, s.cardNumber))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data := tests.GraphQLOpEnvelopeData(body, "createTopup")
	s.Require().NotNil(data)
	s.Equal(float64(50000), data["topup_amount"])
	s.topupID = int(data["id"].(float64))
	s.Require().NotZero(s.topupID)
}

func (s *TopupApiTestSuite) Test2_UpdateTopup() {
	s.Require().NotZero(s.topupID)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation {
		updateTopup(input: { topup_id: %d, card_number: "%s", topup_amount: 75000, topup_method: "mandiri" }) {
			status message
			data { id topup_amount topup_method }
		}
	}`, s.topupID, s.cardNumber))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data := tests.GraphQLOpEnvelopeData(body, "updateTopup")
	s.Require().NotNil(data)
	s.Equal(float64(75000), data["topup_amount"])
}

func (s *TopupApiTestSuite) Test3_FindOperations() {
	s.Require().NotZero(s.topupID)

	// FindById
	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`query { findByIdTopup(input: { topup_id: %d }) { status message data { id topup_amount } } }`, s.topupID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data := tests.GraphQLOpEnvelopeData(body, "findByIdTopup")
	s.Require().NotNil(data)
	s.Equal(float64(s.topupID), data["id"])

	// FindAll
	body, err = tests.DoGraphQL(s.gql, `query { findAllTopup(input: { page: 1, page_size: 10 }) { status message data { id topup_amount } } }`)
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	s.Require().NotNil(tests.GraphQLOpData(body, "findAllTopup"))
}

func (s *TopupApiTestSuite) Test4_TrashAndRestore() {
	s.Require().NotZero(s.topupID)

	// Trash
	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation { trashedTopup(input: { topup_id: %d }) { status message data { id } } }`, s.topupID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))

	// Restore
	body, err = tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation { restoreTopup(input: { topup_id: %d }) { status message data { id } } }`, s.topupID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
}

func TestTopupApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TopupApiTestSuite))
}
