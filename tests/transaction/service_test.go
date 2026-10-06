package transaction_test

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	card_repo_impl "github.com/MamangRust/monolith-payment-gateway-card/repository"
	merchant_repo_impl "github.com/MamangRust/monolith-payment-gateway-merchant/repository"
	models "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	saldo_repo_impl "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/MamangRust/monolith-payment-gateway-transaction/repository"
	"github.com/MamangRust/monolith-payment-gateway-transaction/service"
	user_repo_impl "github.com/MamangRust/monolith-payment-gateway-user/repository"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

type testMerchantRepo struct {
	db *mongo.Database
}

func (r *testMerchantRepo) FindByApiKey(ctx context.Context, api_key string) (*models.MerchantAllFieldsRow, error) {
	var result models.Merchant
	err := r.db.Collection("merchants").FindOne(ctx, bson.M{"api_key": api_key}).Decode(&result)
	if err != nil {
		return nil, err
	}
	var createdAt, updatedAt time.Time
	if result.CreatedAt != nil {
		createdAt = *result.CreatedAt
	}
	if result.UpdatedAt != nil {
		updatedAt = *result.UpdatedAt
	}
	return &models.MerchantAllFieldsRow{
		MerchantID: result.MerchantID, UserID: result.UserID, ApiKey: result.ApiKey,
		CreatedAt: createdAt, UpdatedAt: updatedAt,
	}, nil
}

type testCardRepo struct {
	db *mongo.Database
}

func (r *testCardRepo) FindCardByUserId(ctx context.Context, user_id int) (*models.CardAllFieldsRow, error) {
	var result models.Card
	err := r.db.Collection("cards").FindOne(ctx, bson.M{"user_id": int32(user_id)}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &models.CardAllFieldsRow{CardID: result.CardID, UserID: result.UserID, CardNumber: result.CardNumber, CardType: result.CardType, ExpireDate: result.ExpireDate, Cvv: result.Cvv, CardProvider: result.CardProvider, Status: result.Status}, nil
}

func (r *testCardRepo) FindUserCardByCardNumber(ctx context.Context, card_number string) (*models.CardByEmailRow, error) {
	var card models.Card
	err := r.db.Collection("cards").FindOne(ctx, bson.M{"card_number": card_number}).Decode(&card)
	if err != nil {
		return nil, err
	}
	var user struct{ Email string }
	_ = r.db.Collection("users").FindOne(ctx, bson.M{"user_id": card.UserID}).Decode(&user)
	return &models.CardByEmailRow{CardID: card.CardID, UserID: card.UserID, CardNumber: card.CardNumber, CardType: card.CardType, ExpireDate: card.ExpireDate, Cvv: card.Cvv, CardProvider: card.CardProvider, Email: user.Email}, nil
}

func (r *testCardRepo) FindCardByCardNumber(ctx context.Context, card_number string) (*models.CardAllFieldsRow, error) {
	var result models.Card
	err := r.db.Collection("cards").FindOne(ctx, bson.M{"card_number": card_number}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &models.CardAllFieldsRow{CardID: result.CardID, UserID: result.UserID, CardNumber: result.CardNumber, CardType: result.CardType, ExpireDate: result.ExpireDate, Cvv: result.Cvv, CardProvider: result.CardProvider, Status: result.Status}, nil
}

func (r *testCardRepo) UpdateCard(ctx context.Context, request *requests.UpdateCardRequest) (*models.CardUpdateRow, error) {
	_, err := r.db.Collection("cards").UpdateOne(ctx,
		bson.M{"card_id": int32(request.CardID)},
		bson.M{"$set": bson.M{"card_type": request.CardType, "expire_date": request.ExpireDate, "cvv": request.CVV, "card_provider": request.CardProvider}})
	if err != nil {
		return nil, err
	}
	var result models.Card
	err = r.db.Collection("cards").FindOne(ctx, bson.M{"card_id": int32(request.CardID)}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &models.CardUpdateRow{CardID: result.CardID, UserID: result.UserID, CardNumber: result.CardNumber, CardType: result.CardType, ExpireDate: result.ExpireDate, Cvv: result.Cvv, CardProvider: result.CardProvider}, nil
}

func (r *testCardRepo) UpdateCardOutstandingBalance(ctx context.Context, cardID int, outstandingBalance int) (*models.Card, error) {
	_, err := r.db.Collection("cards").UpdateOne(ctx,
		bson.M{"card_id": int32(cardID)},
		bson.M{"$set": bson.M{"outstanding_balance": outstandingBalance}})
	if err != nil {
		return nil, err
	}
	var result models.Card
	err = r.db.Collection("cards").FindOne(ctx, bson.M{"card_id": int32(cardID)}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *testCardRepo) AddRewardPoints(ctx context.Context, cardID int, points int) (*models.AddRewardPointsRow, error) {
	_, err := r.db.Collection("cards").UpdateOne(ctx,
		bson.M{"card_id": int32(cardID)},
		bson.M{"$inc": bson.M{"reward_points": int32(points)}})
	if err != nil {
		return nil, err
	}
	var result models.Card
	err = r.db.Collection("cards").FindOne(ctx, bson.M{"card_id": int32(cardID)}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &models.AddRewardPointsRow{
		CardID: result.CardID, RewardPoints: result.RewardPoints,
	}, nil
}

type testSaldoRepo struct {
	db *mongo.Database
}

func (r *testSaldoRepo) FindByCardNumber(ctx context.Context, card_number string) (*models.Saldo, error) {
	var result models.Saldo
	err := r.db.Collection("saldos").FindOne(ctx, bson.M{"card_number": card_number}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *testSaldoRepo) UpdateSaldoBalance(ctx context.Context, request *requests.UpdateSaldoBalance) (*models.UpdateSaldoBalanceRow, error) {
	_, err := r.db.Collection("saldos").UpdateOne(ctx,
		bson.M{"card_number": request.CardNumber},
		bson.M{"$set": bson.M{"total_balance": request.TotalBalance}})
	if err != nil {
		return nil, err
	}
	return &models.UpdateSaldoBalanceRow{
		SaldoID: 0, CardNumber: request.CardNumber, TotalBalance: int32(request.TotalBalance),
	}, nil
}

type TransactionServiceTestSuite struct {
	suite.Suite
	ts                  *tests.TestSuite
	transactionService  service.Service
	mongoDB             *mongo.Database
	transactionID       int
	cardNumber          string
	merchantID          int
	merchantApiKey      string
	otherMerchantApiKey string
	userID              int
	merchantUserID      int
	nextDirectTxID      int32
}

func (s *TransactionServiceTestSuite) SetupSuite() {
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
	merchantRepo := &testMerchantRepo{db: mongoDB}
	repos := repository.NewRepositoriesWithAdapters(mongoDB, saldoRepo, cardRepo, merchantRepo)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(redisClient, log, cacheMetrics)

	harness, hErr := tests.NewDependencyHarness(mongoDB, cacheStore, log)
	s.Require().NoError(hErr)
	s.T().Cleanup(harness.Close)

	// Directly seeded transactions use high IDs so they never collide with
	// IDs allocated by the service's own counters.
	s.nextDirectTxID = 9000

	s.transactionService = service.NewService(&service.Deps{
		CardAdapter:     harness.CardAdapter,
		SaldoAdapter:    harness.SaldoAdapter,
		MerchantAdapter: harness.MerchantAdapter,
		DB:              mongoDB,
		Kafka:           nil,
		Repositories:    repos,
		Logger:          log,
		Cache:           cacheStore,
	})

	// Seed User
	userRepo := user_repo_impl.NewUserCommandRepository(mongoDB.Collection("users"))
	user, err := userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Transaction",
		LastName:  "Tester",
		Email:     fmt.Sprintf("transaction.tester.%d@example.com", time.Now().UnixNano()),
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

	// Seed a separate merchant owner.
	merchantOwner, err := userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Merchant",
		LastName:  "Owner",
		Email:     fmt.Sprintf("merchant.owner.%d@example.com", time.Now().UnixNano()),
		Password:  "password123",
	})
	s.Require().NoError(err)
	s.merchantUserID = int(merchantOwner.UserID)

	// Seed Merchant
	merchantRepoImpl := merchant_repo_impl.NewMerchantCommandRepository(mongoDB, mongoDB.Collection("merchants"))
	merchant, err := merchantRepoImpl.CreateMerchant(context.Background(), &requests.CreateMerchantRequest{
		UserID: s.merchantUserID,
		Name:   "Test Merchant",
	})
	s.Require().NoError(err)
	s.merchantID = int(merchant.MerchantID)
	s.merchantApiKey = merchant.ApiKey

	otherMerchantOwner, err := userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Other Merchant",
		LastName:  "Owner",
		Email:     fmt.Sprintf("other.merchant.owner.%d@example.com", time.Now().UnixNano()),
		Password:  "password123",
	})
	s.Require().NoError(err)
	otherMerchant, err := merchantRepoImpl.CreateMerchant(context.Background(), &requests.CreateMerchantRequest{
		UserID: int(otherMerchantOwner.UserID),
		Name:   "Other Test Merchant",
	})
	s.Require().NoError(err)
	s.otherMerchantApiKey = otherMerchant.ApiKey

	// Seed Merchant Card & Saldo
	merchantCard, err := cardCmdRepo.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID:       s.merchantUserID,
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(5, 0, 0),
		CVV:          "456",
		CardProvider: "mastercard",
	})
	s.Require().NoError(err)
	_, err = saldoCmdRepo.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{
		CardNumber:   merchantCard.CardNumber,
		TotalBalance: 1000000,
	})
	s.Require().NoError(err)
}

func (s *TransactionServiceTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

// helper to read saldo balance via GORM
func (s *TransactionServiceTestSuite) getSaldoBalance(ctx context.Context, cardNumber string) int64 {
	var saldo models.Saldo
	err := s.mongoDB.Collection("saldos").FindOne(ctx, bson.M{"card_number": cardNumber}).Decode(&saldo)
	s.Require().NoError(err)
	return saldo.TotalBalance
}

// helper to read merchant card number
func (s *TransactionServiceTestSuite) getMerchantCardNumber(ctx context.Context, userID int) string {
	var card models.Card
	err := s.mongoDB.Collection("cards").FindOne(ctx, bson.M{"user_id": int32(userID)}).Decode(&card)
	s.Require().NoError(err)
	return card.CardNumber
}

// helper to create a transaction directly for ownership/state-machine tests
func (s *TransactionServiceTestSuite) createDirectTransaction(ctx context.Context, cardNumber string, amount int, paymentMethod string, merchantID int32) *models.Transaction {
	now := time.Now()
	txID := s.nextDirectTxID
	s.nextDirectTxID++
	tx := models.Transaction{
		TransactionID:   txID,
		CardNumber:      cardNumber,
		Amount:          int64(amount),
		PaymentMethod:   paymentMethod,
		MerchantID:      merchantID,
		TransactionTime: now,
		Status:          "authorized",
		CreatedAt:       &now,
		UpdatedAt:       &now,
	}
	_, err := s.mongoDB.Collection("transactions").InsertOne(ctx, bson.M{
		"transaction_id": tx.TransactionID, "transaction_no": uuid.New(),
		"card_number": tx.CardNumber, "amount": tx.Amount, "payment_method": tx.PaymentMethod,
		"merchant_id": tx.MerchantID, "transaction_time": tx.TransactionTime,
		"status": tx.Status, "created_at": tx.CreatedAt, "updated_at": tx.UpdatedAt})
	s.Require().NoError(err)
	return &tx
}

// helper to update transaction status directly via GORM
func (s *TransactionServiceTestSuite) updateTransactionStatus(ctx context.Context, transactionID int32, status string) {
	_, err := s.mongoDB.Collection("transactions").UpdateOne(ctx,
		bson.M{"transaction_id": transactionID},
		bson.M{"$set": bson.M{"status": status}})
	s.Require().NoError(err)
}

// helper to count transactions by merchant
func (s *TransactionServiceTestSuite) countTransactionsByMerchant(ctx context.Context, merchantID int) int {
	return s.countTransactionsByMerchantAndStatus(ctx, merchantID, "")
}

// countTransactionsByMerchantAndStatus counts the merchant's transactions,
// optionally filtered by status. The service records losing or failed attempts
// as failed/unknown rows for reconciliation, so tests that assert "settled
// once" must count the settled (success) rows rather than every row.
func (s *TransactionServiceTestSuite) countTransactionsByMerchantAndStatus(ctx context.Context, merchantID int, status string) int {
	filter := bson.M{"merchant_id": int32(merchantID)}
	if status != "" {
		filter["status"] = status
	}
	count, err := s.mongoDB.Collection("transactions").CountDocuments(ctx, filter)
	s.Require().NoError(err)
	return int(count)
}

func (s *TransactionServiceTestSuite) Test1_TransactionLifecycle() {
	ctx := context.Background()

	// Create Transaction
	merchantID := int(s.merchantID)
	createReq := &requests.CreateTransactionRequest{
		CardNumber:      s.cardNumber,
		Amount:          100000,
		PaymentMethod:   "gopay",
		MerchantID:      &merchantID,
		TransactionTime: time.Now(),
	}
	res, err := s.transactionService.Create(ctx, s.merchantApiKey, createReq)
	s.NoError(err)
	s.NotNil(res)
	s.transactionID = int(res.TransactionID)
	s.Equal("success", res.Status)

	userSaldo := s.getSaldoBalance(ctx, s.cardNumber)
	s.Equal(int64(900000), userSaldo)

	// FindById
	found, err := s.transactionService.FindById(ctx, s.transactionID)
	s.NoError(err)
	s.NotNil(found)
	s.Equal(int32(s.transactionID), found.TransactionID)

	// Update Transaction
	updateReq := &requests.UpdateTransactionRequest{
		TransactionID:   &s.transactionID,
		CardNumber:      s.cardNumber,
		Amount:          200000,
		PaymentMethod:   "dana",
		MerchantID:      &merchantID,
		TransactionTime: time.Now(),
	}
	updated, err := s.transactionService.Update(ctx, s.merchantApiKey, updateReq)
	s.NoError(err)
	s.NotNil(updated)
	s.Equal(int64(200000), updated.Amount)
}

func (s *TransactionServiceTestSuite) TestOwnershipChecksRejectVoidAndRefund() {
	ctx := context.Background()
	merchantID := int32(s.merchantID)

	beforeUser := s.getSaldoBalance(ctx, s.cardNumber)
	merchantCard := s.getMerchantCardNumber(ctx, s.merchantUserID)
	beforeMerchant := s.getSaldoBalance(ctx, merchantCard)

	voidTx := s.createDirectTransaction(ctx, s.cardNumber, 100, "gopay", merchantID)
	s.updateTransactionStatus(ctx, voidTx.TransactionID, "authorized")

	_, err := s.transactionService.VoidTransaction(ctx, s.otherMerchantApiKey, int(voidTx.TransactionID))
	s.Require().Error(err)
	s.Contains(err.Error(), "merchant does not own transaction")

	refundTx := s.createDirectTransaction(ctx, s.cardNumber, 100, "gopay", merchantID)
	s.updateTransactionStatus(ctx, refundTx.TransactionID, "captured")

	_, err = s.transactionService.RefundTransaction(ctx, s.otherMerchantApiKey, int(refundTx.TransactionID))
	s.Require().Error(err)
	s.Contains(err.Error(), "merchant does not own transaction")

	afterUser := s.getSaldoBalance(ctx, s.cardNumber)
	s.Equal(beforeUser, afterUser)
	afterMerchant := s.getSaldoBalance(ctx, merchantCard)
	s.Equal(beforeMerchant, afterMerchant)
}

func (s *TransactionServiceTestSuite) Test2_AtomicCreateRejectsInsufficientBalance() {
	ctx := context.Background()
	merchantID := s.merchantID
	beforeUser := s.getSaldoBalance(ctx, s.cardNumber)

	merchantCard := s.getMerchantCardNumber(ctx, s.merchantUserID)
	beforeMerchant := s.getSaldoBalance(ctx, merchantCard)
	beforeTransactions := s.countTransactionsByMerchant(ctx, s.merchantID)

	res, err := s.transactionService.Create(ctx, s.merchantApiKey, &requests.CreateTransactionRequest{
		CardNumber:      s.cardNumber,
		Amount:          int(beforeUser) + 1,
		PaymentMethod:   "gopay",
		MerchantID:      &merchantID,
		TransactionTime: time.Now(),
	})
	s.Error(err)
	s.Nil(res)

	afterUser := s.getSaldoBalance(ctx, s.cardNumber)
	s.Equal(beforeUser, afterUser)
	afterMerchant := s.getSaldoBalance(ctx, merchantCard)
	s.Equal(beforeMerchant, afterMerchant)
	afterTransactions := s.countTransactionsByMerchant(ctx, s.merchantID)
	s.Equal(beforeTransactions, afterTransactions)
}

func (s *TransactionServiceTestSuite) Test3_ConcurrentAtomicCreatesSettleOnce() {
	ctx := context.Background()
	merchantID := s.merchantID
	beforeUser := s.getSaldoBalance(ctx, s.cardNumber)
	merchantCard := s.getMerchantCardNumber(ctx, s.merchantUserID)
	beforeMerchant := s.getSaldoBalance(ctx, merchantCard)
	beforeSettled := s.countTransactionsByMerchantAndStatus(ctx, s.merchantID, "success")
	beforeAttempts := s.countTransactionsByMerchant(ctx, s.merchantID)

	amount := int(beforeUser)/2 + 1
	s.Require().GreaterOrEqual(int(beforeUser), amount)
	s.Require().Less(int(beforeUser), amount*2)

	requestsReady := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func() {
			defer wg.Done()
			<-requestsReady
			_, err := s.transactionService.Create(ctx, s.merchantApiKey, &requests.CreateTransactionRequest{
				CardNumber:      s.cardNumber,
				Amount:          amount,
				PaymentMethod:   "gopay",
				MerchantID:      &merchantID,
				TransactionTime: time.Now(),
			})
			results <- err
		}()
	}
	close(requestsReady)
	wg.Wait()
	close(results)

	successes := 0
	for err := range results {
		if err == nil {
			successes++
		}
	}
	s.Equal(1, successes)

	afterUser := s.getSaldoBalance(ctx, s.cardNumber)
	s.Equal(beforeUser-int64(amount), afterUser)
	afterMerchant := s.getSaldoBalance(ctx, merchantCard)
	s.Equal(beforeMerchant+int64(amount), afterMerchant)

	// Exactly one attempt settled. Whether the losing attempt leaves a row is
	// timing-dependent: if it reaches the balance pre-check after the winner
	// debited, insufficient balance short-circuits before any row is written
	// (see Test2); if both pass the pre-check concurrently, the loser is
	// recorded and then failed by the atomic debit guard. Both outcomes are
	// correct, so accept one or two new attempts while still requiring exactly
	// one settled row.
	s.Equal(beforeSettled+1, s.countTransactionsByMerchantAndStatus(ctx, s.merchantID, "success"))
	afterAttempts := s.countTransactionsByMerchant(ctx, s.merchantID)
	s.Require().Contains([]int{beforeAttempts + 1, beforeAttempts + 2}, afterAttempts,
		"expected one settled attempt and at most one recorded losing attempt")
}

func (s *TransactionServiceTestSuite) Test4_QueryOperations() {
	ctx := context.Background()

	// FindAll
	all, total, err := s.transactionService.FindAll(ctx, &requests.FindAllTransactions{
		Page:     1,
		PageSize: 10,
	})
	s.NoError(err)
	s.NotNil(all)
	s.GreaterOrEqual(*total, 1)

	// FindByActive
	active, totalActive, err := s.transactionService.FindByActive(ctx, &requests.FindAllTransactions{
		Page:     1,
		PageSize: 10,
	})
	s.NoError(err)
	s.NotNil(active)
	s.GreaterOrEqual(*totalActive, 1)

	// FindByCardNumber
	byCard, totalCard, err := s.transactionService.FindAllByCardNumber(ctx, &requests.FindAllTransactionCardNumber{
		CardNumber: s.cardNumber,
		Page:       1,
		PageSize:   10,
	})
	s.NoError(err)
	s.NotNil(byCard)
	s.GreaterOrEqual(*totalCard, 1)

	// FindByMerchantId
	byMerchant, err := s.transactionService.FindTransactionByMerchantId(ctx, s.merchantID)
	s.NoError(err)
	s.NotNil(byMerchant)
	s.GreaterOrEqual(len(byMerchant), 1)
}

func (s *TransactionServiceTestSuite) Test5_TrashAndRestore() {
	ctx := context.Background()
	s.Require().NotZero(s.transactionID)

	// Trash
	trashed, err := s.transactionService.TrashedTransaction(ctx, s.transactionID)
	s.NoError(err)
	s.NotNil(trashed)

	// FindByTrashed
	trashedList, totalTrashed, err := s.transactionService.FindByTrashed(ctx, &requests.FindAllTransactions{
		Page:     1,
		PageSize: 10,
	})
	s.NoError(err)
	s.NotNil(trashedList)
	s.GreaterOrEqual(*totalTrashed, 1)

	// Restore
	restored, err := s.transactionService.RestoreTransaction(ctx, s.transactionID)
	s.NoError(err)
	s.NotNil(restored)
}

func (s *TransactionServiceTestSuite) Test6_BulkOperations() {
	ctx := context.Background()

	// Restore All
	ok, err := s.transactionService.RestoreAllTransaction(ctx)
	s.NoError(err)
	s.True(ok)

	// Delete All Permanent
	ok, err = s.transactionService.DeleteAllTransactionPermanent(ctx)
	s.NoError(err)
	s.True(ok)
}

func (s *TransactionServiceTestSuite) Test7_TransactionIdempotentReplay() {
	ctx := context.Background()
	key := fmt.Sprintf("transaction:create:merchant:%d:op:%d", s.merchantID, time.Now().UnixNano())
	amount := 50000
	merchantID := s.merchantID

	userBefore := s.getSaldoBalance(ctx, s.cardNumber)

	req := &requests.CreateTransactionRequest{
		CardNumber:      s.cardNumber,
		Amount:          amount,
		PaymentMethod:   "gopay",
		MerchantID:      &merchantID,
		TransactionTime: time.Now(),
		IdempotencyKey:  key,
	}

	// A replay with the same key must return the original transaction and must
	// NOT settle a second time.
	first, err := s.transactionService.Create(ctx, s.merchantApiKey, req)
	s.Require().NoError(err)
	s.Require().NotNil(first)

	second, err := s.transactionService.Create(ctx, s.merchantApiKey, req)
	s.Require().NoError(err)
	s.Require().NotNil(second)
	s.Equal(first.TransactionID, second.TransactionID, "replay must return the original transaction")

	userAfter := s.getSaldoBalance(ctx, s.cardNumber)
	s.Equal(userBefore-int64(amount), userAfter, "user must be debited exactly once")
}

func (s *TransactionServiceTestSuite) Test8_InvalidStateTransitionRejected400() {
	ctx := context.Background()
	merchantID := s.merchantID

	merchantCard := s.getMerchantCardNumber(ctx, s.merchantUserID)
	userBefore := s.getSaldoBalance(ctx, s.cardNumber)
	merchantBefore := s.getSaldoBalance(ctx, merchantCard)

	// The atomic create settles immediately into "success"; lifecycle
	// operations whose source state precondition is not met must be rejected
	// 400 via the state machine WITHOUT moving any balance.
	key := fmt.Sprintf("transaction:create:merchant:%d:transition:%d", s.merchantID, time.Now().UnixNano())
	res, err := s.transactionService.Create(ctx, s.merchantApiKey, &requests.CreateTransactionRequest{
		CardNumber:      s.cardNumber,
		Amount:          50000,
		PaymentMethod:   "gopay",
		MerchantID:      &merchantID,
		TransactionTime: time.Now(),
		IdempotencyKey:  key,
	})
	s.Require().NoError(err)
	txID := int(res.TransactionID)

	_, err = s.transactionService.CaptureTransaction(ctx, s.merchantApiKey, txID)
	s.Require().Error(err)
	s.Contains(err.Error(), "invalid status transition", "capture from success must be 400")

	_, err = s.transactionService.VoidTransaction(ctx, s.merchantApiKey, txID)
	s.Require().Error(err)
	s.Contains(err.Error(), "invalid status transition", "void from success must be 400")

	_, err = s.transactionService.RefundTransaction(ctx, s.merchantApiKey, txID)
	s.Require().Error(err)
	s.Contains(err.Error(), "invalid status transition", "refund from success must be 400")

	userAfter := s.getSaldoBalance(ctx, s.cardNumber)
	merchantAfter := s.getSaldoBalance(ctx, merchantCard)
	s.Equal(userBefore-int64(50000), userAfter, "only the original settlement may move money")
	s.Equal(merchantBefore+int64(50000), merchantAfter)
}

func (s *TransactionServiceTestSuite) TestZZ_NotFound404() {
	ctx := context.Background()
	_, err := s.transactionService.FindById(ctx, 999999999)
	s.Require().Error(err)
	s.Contains(err.Error(), "not found")
}

func (s *TransactionServiceTestSuite) Test9_FindByIdNotFound() {
	ctx := context.Background()
	_, err := s.transactionService.FindById(ctx, 999999999)
	s.Require().Error(err, "a non-existent transaction must not resolve")
	s.Contains(err.Error(), "not found", "missing transaction must surface as a 404-style not-found error")
}

func TestTransactionServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransactionServiceTestSuite))
}
