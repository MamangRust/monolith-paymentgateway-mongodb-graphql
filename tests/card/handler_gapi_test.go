package card_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-card/handler"
	"github.com/MamangRust/monolith-payment-gateway-card/repository"
	"github.com/MamangRust/monolith-payment-gateway-card/service"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/card"
	pbuser "github.com/MamangRust/monolith-payment-gateway-pb/user"
	adapter "github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
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
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CardGapiTestSuite struct {
	suite.Suite
	ts     *tests.TestSuite
	cardH  handler.Handler
	userH  user_handler.Handler
	userID int32
	cardID int32
}

func (s *CardGapiTestSuite) SetupSuite() {
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
	repos := repository.NewRepositories(mongoDB, nil)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)

	// CreateUser assigns the default ROLE_ADMIN through the role service.
	roleQueryClient, roleCommandClient, err := s.ts.UserRoleAdapters(log)
	s.Require().NoError(err)
	userRepos := user_repository.NewRepositories(mongoDB, roleQueryClient, roleCommandClient)

	hasher := hash.NewHashingPassword()
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(redisClient, log, cacheMetrics)

	userSvc := user_service.NewService(&user_service.Deps{
		Cache:        cacheStore,
		Repositories: userRepos,
		Hash:         hasher,
		Logger:       log,
	})
	s.userH = user_handler.NewHandler(userSvc)

	// The card service resolves users through its UserAdapter, so serve the
	// in-process user service over bufconn and point the adapter at it.
	userLis := bufconn.Listen(1024 * 1024)
	userServer := grpc.NewServer()
	pbuser.RegisterUserQueryServiceServer(userServer, s.userH)
	go func() { _ = userServer.Serve(userLis) }()
	s.T().Cleanup(userServer.Stop)

	userConn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return userLis.Dial()
		}),
		grpc.WithInsecure())
	s.Require().NoError(err)
	s.T().Cleanup(func() { _ = userConn.Close() })

	cardSvc := service.NewService(&service.Deps{
		Cache:        cacheStore,
		Repositories: repos,
		UserAdapter:  adapter.NewUserAdapter(pbuser.NewUserQueryServiceClient(userConn)),
		Logger:       log,
		Kafka:        nil,
	})
	s.cardH = handler.NewHandler(cardSvc)

	// Create a user for testing
	ctx := context.Background()
	userRes, err := s.userH.Create(ctx, &pbuser.CreateUserRequest{
		Firstname:       "Card",
		Lastname:        "User",
		Email:           fmt.Sprintf("card.user.%d.%d@example.com", time.Now().UnixNano(), time.Now().UnixNano()%10000),
		Password:        "Password123!",
		ConfirmPassword: "Password123!",
	})
	s.Require().NoError(err)
	s.userID = userRes.Data.Id
}

func (s *CardGapiTestSuite) TearDownSuite() {
	if s.ts != nil {
		s.ts.Teardown()
	}
}

func (s *CardGapiTestSuite) Test1_CardLifecycle() {
	ctx := context.Background()

	// Create
	createReq := &pb.CreateCardRequest{
		UserId:       s.userID,
		CardType:     "debit",
		ExpireDate:   timestamppb.New(time.Now().AddDate(5, 0, 0)),
		Cvv:          "123",
		CardProvider: "visa",
	}
	res, err := s.cardH.CreateCard(ctx, createReq)
	s.NoError(err)
	s.NotNil(res)
	s.cardID = res.Data.Id
	s.Equal("debit", res.Data.CardType)

	// FindById
	findReq := &pb.FindByIdCardRequest{CardId: s.cardID}
	resF, err := s.cardH.FindByIdCard(ctx, findReq)
	s.NoError(err)
	s.Equal("debit", resF.Data.CardType)

	// Update
	updateReq := &pb.UpdateCardRequest{
		CardId:       s.cardID,
		UserId:       s.userID,
		CardType:     "credit",
		ExpireDate:   createReq.ExpireDate,
		Cvv:          "456",
		CardProvider: "mastercard",
	}
	resU, err := s.cardH.UpdateCard(ctx, updateReq)
	s.NoError(err)
	s.Equal("credit", resU.Data.CardType)
}

func (s *CardGapiTestSuite) Test2_QueryOperations() {
	ctx := context.Background()
	s.Require().NotZero(s.cardID)

	// FindAll
	allReq := &pb.FindAllCardRequest{Page: 1, PageSize: 10}
	resA, err := s.cardH.FindAllCard(ctx, allReq)
	s.NoError(err)
	s.GreaterOrEqual(resA.PaginationMeta.TotalRecords, int32(1))

	// FindByActive
	resAc, err := s.cardH.FindByActiveCard(ctx, allReq)
	s.NoError(err)
	s.GreaterOrEqual(resAc.PaginationMeta.TotalRecords, int32(1))
}

func (s *CardGapiTestSuite) Test3_TrashAndRestore() {
	ctx := context.Background()
	s.Require().NotZero(s.cardID)

	// Trash
	resT, err := s.cardH.TrashedCard(ctx, &pb.FindByIdCardRequest{CardId: s.cardID})
	s.NoError(err)
	s.NotNil(resT)

	// FindByTrashed
	resTL, err := s.cardH.FindByTrashedCard(ctx, &pb.FindAllCardRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.GreaterOrEqual(resTL.PaginationMeta.TotalRecords, int32(1))

	// Restore
	resR, err := s.cardH.RestoreCard(ctx, &pb.FindByIdCardRequest{CardId: s.cardID})
	s.NoError(err)
	s.NotNil(resR)
}

func (s *CardGapiTestSuite) Test4_BulkOperations() {
	ctx := context.Background()

	// Restore All
	resR, err := s.cardH.RestoreAllCard(ctx, &emptypb.Empty{})
	s.NoError(err)
	s.Equal("success", resR.Status)

	// Delete All Permanent
	resD, err := s.cardH.DeleteAllCardPermanent(ctx, &emptypb.Empty{})
	s.NoError(err)
	s.Equal("success", resD.Status)
}

func TestCardGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(CardGapiTestSuite))
}
