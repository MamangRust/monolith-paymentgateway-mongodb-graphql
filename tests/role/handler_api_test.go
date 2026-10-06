package role_test

import (
	"fmt"
	"net/http"
	"testing"

	pb "github.com/MamangRust/monolith-payment-gateway-pb/role"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-role/handler"
	"github.com/MamangRust/monolith-payment-gateway-role/repository"
	"github.com/MamangRust/monolith-payment-gateway-role/service"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RoleApiTestSuite struct {
	suite.Suite
	ts         *tests.TestSuite
	conn       *grpc.ClientConn
	grpcServer *grpc.Server
	gql        http.Handler
	roleID     int
}

func (s *RoleApiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	redisClient := redis.NewClient(opts)

	mongoDB, err := s.ts.MongoDB()
	s.Require().NoError(err)

	repos := repository.NewRepositories(mongoDB)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(redisClient, log, cacheMetrics)

	svc := service.NewService(&service.Deps{
		Repositories: repos,
		Logger:       log,
		Cache:        cacheStore,
	})

	roleHandler := handler.NewHandler(svc)

	s.grpcServer = grpc.NewServer()
	pb.RegisterRoleQueryServiceServer(s.grpcServer, roleHandler.RoleQuery)
	pb.RegisterRoleCommandServiceServer(s.grpcServer, roleHandler.RoleCommand)

	addr, err := tests.RunGRPCServer(s.grpcServer)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.gql = tests.NewGraphQLHandler(map[string]*grpc.ClientConn{"role": conn}, redisClient, log)
}

func (s *RoleApiTestSuite) TearDownSuite() {
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

func (s *RoleApiTestSuite) Test1_RoleLifecycle() {
	// 1. Create
	body, err := tests.DoGraphQL(s.gql, `mutation {
		createRole(input: { name: "Test API Role" }) {
			status message
			data { id name }
		}
	}`)
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data := tests.GraphQLOpEnvelopeData(body, "createRole")
	s.Require().NotNil(data)
	s.Equal("Test API Role", data["name"])
	s.roleID = int(data["id"].(float64))
	s.Require().NotZero(s.roleID)

	// 2. FindById
	body, err = tests.DoGraphQL(s.gql, fmt.Sprintf(`query { findByIdRole(input: { role_id: %d }) { status message data { id name } } }`, s.roleID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data = tests.GraphQLOpEnvelopeData(body, "findByIdRole")
	s.Require().NotNil(data)
	s.Equal(float64(s.roleID), data["id"])

	// 3. Update
	body, err = tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation {
		updateRole(input: { id: %d, name: "Updated API Role" }) {
			status message
			data { id name }
		}
	}`, s.roleID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data = tests.GraphQLOpEnvelopeData(body, "updateRole")
	s.Require().NotNil(data)
	s.Equal("Updated API Role", data["name"])
}

func (s *RoleApiTestSuite) Test2_QueryOperations() {
	// FindAll
	body, err := tests.DoGraphQL(s.gql, `query { findAllRole(input: { page: 1, page_size: 10 }) { status message data { id name } } }`)
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	s.Require().NotNil(tests.GraphQLOpData(body, "findAllRole"))

	// FindByActive
	body, err = tests.DoGraphQL(s.gql, `query { findByActiveRole(input: { page: 1, page_size: 10 }) { status message data { id name } } }`)
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	s.Require().NotNil(tests.GraphQLOpData(body, "findByActiveRole"))
}

func (s *RoleApiTestSuite) Test3_TrashAndRestore() {
	s.Require().NotZero(s.roleID)

	// Trash
	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation { trashedRole(input: { role_id: %d }) { status message data { id } } }`, s.roleID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))

	// FindByTrashed
	body, err = tests.DoGraphQL(s.gql, `query { findByTrashedRole(input: { page: 1, page_size: 10 }) { status message data { id } } }`)
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	s.Require().NotNil(tests.GraphQLOpData(body, "findByTrashedRole"))

	// Restore
	body, err = tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation { restoreRole(input: { role_id: %d }) { status message data { id } } }`, s.roleID))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
}

func (s *RoleApiTestSuite) Test4_BulkOperations() {
	// RestoreAll
	body, err := tests.DoGraphQL(s.gql, `mutation { restoreAllRole { status message } }`)
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))

	// DeleteAllPermanent
	body, err = tests.DoGraphQL(s.gql, `mutation { deleteAllRolePermanent { status message } }`)
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
}

func TestRoleApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(RoleApiTestSuite))
}
