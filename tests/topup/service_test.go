package topup_test

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"sync"
	"testing"
	"time"

	card_repo_impl "github.com/MamangRust/monolith-payment-gateway-card/repository"
	models "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	saldo_repo_impl "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/MamangRust/monolith-payment-gateway-topup/repository"
	"github.com/MamangRust/monolith-payment-gateway-topup/service"
	user_repo "github.com/MamangRust/monolith-payment-gateway-user/repository"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

type testCardRepo struct {
	db *mongo.Database
}

func (r *testCardRepo) FindUserCardByCardNumber(ctx context.Context, card_number string) (*models.CardByEmailRow, error) {
	var card models.Card
	err := r.db.Collection("cards").FindOne(ctx, bson.M{"card_number": card_number}).Decode(&card)
	if err != nil {
		return nil, err
	}
	var user models.User
	_ = r.db.Collection("users").FindOne(ctx, bson.M{"user_id": card.UserID}).Decode(&user)
	return &models.CardByEmailRow{
		CardNumber: card.CardNumber, Email: user.Email,
	}, nil
}

func (r *testCardRepo) FindCardByCardNumber(ctx context.Context, card_number string) (*models.CardAllFieldsRow, error) {
	var result models.Card
	err := r.db.Collection("cards").FindOne(ctx, bson.M{"card_number": card_number}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &models.CardAllFieldsRow{
		CardID: result.CardID, UserID: result.UserID, CardNumber: result.CardNumber,
	}, nil
}

func (r *testCardRepo) UpdateCard(ctx context.Context, request *requests.UpdateCardRequest) (*models.CardUpdateRow, error) {
	_, err := r.db.Collection("cards").UpdateOne(ctx, bson.M{"card_id": request.CardID}, bson.M{"$set": bson.M{
		"card_type":     request.CardType,
		"expire_date":   request.ExpireDate,
		"cvv":           request.CVV,
		"card_provider": request.CardProvider,
	}})
	if err != nil {
		return nil, err
	}
	var result models.Card
	err = r.db.Collection("cards").FindOne(ctx, bson.M{"card_id": request.CardID}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &models.CardUpdateRow{
		CardID: result.CardID, CardNumber: result.CardNumber, CardType: result.CardType,
	}, nil
}

type testSaldoRepo struct {
	db *mongo.Database
}

func (r *testSaldoRepo) FindByCardNumber(ctx context.Context, card_number string) (*models.Saldo, error) {
	var result models.Saldo
	err := r.db.Collection("saldos").FindOne(ctx, bson.M{"card_number": card_number}).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("saldo not found for card %s: %w", card_number, err)
	}
	return &result, nil
}

func (r *testSaldoRepo) UpdateSaldoBalance(ctx context.Context, request *requests.UpdateSaldoBalance) (*models.UpdateSaldoBalanceRow, error) {
	_, err := r.db.Collection("saldos").UpdateOne(ctx, bson.M{"card_number": request.CardNumber}, bson.M{"$set": bson.M{"total_balance": request.TotalBalance}})
	if err != nil {
		return nil, err
	}
	var result models.Saldo
	err = r.db.Collection("saldos").FindOne(ctx, bson.M{"card_number": request.CardNumber}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &models.UpdateSaldoBalanceRow{
		SaldoID: result.SaldoID, CardNumber: result.CardNumber, TotalBalance: int32(result.TotalBalance),
	}, nil
}

type TopupServiceTestSuite struct {
	suite.Suite
	ts           *tests.TestSuite
	topupService service.Service
	mongoDB      *mongo.Database
	topupID      int
	cardNumber   string
	userID       int
	harness      *tests.DependencyHarness
}

func (s *TopupServiceTestSuite) SetupSuite() {
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
	s.mongoDB = mongoDB

	cardRepo := &testCardRepo{db: mongoDB}
	saldoRepo := &testSaldoRepo{db: mongoDB}
	repos := repository.NewRepositoriesWithAdapters(mongoDB, cardRepo, saldoRepo)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(redisClient, log, cacheMetrics)

	harness, err := tests.NewDependencyHarness(mongoDB, cacheStore, log)
	s.Require().NoError(err)
	s.harness = harness

	s.topupService = service.NewService(&service.Deps{
		Kafka:        nil,
		Repositories: repos,
		CardAdapter:  harness.CardAdapter,
		SaldoAdapter: harness.SaldoAdapter,
		Logger:       log,
		Cache:        cacheStore,
		DB:           mongoDB,
	})

	// Seed User
	userRepo := user_repo.NewUserCommandRepository(mongoDB.Collection("users"))
	user, err := userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Topup",
		LastName:  "Tester",
		Email:     fmt.Sprintf("topup.tester.%d@example.com", time.Now().UnixNano()),
		Password:  "password123",
	})
	s.Require().NoError(err)
	s.userID = int(user.UserID)

	// Seed Card
	cardCmdRepo := card_repo_impl.NewCardCommandRepository(mongoDB.Collection("cards"))
	card, err := cardCmdRepo.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID:       s.userID,
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(5, 0, 0),
		CVV:          "123",
		CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.cardNumber = card.CardNumber

	// Seed Saldo
	saldoCmdRepo := saldo_repo_impl.NewSaldoCommandRepository(mongoDB.Collection("saldos"))
	_, err = saldoCmdRepo.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{
		CardNumber:   s.cardNumber,
		TotalBalance: 1000000,
	})
	s.Require().NoError(err)
}

func (s *TopupServiceTestSuite) TearDownSuite() {
	if s.harness != nil {
		s.harness.Close()
	}
	s.ts.Teardown()
}

// helper
func (s *TopupServiceTestSuite) getSaldoBalance(ctx context.Context, cardNumber string) int32 {
	var saldo models.Saldo
	err := s.mongoDB.Collection("saldos").FindOne(ctx, bson.M{"card_number": cardNumber}).Decode(&saldo)
	s.Require().NoError(err)
	return int32(saldo.TotalBalance)
}

// helper
func (s *TopupServiceTestSuite) getTopupAmount(ctx context.Context, topupID int) int32 {
	var t models.Topup
	err := s.mongoDB.Collection("topups").FindOne(ctx, bson.M{"topup_id": topupID}).Decode(&t)
	s.Require().NoError(err)
	return int32(t.TopupAmount)
}

func (s *TopupServiceTestSuite) Test1_TopupLifecycle() {
	ctx := context.Background()

	// Create Topup
	createReq := &requests.CreateTopupRequest{
		CardNumber:  s.cardNumber,
		TopupAmount: 100000,
		TopupMethod: "gopay",
	}
	res, err := s.topupService.CreateTopup(ctx, createReq)
	s.NoError(err)
	s.NotNil(res)
	s.topupID = int(res.TopupID)
	s.Equal("success", res.Status)

	// FindById
	found, err := s.topupService.FindById(ctx, s.topupID)
	s.NoError(err)
	s.NotNil(found)
	s.Equal(int32(s.topupID), found.TopupID)

	// Update Topup
	updateReq := &requests.UpdateTopupRequest{
		TopupID:     &s.topupID,
		CardNumber:  s.cardNumber,
		TopupAmount: 200000,
		TopupMethod: "dana",
	}
	updated, err := s.topupService.UpdateTopup(ctx, updateReq)
	s.NoError(err)
	s.NotNil(updated)
	s.Equal("success", updated.Status)
}

func (s *TopupServiceTestSuite) Test2_QueryOperations() {
	ctx := context.Background()

	// FindAll
	all, total, err := s.topupService.FindAll(ctx, &requests.FindAllTopups{
		Page:     1,
		PageSize: 10,
	})
	s.NoError(err)
	s.NotNil(all)
	s.GreaterOrEqual(*total, 1)

	// FindByActive
	active, totalActive, err := s.topupService.FindByActive(ctx, &requests.FindAllTopups{
		Page:     1,
		PageSize: 10,
	})
	s.NoError(err)
	s.NotNil(active)
	s.GreaterOrEqual(*totalActive, 1)

	// FindByCardNumber
	byCard, _, err := s.topupService.FindAllByCardNumber(ctx, &requests.FindAllTopupsByCardNumber{
		CardNumber: s.cardNumber,
		Page:       1,
		PageSize:   10,
	})
	s.NoError(err)
	s.NotNil(byCard)
	s.GreaterOrEqual(len(byCard), 1)
}

func (s *TopupServiceTestSuite) Test3_TrashAndRestore() {
	ctx := context.Background()
	s.Require().NotZero(s.topupID)

	// Trash
	trashed, err := s.topupService.TrashedTopup(ctx, s.topupID)
	s.NoError(err)
	s.NotNil(trashed)

	// FindByTrashed
	trashedList, totalTrashed, err := s.topupService.FindByTrashed(ctx, &requests.FindAllTopups{
		Page:     1,
		PageSize: 10,
	})
	s.NoError(err)
	s.NotNil(trashedList)
	s.GreaterOrEqual(*totalTrashed, 1)

	// Restore
	restored, err := s.topupService.RestoreTopup(ctx, s.topupID)
	s.NoError(err)
	s.NotNil(restored)
}

func (s *TopupServiceTestSuite) Test4_BulkOperations() {
	ctx := context.Background()

	// Restore All
	ok, err := s.topupService.RestoreAllTopup(ctx)
	s.NoError(err)
	s.True(ok)

	// Delete All Permanent
	ok, err = s.topupService.DeleteAllTopupPermanent(ctx)
	s.NoError(err)
	s.True(ok)
}

func (s *TopupServiceTestSuite) Test5_TopupIdempotentReplay() {
	ctx := context.Background()
	key := fmt.Sprintf("topup:create:user:%d:op:%d", s.userID, time.Now().UnixNano())

	req := &requests.CreateTopupRequest{
		CardNumber:     s.cardNumber,
		TopupAmount:    150000,
		TopupMethod:    "gopay",
		IdempotencyKey: key,
	}

	first, err := s.topupService.CreateTopup(ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(first)
	s.Equal("success", first.Status)

	balanceAfterCreate := s.getSaldoBalance(ctx, s.cardNumber)

	// Replay with the same key: the original record is returned and the
	// balance must not be credited a second time.
	second, err := s.topupService.CreateTopup(ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(second)
	s.Equal(first.TopupID, second.TopupID, "replay must return the original record")

	balanceAfterReplay := s.getSaldoBalance(ctx, s.cardNumber)
	s.Equal(balanceAfterCreate, balanceAfterReplay, "replay must not credit the balance twice")
}

func (s *TopupServiceTestSuite) Test6_TopupConcurrentIdempotentReplay() {
	ctx := context.Background()
	key := fmt.Sprintf("topup:create:user:%d:concurrent:%d", s.userID, time.Now().UnixNano())
	amount := 120000

	before := s.getSaldoBalance(ctx, s.cardNumber)

	// Two concurrent creates with the SAME idempotency key: both must succeed
	// (the loser hits the unique index and falls back to returning the winner's
	// record — replay, not 409) and only one record/credit may exist.
	requestsReady := make(chan struct{})
	results := make(chan error, 2)
	ids := make(chan int32, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func() {
			defer wg.Done()
			<-requestsReady
			res, err := s.topupService.CreateTopup(ctx, &requests.CreateTopupRequest{
				CardNumber:     s.cardNumber,
				TopupAmount:    amount,
				TopupMethod:    "gopay",
				IdempotencyKey: key,
			})
			results <- err
			if err == nil {
				ids <- res.TopupID
			}
		}()
	}
	close(requestsReady)
	wg.Wait()
	close(results)
	close(ids)

	successes := 0
	for err := range results {
		if err == nil {
			successes++
		}
	}
	s.Equal(2, successes, "concurrent duplicate with the same key must both succeed (replay, not 409)")

	var firstID int32
	first := true
	for id := range ids {
		if first {
			firstID = id
			first = false
		} else {
			s.Equal(firstID, id, "both concurrent requests must return the same record id")
		}
	}

	rowCount, err := s.mongoDB.Collection("topups").CountDocuments(ctx, bson.M{"idempotency_key": key})
	s.Require().NoError(err)
	s.Equal(int64(1), rowCount, "exactly one record may exist for the key")

	after := s.getSaldoBalance(ctx, s.cardNumber)
	s.Equal(before+int32(amount), after, "balance must be credited exactly once")
}

func (s *TopupServiceTestSuite) Test7_TopupConcurrentUpdatesApplyAllDeltas() {
	ctx := context.Background()

	// Seed a topup to update concurrently.
	key := fmt.Sprintf("topup:create:user:%d:update:%d", s.userID, time.Now().UnixNano())
	res, err := s.topupService.CreateTopup(ctx, &requests.CreateTopupRequest{
		CardNumber:     s.cardNumber,
		TopupAmount:    100000,
		TopupMethod:    "gopay",
		IdempotencyKey: key,
	})
	s.Require().NoError(err)
	topupID := int(res.TopupID)

	before := s.getSaldoBalance(ctx, s.cardNumber)

	// Two concurrent updates with different target amounts.
	requestsReady := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	for _, newAmount := range []int{150000, 200000} {
		newAmount := newAmount
		go func() {
			defer wg.Done()
			<-requestsReady
			_, err := s.topupService.UpdateTopup(ctx, &requests.UpdateTopupRequest{
				TopupID:     &topupID,
				CardNumber:  s.cardNumber,
				TopupAmount: newAmount,
				TopupMethod: "gopay",
			})
			results <- err
		}()
	}
	close(requestsReady)
	wg.Wait()
	close(results)

	for err := range results {
		s.NoError(err, "concurrent updates must not fail")
	}

	finalAmount := s.getTopupAmount(ctx, topupID)
	after := s.getSaldoBalance(ctx, s.cardNumber)
	expected := before - 100000 + finalAmount
	s.Equal(expected, after, "balance must be consistent with the final topup amount")
}

func (s *TopupServiceTestSuite) TestZZ_NotFound404() {
	ctx := context.Background()
	_, err := s.topupService.FindById(ctx, 999999999)
	s.Require().Error(err)
	s.Contains(err.Error(), "not found")
}

func (s *TopupServiceTestSuite) Test8_FindByIdNotFound() {
	ctx := context.Background()
	_, err := s.topupService.FindById(ctx, 999999999)
	s.Require().Error(err, "a non-existent topup must not resolve")
	s.Contains(err.Error(), "not found", "missing topup must surface as a 404-style not-found error")
}

func TestTopupServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TopupServiceTestSuite))
}
