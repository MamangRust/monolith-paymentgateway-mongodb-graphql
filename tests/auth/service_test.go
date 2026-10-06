package auth_test

import (
	"context"
	"net"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/MamangRust/monolith-payment-gateway-auth/repository"
	"github.com/MamangRust/monolith-payment-gateway-auth/service"
	pbrole "github.com/MamangRust/monolith-payment-gateway-pb/role"
	pbuserrole "github.com/MamangRust/monolith-payment-gateway-pb/user_role"
	"github.com/MamangRust/monolith-payment-gateway-pkg/auth"
	"github.com/MamangRust/monolith-payment-gateway-pkg/hash"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	role_handler "github.com/MamangRust/monolith-payment-gateway-role/handler"
	role_repository "github.com/MamangRust/monolith-payment-gateway-role/repository"
	role_service "github.com/MamangRust/monolith-payment-gateway-role/service"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	tests "github.com/MamangRust/monolith-payment-gateway-test"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type AuthServiceTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	redisClient *redis.Client
	service     *service.Service
	grpcServer  *grpc.Server
	email       string
	password    string
}

func (s *AuthServiceTestSuite) SetupSuite() {
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

	// Create role service via bufconn
	roleRepos := role_repository.NewRepositories(mongoDB)
	roleSvc := role_service.NewService(&role_service.Deps{
		Cache:        cacheStore,
		Repositories: roleRepos,
		Logger:       log,
	})
	roleH := role_handler.NewHandler(roleSvc)

	lis := bufconn.Listen(1024 * 1024)
	s.grpcServer = grpc.NewServer()
	pbrole.RegisterRoleQueryServiceServer(s.grpcServer, roleH.RoleQuery)
	pbrole.RegisterRoleCommandServiceServer(s.grpcServer, roleH.RoleCommand)
	pbuserrole.RegisterUserRoleServiceServer(s.grpcServer, roleH.UserRole)
	go func() { _ = s.grpcServer.Serve(lis) }()

	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)

	roleQueryClient := pbrole.NewRoleQueryServiceClient(conn)

	repos := repository.NewRepositories(&repository.RepositoriesDeps{
		DB:              mongoDB,
		RoleQueryClient: roleQueryClient,
		UserRoleClient:  pbuserrole.NewUserRoleServiceClient(conn),
	})

	tokenManager, _ := auth.NewManager("mysecret")
	hasher := hash.NewHashingPassword()

	s.service = service.NewService(&service.Deps{
		Repositories: repos,
		Logger:       log,
		Cache:        cacheStore,
		Token:        tokenManager,
		Hash:         hasher,
		Kafka:        nil,
	})

	s.email = "auth.service.test@example.com"
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

func (s *AuthServiceTestSuite) TearDownSuite() {
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
	s.redisClient.Close()
	s.ts.Teardown()
}

func (s *AuthServiceTestSuite) Test1_Register() {
	ctx := context.Background()
	req := &requests.RegisterRequest{
		FirstName:       "Auth",
		LastName:        "Service",
		Email:           s.email,
		Password:        s.password,
		ConfirmPassword: s.password,
	}

	res, err := s.service.Register.Register(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(s.email, res.Email)

	// Verify the user so login can proceed
	mongoDB, _ := s.ts.MongoDB()
	_, _ = mongoDB.Collection("users").UpdateOne(ctx,
		bson.M{"email": s.email},
		bson.M{"$set": bson.M{"is_verified": true}})
}

func (s *AuthServiceTestSuite) Test2_Login() {
	ctx := context.Background()
	req := &requests.AuthRequest{
		Email:    s.email,
		Password: s.password,
	}

	res, err := s.service.Login.Login(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.NotEmpty(res.AccessToken)
	s.NotEmpty(res.RefreshToken)
}

func (s *AuthServiceTestSuite) Test3_GetMe() {
	ctx := context.Background()

	// First login to get a user ID
	loginRes, err := s.service.Login.Login(ctx, &requests.AuthRequest{
		Email:    s.email,
		Password: s.password,
	})
	s.NoError(err)
	s.NotNil(loginRes)

	// Get user from DB to get the user ID
	mongoDB, _ := s.ts.MongoDB()
	var user struct {
		UserID int32 `bson:"user_id"`
	}
	_ = mongoDB.Collection("users").FindOne(ctx, bson.M{"email": s.email}).Decode(&user)

	res, err := s.service.Identify.GetMe(ctx, int(user.UserID))
	s.NoError(err)
	s.NotNil(res)
	s.Equal(s.email, res.Email)
}

func TestAuthServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(AuthServiceTestSuite))
}
