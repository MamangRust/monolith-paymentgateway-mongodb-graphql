package auth_test

import (
	"context"
	"net"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/MamangRust/monolith-payment-gateway-auth/handler"
	"github.com/MamangRust/monolith-payment-gateway-auth/repository"
	"github.com/MamangRust/monolith-payment-gateway-auth/service"
	pb "github.com/MamangRust/monolith-payment-gateway-pb"
	pbrole "github.com/MamangRust/monolith-payment-gateway-pb/role"
	pbuserrole "github.com/MamangRust/monolith-payment-gateway-pb/user_role"
	"github.com/MamangRust/monolith-payment-gateway-pkg/auth"
	"github.com/MamangRust/monolith-payment-gateway-pkg/hash"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	role_handler "github.com/MamangRust/monolith-payment-gateway-role/handler"
	role_repository "github.com/MamangRust/monolith-payment-gateway-role/repository"
	role_service "github.com/MamangRust/monolith-payment-gateway-role/service"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	tests "github.com/MamangRust/monolith-payment-gateway-test"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type AuthHandlerGapiTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	redisClient *redis.Client
	client      pb.AuthServiceClient
	conn        *grpc.ClientConn
	grpcServer  *grpc.Server
	email       string
	password    string
	accessToken string
}

func (s *AuthHandlerGapiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	mongoDB, err := s.ts.MongoDB()
	if err != nil {
		s.Require().NoError(err)
	}

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	// Create role service
	roleRepos := role_repository.NewRepositories(mongoDB)
	roleSvc := role_service.NewService(&role_service.Deps{
		Cache:        cacheStore,
		Repositories: roleRepos,
		Logger:       log,
	})
	roleH := role_handler.NewHandler(roleSvc)

	// Create bufconn-based gRPC server with all services registered BEFORE serving
	lis := bufconn.Listen(1024 * 1024)
	s.grpcServer = grpc.NewServer()

	pbrole.RegisterRoleQueryServiceServer(s.grpcServer, roleH.RoleQuery)
	pbrole.RegisterRoleCommandServiceServer(s.grpcServer, roleH.RoleCommand)
	pbuserrole.RegisterUserRoleServiceServer(s.grpcServer, roleH.UserRole)

	// Connect for role query client
	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)

	roleQueryClient := pbrole.NewRoleQueryServiceClient(conn)

	// Create auth service
	repos := repository.NewRepositories(&repository.RepositoriesDeps{
		DB:              mongoDB,
		RoleQueryClient: roleQueryClient,
		UserRoleClient:  pbuserrole.NewUserRoleServiceClient(conn),
	})

	tokenManager, _ := auth.NewManager("mysecret")
	hasher := hash.NewHashingPassword()

	svc := service.NewService(&service.Deps{
		Repositories: repos,
		Logger:       log,
		Cache:        cacheStore,
		Token:        tokenManager,
		Hash:         hasher,
		Kafka:        nil,
	})

	h := handler.NewAuthHandleGrpc(svc, log)
	pb.RegisterAuthServiceServer(s.grpcServer, h)

	go func() { _ = s.grpcServer.Serve(lis) }()

	s.conn = conn
	s.client = pb.NewAuthServiceClient(conn)

	s.email = "auth.handler.gapi.test@example.com"
	s.password = "password123"

	// Seed ROLE_ADMIN
	now := time.Now()
	_, err = mongoDB.Collection("roles").InsertOne(context.Background(), bson.M{
		"role_id":    1,
		"role_name":  "ROLE_ADMIN",
		"created_at": &now,
		"updated_at": &now,
	})
	s.Require().NoError(err)
}

func (s *AuthHandlerGapiTestSuite) TearDownSuite() {
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

func (s *AuthHandlerGapiTestSuite) Test1_Register() {
	res, err := s.client.RegisterUser(context.Background(), &pb.RegisterRequest{
		Firstname:       "Auth",
		Lastname:        "Gapi",
		Email:           s.email,
		Password:        s.password,
		ConfirmPassword: s.password,
	})
	s.Require().NoError(err)
	s.Require().Equal("success", res.GetStatus())

	data := res.GetData()
	s.Require().NotNil(data)

	// Verify the user so login can proceed
	mongoDB, _ := s.ts.MongoDB()
	_, _ = mongoDB.Collection("users").UpdateOne(context.Background(),
		bson.M{"email": s.email},
		bson.M{"$set": bson.M{"is_verified": true}})
}

func (s *AuthHandlerGapiTestSuite) Test2_Login() {
	res, err := s.client.LoginUser(context.Background(), &pb.LoginRequest{
		Email:    s.email,
		Password: s.password,
	})
	s.Require().NoError(err)
	s.Require().Equal("success", res.GetStatus())

	data := res.GetData()
	s.Require().NotNil(data)
	s.Require().NotEmpty(data.GetAccessToken())
	s.accessToken = data.GetAccessToken()
}

func (s *AuthHandlerGapiTestSuite) Test3_GetMe() {
	res, err := s.client.GetMe(context.Background(), &pb.GetMeRequest{
		UserId: 1,
	})
	s.Require().NoError(err)
	s.Require().Equal("success", res.GetStatus())

	data := res.GetData()
	s.Require().NotNil(data)
}

func (s *AuthHandlerGapiTestSuite) Test4_RefreshToken() {
	loginRes, err := s.client.LoginUser(context.Background(), &pb.LoginRequest{
		Email:    s.email,
		Password: s.password,
	})
	s.Require().NoError(err)
	refreshToken := loginRes.GetData().GetRefreshToken()
	s.Require().NotEmpty(refreshToken)

	res, err := s.client.RefreshToken(context.Background(), &pb.RefreshTokenRequest{
		RefreshToken: refreshToken,
	})
	s.Require().NoError(err)
	s.Require().Equal("success", res.GetStatus())
}

func TestAuthHandlerGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(AuthHandlerGapiTestSuite))
}
