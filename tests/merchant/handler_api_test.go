package merchant_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-merchant/handler"
	"github.com/MamangRust/monolith-payment-gateway-merchant/repository"
	"github.com/MamangRust/monolith-payment-gateway-merchant/service"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/merchant"
	pbuser "github.com/MamangRust/monolith-payment-gateway-pb/user"
	"github.com/MamangRust/monolith-payment-gateway-pkg/hash"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	user_handler "github.com/MamangRust/monolith-payment-gateway-user/handler"
	user_repository "github.com/MamangRust/monolith-payment-gateway-user/repository"
	user_service "github.com/MamangRust/monolith-payment-gateway-user/service"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MerchantApiTestSuite struct {
	suite.Suite
	harness    *tests.DependencyHarness
	ts         *tests.TestSuite
	grpcServer *grpc.Server
	conn       *grpc.ClientConn
	gql        http.Handler
	userID     int
	merchantID int
	apiKey     string
}

func (s *MerchantApiTestSuite) SetupSuite() {
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

	// CreateUser assigns the default ROLE_ADMIN through the role service, so
	// the role service must be reachable and the role must exist.
	roleQueryClient, roleCommandClient, err := s.ts.UserRoleAdapters(log)
	s.Require().NoError(err)

	userRepos := user_repository.NewRepositories(mongoDB, roleQueryClient, roleCommandClient)
	hasher := hash.NewHashingPassword()
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(redisClient, log, cacheMetrics)

	// The merchant service resolves users through the in-process dependency
	// harness; its user service is backed by the same MongoDB, so users
	// created below are visible to it.
	harness, err := tests.NewDependencyHarness(mongoDB, cacheStore, log)
	s.Require().NoError(err)
	s.harness = harness

	repos := repository.NewRepositories(mongoDB, nil)

	merchantSvc := service.NewService(&service.Deps{
		Cache:        cacheStore,
		DB:           mongoDB,
		Repositories: repos,
		UserAdapter:  harness.UserAdapter,
		Logger:       log,
		Kafka:        nil,
	})

	userSvc := user_service.NewService(&user_service.Deps{
		Cache:        cacheStore,
		Repositories: userRepos,
		Hash:         hasher,
		Logger:       log,
	})

	merchantH := handler.NewHandler(merchantSvc)
	userH := user_handler.NewHandler(userSvc)

	// Create a user to own the merchant
	ctx := context.Background()
	userRes, err := userH.Create(ctx, &pbuser.CreateUserRequest{
		Firstname:       "Merchant",
		Lastname:        "Api",
		Email:           fmt.Sprintf("merchant.api.%d@example.com", time.Now().UnixNano()),
		Password:        "Password123!",
		ConfirmPassword: "Password123!",
	})
	s.Require().NoError(err)
	s.userID = int(userRes.Data.Id)

	// Start the merchant gRPC service
	s.grpcServer = grpc.NewServer()
	pb.RegisterMerchantQueryServiceServer(s.grpcServer, merchantH)
	pb.RegisterMerchantCommandServiceServer(s.grpcServer, merchantH)

	addr, err := tests.RunGRPCServer(s.grpcServer)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.gql = tests.NewGraphQLHandler(map[string]*grpc.ClientConn{"merchant": conn}, redisClient, log)
}

func (s *MerchantApiTestSuite) TearDownSuite() {
	if s.harness != nil {
		s.harness.Close()
	}
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

func (s *MerchantApiTestSuite) Test1_MerchantLifecycle() {
	// Create
	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation {
		createMerchant(input: { name: "Api Merchant", user_id: %d }) {
			status message
			data { id name api_key status user_id }
		}
	}`, s.userID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data := tests.GraphQLOpEnvelopeData(body, "createMerchant")
	s.Require().NotNil(data)
	s.Equal("Api Merchant", data["name"])
	s.Equal(float64(s.userID), data["user_id"])
	s.merchantID = int(data["id"].(float64))
	s.apiKey = data["api_key"].(string)
	s.Require().NotZero(s.merchantID)
	s.Require().NotEmpty(s.apiKey)

	// FindById
	body, err = tests.DoGraphQL(s.gql, fmt.Sprintf(`query { findByIdMerchant(input: { merchant_id: %d }) { status message data { id name } } }`, s.merchantID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data = tests.GraphQLOpEnvelopeData(body, "findByIdMerchant")
	s.Require().NotNil(data)
	s.Equal(float64(s.merchantID), data["id"])
	s.Equal("Api Merchant", data["name"])

	// Update
	body, err = tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation {
		updateMerchant(input: { merchant_id: %d, name: "Updated Api Merchant", user_id: %d, status: "active" }) {
			status message
			data { id name }
		}
	}`, s.merchantID, s.userID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data = tests.GraphQLOpEnvelopeData(body, "updateMerchant")
	s.Require().NotNil(data)
	s.Equal("Updated Api Merchant", data["name"])
}

func (s *MerchantApiTestSuite) Test2_QueryOperations() {
	s.Require().NotZero(s.merchantID)
	s.Require().NotEmpty(s.apiKey)

	// FindAll
	body, err := tests.DoGraphQL(s.gql, `query { findAllMerchant(input: { page: 1, page_size: 10 }) { status message data { id name } } }`)
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	s.Require().NotNil(tests.GraphQLOpData(body, "findAllMerchant"))

	// FindByActive
	body, err = tests.DoGraphQL(s.gql, `query { findByActive(input: { page: 1, page_size: 10 }) { status message data { id name } } }`)
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	s.Require().NotNil(tests.GraphQLOpData(body, "findByActive"))

	// FindByApiKey
	body, err = tests.DoGraphQL(s.gql, fmt.Sprintf(`query { findByApiKey(input: { api_key: "%s" }) { status message data { id api_key } } }`, s.apiKey))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data := tests.GraphQLOpEnvelopeData(body, "findByApiKey")
	s.Require().NotNil(data)
	s.Equal(s.apiKey, data["api_key"])

	// FindByMerchantUserId
	body, err = tests.DoGraphQL(s.gql, fmt.Sprintf(`query { findByMerchantUserId(input: { user_id: %d }) { status message data { id } } }`, s.userID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	s.Require().NotNil(tests.GraphQLOpData(body, "findByMerchantUserId"))
}

func (s *MerchantApiTestSuite) Test3_TrashAndRestore() {
	s.Require().NotZero(s.merchantID)

	// Trash
	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation { trashedMerchant(input: { merchant_id: %d }) { status message data { id } } }`, s.merchantID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))

	// FindTrashed
	body, err = tests.DoGraphQL(s.gql, `query { findByTrashed(input: { page: 1, page_size: 10 }) { status message data { id } } }`)
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	s.Require().NotNil(tests.GraphQLOpData(body, "findByTrashed"))

	// Restore
	body, err = tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation { restoreMerchant(input: { merchant_id: %d }) { status message data { id } } }`, s.merchantID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
}

func TestMerchantApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(MerchantApiTestSuite))
}
