package user_test

import (
	"context"
	"fmt"
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
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserGapiTestSuite struct {
	suite.Suite
	ts     *tests.TestSuite
	userH  handler.Handler
	userID int32
}

func (s *UserGapiTestSuite) SetupSuite() {
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

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)

	// CreateUser assigns the default ROLE_ADMIN, so the role service must be
	// reachable and the role must exist.
	roleConn, err := s.ts.RoleConn(log)
	s.Require().NoError(err)
	_, err = s.ts.SeedRole(s.ts.Ctx, "ROLE_ADMIN")
	s.Require().NoError(err)

	repos := repository.NewRepositories(mongoDB,
		pbrole.NewRoleQueryServiceClient(roleConn),
		pbuserrole.NewUserRoleServiceClient(roleConn),
	)

	hasher := hash.NewHashingPassword()
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(redisClient, log, cacheMetrics)

	userSvc := service.NewService(&service.Deps{
		Cache:        cacheStore,
		Repositories: repos,
		Hash:         hasher,
		Logger:       log,
	})

	s.userH = handler.NewHandler(userSvc)
}

func (s *UserGapiTestSuite) TearDownSuite() {
	if s.ts != nil {
		s.ts.Teardown()
	}
}

func (s *UserGapiTestSuite) Test1_UserLifecycle() {
	ctx := context.Background()

	// Create
	createReq := &pb.CreateUserRequest{
		Firstname:       "John",
		Lastname:        "Doe",
		Email:           fmt.Sprintf("john.%d@example.com", time.Now().UnixNano()),
		Password:        "Password123!",
		ConfirmPassword: "Password123!",
	}
	res, err := s.userH.Create(ctx, createReq)
	s.NoError(err)
	s.NotNil(res)
	s.userID = res.Data.Id
	s.Equal("John", res.Data.Firstname)

	// FindById
	findReq := &pb.FindByIdUserRequest{Id: s.userID}
	resF, err := s.userH.FindById(ctx, findReq)
	s.NoError(err)
	s.Equal("John", resF.Data.Firstname)

	// Update
	updateReq := &pb.UpdateUserRequest{
		Id:              s.userID,
		Firstname:       "JohnUpdated",
		Lastname:        "Doe",
		Email:           createReq.Email,
		Password:        "Password123!",
		ConfirmPassword: "Password123!",
	}
	resU, err := s.userH.Update(ctx, updateReq)
	s.NoError(err)
	s.Equal("JohnUpdated", resU.Data.Firstname)
}

func (s *UserGapiTestSuite) Test2_QueryOperations() {
	ctx := context.Background()
	s.Require().NotZero(s.userID)

	// FindAll
	allReq := &pb.FindAllUserRequest{Page: 1, PageSize: 10}
	resA, err := s.userH.FindAll(ctx, allReq)
	s.NoError(err)
	s.GreaterOrEqual(resA.PaginationMeta.TotalRecords, int32(1))

	// FindByActive
	active, err := s.userH.FindByActive(ctx, allReq)
	s.NoError(err)
	s.GreaterOrEqual(active.PaginationMeta.TotalRecords, int32(1))
}

func (s *UserGapiTestSuite) Test3_TrashAndRestore() {
	ctx := context.Background()
	s.Require().NotZero(s.userID)

	// Trash
	trashed, err := s.userH.TrashedUser(ctx, &pb.FindByIdUserRequest{Id: s.userID})
	s.NoError(err)
	s.NotNil(trashed)

	// FindByTrashed
	trashedList, err := s.userH.FindByTrashed(ctx, &pb.FindAllUserRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.GreaterOrEqual(trashedList.PaginationMeta.TotalRecords, int32(1))

	// Restore
	restored, err := s.userH.RestoreUser(ctx, &pb.FindByIdUserRequest{Id: s.userID})
	s.NoError(err)
	s.NotNil(restored)
}

func (s *UserGapiTestSuite) Test4_BulkOperations() {
	ctx := context.Background()

	// Restore All
	resR, err := s.userH.RestoreAllUser(ctx, &emptypb.Empty{})
	s.NoError(err)
	s.Equal("success", resR.Status)

	// Delete All Permanent
	resD, err := s.userH.DeleteAllUserPermanent(ctx, &emptypb.Empty{})
	s.NoError(err)
	s.Equal("success", resD.Status)
}

func TestUserGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(UserGapiTestSuite))
}
