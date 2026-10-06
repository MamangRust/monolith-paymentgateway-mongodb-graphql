package auth_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-auth/handler"
	"github.com/MamangRust/monolith-payment-gateway-auth/repository"
	"github.com/MamangRust/monolith-payment-gateway-auth/service"
	pb "github.com/MamangRust/monolith-payment-gateway-pb"
	pbrole "github.com/MamangRust/monolith-payment-gateway-pb/role"
	pbuserrole "github.com/MamangRust/monolith-payment-gateway-pb/user_role"
	"github.com/MamangRust/monolith-payment-gateway-pkg/auth"
	"github.com/MamangRust/monolith-payment-gateway-pkg/hash"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthHandlerApiTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	conn        *grpc.ClientConn
	redisClient *redis.Client
	grpcServer  *grpc.Server
	gql         http.Handler
	gqlAuth     http.Handler
	email       string
	password    string
	accessToken string
	userID      int
}

func (s *AuthHandlerApiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	mongoDB, err := s.ts.MongoDB()
	s.Require().NoError(err)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	// The auth service reaches roles over gRPC, so start the role service
	// in-process and seed ROLE_ADMIN (register assigns it to new users).
	roleConn, err := s.ts.RoleConn(log)
	s.Require().NoError(err)
	_, err = s.ts.SeedRole(s.ts.Ctx, tests.DefaultRoleName)
	s.Require().NoError(err)

	repos := repository.NewRepositories(&repository.RepositoriesDeps{
		DB:              mongoDB,
		RoleQueryClient: pbrole.NewRoleQueryServiceClient(roleConn),
		UserRoleClient:  pbuserrole.NewUserRoleServiceClient(roleConn),
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

	s.grpcServer = grpc.NewServer()
	pb.RegisterAuthServiceServer(s.grpcServer, h)

	addr, err := tests.RunGRPCServer(s.grpcServer)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.gql = tests.NewGraphQLHandler(map[string]*grpc.ClientConn{"auth": conn}, s.redisClient, log)
	s.gqlAuth = tests.NewAuthenticatedGraphQLHandler(map[string]*grpc.ClientConn{"auth": conn}, s.redisClient, log, tokenManager)

	s.email = fmt.Sprintf("auth.api.%d@example.com", time.Now().UnixNano())
	s.password = "password123"
}

func (s *AuthHandlerApiTestSuite) TearDownSuite() {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
	if s.redisClient != nil {
		s.redisClient.Close()
	}
	if s.ts != nil {
		s.ts.Teardown()
	}
}

func (s *AuthHandlerApiTestSuite) Test1_Register() {
	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation {
		registerUser(input: { firstname: "Auth", lastname: "API", email: "%s", password: "%s", confirm_password: "%s" }) {
			status message
			data { id firstname lastname email }
		}
	}`, s.email, s.password, s.password))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data := tests.GraphQLOpEnvelopeData(body, "registerUser")
	s.Require().NotNil(data)
	s.userID = int(data["id"].(float64))
	s.Require().NotZero(s.userID)
	s.Equal(s.email, data["email"])

	// Login requires a verified account — complete verification via the code
	// Register cached in Redis (mirrors the real email-verification flow).
	s.verifyUser(s.email)
}

func (s *AuthHandlerApiTestSuite) Test2_Login() {
	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation {
		loginUser(input: { email: "%s", password: "%s" }) {
			status message
			data { access_token refresh_token }
		}
	}`, s.email, s.password))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data := tests.GraphQLOpEnvelopeData(body, "loginUser")
	s.Require().NotNil(data)
	s.accessToken = data["access_token"].(string)
	s.Require().NotEmpty(s.accessToken)
}

func (s *AuthHandlerApiTestSuite) Test3_GetMe() {
	s.Require().NotEmpty(s.accessToken)

	body, err := tests.DoGraphQLWithAuth(s.gqlAuth, fmt.Sprintf(`query {
		getMe(input: { access_token: "%s" }) {
			status message
			data { id firstname lastname email }
		}
	}`, s.accessToken), s.accessToken)
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
	data := tests.GraphQLOpEnvelopeData(body, "getMe")
	s.Require().NotNil(data)
	s.Equal(s.email, data["email"])
	s.Equal(float64(s.userID), data["id"])
}

func (s *AuthHandlerApiTestSuite) Test4_LoginLockout() {
	email := fmt.Sprintf("locked.api.%d@example.com", time.Now().UnixNano())
	password := "wrongpassword"

	// Register user first
	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation {
		registerUser(input: { firstname: "Locked", lastname: "API", email: "%s", password: "correctpassword", confirm_password: "correctpassword" }) {
			status message
			data { id email }
		}
	}`, email))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))

	s.verifyUser(email)

	loginQuery := fmt.Sprintf(`mutation {
		loginUser(input: { email: "%s", password: "%s" }) {
			status message
			data { access_token }
		}
	}`, email, password)

	// Fail login 5 times
	for i := 0; i < 5; i++ {
		body, err = tests.DoGraphQL(s.gql, loginQuery)
		s.Require().NoError(err)
		s.NotEmpty(tests.GraphQLErrorMessages(body), "expected login attempt %d to fail", i+1)
	}

	// 6th attempt should be rejected with the account-locked error
	body, err = tests.DoGraphQL(s.gql, loginQuery)
	s.Require().NoError(err)
	messages := tests.GraphQLErrorMessages(body)
	s.NotEmpty(messages, "expected 6th login attempt to be rejected")
	s.Contains(messages[0], "locked")
}

// verifyUser marks a just-registered user as verified via verifyCode using the
// code Register cached in Redis, mirroring the real email-verification flow.
//
// The MongoDB register flow only caches the code in Redis; it does not persist
// it on the user document, while VerifyCode looks the user up by the Mongo
// `verification_code` field. Persist the cached code first so the real
// VerifyCode mutation can run.
func (s *AuthHandlerApiTestSuite) verifyUser(email string) {
	ctx := context.Background()
	raw, err := s.redisClient.Get(ctx, "register:verify_code:"+email).Result()
	s.Require().NoError(err, "verification code should be cached after register")
	var code string
	s.Require().NoError(json.Unmarshal([]byte(raw), &code), "cached code is JSON-encoded")

	mongoDB, err := s.ts.MongoDB()
	s.Require().NoError(err)
	_, err = mongoDB.Collection("users").UpdateOne(ctx,
		bson.M{"email": email},
		bson.M{"$set": bson.M{"verification_code": code}},
	)
	s.Require().NoError(err)

	body, err := tests.DoGraphQL(s.gql, fmt.Sprintf(`mutation { verifyCode(input: { code: "%s" }) { status message } }`, code))
	s.Require().NoError(err)
	s.Require().Empty(tests.GraphQLErrorMessages(body), tests.GraphQLErrorMessages(body))
}

func TestAuthHandlerApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(AuthHandlerApiTestSuite))
}
