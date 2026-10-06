package user_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	pbrole "github.com/MamangRust/monolith-payment-gateway-pb/role"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/user"
	pbuserrole "github.com/MamangRust/monolith-payment-gateway-pb/user_role"
	"github.com/MamangRust/monolith-payment-gateway-pkg/hash"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/MamangRust/monolith-payment-gateway-user/handler"
	"github.com/MamangRust/monolith-payment-gateway-user/repository"
	"github.com/MamangRust/monolith-payment-gateway-user/service"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserApiTestSuite struct {
	suite.Suite
	ts         *tests.TestSuite
	grpcServer *grpc.Server
	conn       *grpc.ClientConn
	gql        http.Handler
	userID     int
	email      string
}

func (s *UserApiTestSuite) SetupSuite() {
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
	roleConn, err := s.ts.RoleConn(log)
	s.Require().NoError(err)
	_, err = s.ts.SeedRole(s.ts.Ctx, tests.DefaultRoleName)
	s.Require().NoError(err)

	repos := repository.NewRepositories(mongoDB,
		pbrole.NewRoleQueryServiceClient(roleConn),
		pbuserrole.NewUserRoleServiceClient(roleConn),
	)

	hasher := hash.NewHashingPassword()
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(redisClient, log, cacheMetrics)

	svc := service.NewService(&service.Deps{
		Cache:        cacheStore,
		Repositories: repos,
		Hash:         hasher,
		Logger:       log,
	})
	userHandler := handler.NewHandler(svc)

	s.grpcServer = grpc.NewServer()
	pb.RegisterUserQueryServiceServer(s.grpcServer, userHandler)
	pb.RegisterUserCommandServiceServer(s.grpcServer, userHandler)

	addr, err := tests.RunGRPCServer(s.grpcServer)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.gql = tests.NewGraphQLHandler(map[string]*grpc.ClientConn{"user": conn}, redisClient, log)

	s.email = fmt.Sprintf("api.%d@example.com", time.Now().UnixNano())
}

func (s *UserApiTestSuite) TearDownSuite() {
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

func (s *UserApiTestSuite) Test1_UserLifecycle() {
	// Create
	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation {
		createUser(input: { firstname: "JohnApi", lastname: "DoeApi", email: "%s", password: "Password123!", confirm_password: "Password123!" }) {
			status message
			data { id firstname lastname email }
		}
	}`, s.email))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data := tests.GraphQLOpEnvelopeData(body, "createUser")
	s.Require().NotNil(data)
	s.Equal("JohnApi", data["firstname"])
	s.Equal(s.email, data["email"])
	s.userID = int(data["id"].(float64))
	s.Require().NotZero(s.userID)

	// FindById
	body, err = tests.DoGraphQL(s.gql, fmt.Sprintf(`query { findByIdUser(input: { id: %d }) { status message data { id firstname email } } }`, s.userID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data = tests.GraphQLOpEnvelopeData(body, "findByIdUser")
	s.Require().NotNil(data)
	s.Equal(float64(s.userID), data["id"])

	// Update
	body, err = tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation {
		updateUser(input: { id: %d, firstname: "JohnUpdated", lastname: "DoeApi", email: "%s", password: "Password123!", confirm_password: "Password123!" }) {
			status message
			data { id firstname }
		}
	}`, s.userID, s.email))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data = tests.GraphQLOpEnvelopeData(body, "updateUser")
	s.Require().NotNil(data)
	s.Equal("JohnUpdated", data["firstname"])
}

func (s *UserApiTestSuite) Test2_QueryOperations() {
	// FindAll
	body, err := tests.DoGraphQL(s.gql, `query { findAllUser(input: { page: 1, page_size: 10 }) { status message data { id firstname } } }`)
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	s.Require().NotNil(tests.GraphQLOpData(body, "findAllUser"))

	// FindByActive
	body, err = tests.DoGraphQL(s.gql, `query { findByActiveUser(input: { page: 1, page_size: 10 }) { status message data { id firstname } } }`)
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	s.Require().NotNil(tests.GraphQLOpData(body, "findByActiveUser"))
}

func TestUserApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(UserApiTestSuite))
}
