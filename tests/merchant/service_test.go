package merchant_test

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-merchant/repository"
	"github.com/MamangRust/monolith-payment-gateway-merchant/service"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	user_repo "github.com/MamangRust/monolith-payment-gateway-user/repository"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

type MerchantServiceTestSuite struct {
	suite.Suite
	harness         *tests.DependencyHarness
	ts              *tests.TestSuite
	merchantService service.Service
	mongoDB         *mongo.Database
	merchantID      int
	documentID      int
	userID          int
	apiKey          string
}

func (s *MerchantServiceTestSuite) SetupSuite() {
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
	repos := repository.NewRepositories(mongoDB, nil)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(redisClient, log, cacheMetrics)

	harness, err := tests.NewDependencyHarness(mongoDB, cacheStore, log)
	s.Require().NoError(err)
	s.harness = harness

	s.merchantService = service.NewService(&service.Deps{
		DB:           mongoDB,
		Repositories: repos,
		UserAdapter:  harness.UserAdapter,
		Logger:       log,
		Cache:        cacheStore,
		Kafka:        nil,
	})

	// Seed User
	userRepo := user_repo.NewUserCommandRepository(mongoDB.Collection("users"))
	user, err := userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Merchant",
		LastName:  "Tester",
		Email:     fmt.Sprintf("merchant.tester.%d@example.com", time.Now().UnixNano()),
		Password:  "password123",
	})
	s.Require().NoError(err)
	s.userID = int(user.UserID)
}

func (s *MerchantServiceTestSuite) TearDownSuite() {
	if s.harness != nil {
		s.harness.Close()
	}
	s.ts.Teardown()
}

// ────────────────────────────────────────────────────────────────────────────────
// MerchantCommandService tests
// ────────────────────────────────────────────────────────────────────────────────

func (s *MerchantServiceTestSuite) Test01_CreateMerchant() {
	ctx := context.Background()

	res, err := s.merchantService.MerchantCommandService().CreateMerchant(ctx, &requests.CreateMerchantRequest{
		UserID: s.userID,
		Name:   fmt.Sprintf("Test Merchant %d", time.Now().UnixNano()),
	})
	s.NoError(err)
	s.NotNil(res)
	s.merchantID = int(res.MerchantID)
}

func (s *MerchantServiceTestSuite) Test02_UpdateMerchant() {
	ctx := context.Background()
	s.Require().NotZero(s.merchantID)

	res, err := s.merchantService.MerchantCommandService().UpdateMerchant(ctx, &requests.UpdateMerchantRequest{
		MerchantID: &s.merchantID,
		Name:       "Updated Merchant Name",
		UserID:     s.userID,
		Status:     "inactive",
	})
	s.NoError(err)
	s.NotNil(res)
	s.Equal("Updated Merchant Name", res.Name)
}

func (s *MerchantServiceTestSuite) Test03_UpdateMerchantStatus() {
	ctx := context.Background()
	s.Require().NotZero(s.merchantID)

	res, err := s.merchantService.MerchantCommandService().UpdateMerchantStatus(ctx, &requests.UpdateMerchantStatusRequest{
		MerchantID: &s.merchantID,
		Status:     "active",
	})
	s.NoError(err)
	s.NotNil(res)
	s.Equal("active", res.Status)
}

// ────────────────────────────────────────────────────────────────────────────────
// MerchantQueryService tests
// ────────────────────────────────────────────────────────────────────────────────

func (s *MerchantServiceTestSuite) Test04_FindAll() {
	ctx := context.Background()
	s.Require().NotZero(s.merchantID)

	res, total, err := s.merchantService.MerchantQueryService().FindAll(ctx, &requests.FindAllMerchants{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.NotNil(total)
	s.GreaterOrEqual(len(res), 1)
}

func (s *MerchantServiceTestSuite) Test05_FindById() {
	ctx := context.Background()
	s.Require().NotZero(s.merchantID)

	res, err := s.merchantService.MerchantQueryService().FindById(ctx, s.merchantID)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(int32(s.merchantID), res.MerchantID)

	// Capture api_key for later transaction tests
	s.apiKey = res.ApiKey
}

func (s *MerchantServiceTestSuite) Test06_FindByActive() {
	ctx := context.Background()
	s.Require().NotZero(s.merchantID)

	res, total, err := s.merchantService.MerchantQueryService().FindByActive(ctx, &requests.FindAllMerchants{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.NotNil(total)
	s.GreaterOrEqual(len(res), 1)
}

func (s *MerchantServiceTestSuite) Test07_FindByApiKey() {
	ctx := context.Background()
	s.Require().NotEmpty(s.apiKey)

	res, err := s.merchantService.MerchantQueryService().FindByApiKey(ctx, s.apiKey)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(s.apiKey, res.ApiKey)
}

func (s *MerchantServiceTestSuite) Test08_FindByMerchantUserId() {
	ctx := context.Background()

	res, err := s.merchantService.MerchantQueryService().FindByMerchantUserId(ctx, s.userID)
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

// ────────────────────────────────────────────────────────────────────────────────
// MerchantDocumentCommandService tests
// ────────────────────────────────────────────────────────────────────────────────

func (s *MerchantServiceTestSuite) Test09_CreateMerchantDocument() {
	ctx := context.Background()
	s.Require().NotZero(s.merchantID)

	res, err := s.merchantService.MerchantDocumentCommandService().CreateMerchantDocument(ctx, &requests.CreateMerchantDocumentRequest{
		MerchantID:   s.merchantID,
		DocumentType: "Identity Proof",
		DocumentUrl:  "http://example.com/doc.pdf",
	})
	s.NoError(err)
	s.NotNil(res)
	s.documentID = int(res.DocumentID)
}

func (s *MerchantServiceTestSuite) Test10_UpdateMerchantDocument() {
	ctx := context.Background()
	s.Require().NotZero(s.documentID)

	res, err := s.merchantService.MerchantDocumentCommandService().UpdateMerchantDocument(ctx, &requests.UpdateMerchantDocumentRequest{
		DocumentID:   &s.documentID,
		MerchantID:   s.merchantID,
		DocumentType: "Updated Type",
		DocumentUrl:  "http://example.com/updated.pdf",
		Status:       "pending",
		Note:         "Please re-upload",
	})
	s.NoError(err)
	s.NotNil(res)
	s.Equal("Updated Type", res.DocumentType)
}

func (s *MerchantServiceTestSuite) Test11_UpdateMerchantDocumentStatus() {
	ctx := context.Background()
	s.Require().NotZero(s.documentID)

	res, err := s.merchantService.MerchantDocumentCommandService().UpdateMerchantDocumentStatus(ctx, &requests.UpdateMerchantDocumentStatusRequest{
		DocumentID: &s.documentID,
		MerchantID: s.merchantID,
		Status:     "approved",
		Note:       "All good",
	})
	s.NoError(err)
	s.NotNil(res)
	s.Equal("approved", res.Status)
}

// ────────────────────────────────────────────────────────────────────────────────
// MerchantDocumentQueryService tests
// ────────────────────────────────────────────────────────────────────────────────

func (s *MerchantServiceTestSuite) Test12_FindAllDocuments() {
	ctx := context.Background()
	s.Require().NotZero(s.documentID)

	res, total, err := s.merchantService.MerchantDocumentQueryService().FindAll(ctx, &requests.FindAllMerchantDocuments{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.NotNil(total)
	s.GreaterOrEqual(len(res), 1)
}

func (s *MerchantServiceTestSuite) Test13_FindByIdDocument() {
	ctx := context.Background()
	s.Require().NotZero(s.documentID)

	res, err := s.merchantService.MerchantDocumentQueryService().FindById(ctx, s.documentID)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(int32(s.documentID), res.DocumentID)
}

func (s *MerchantServiceTestSuite) Test14_FindByActiveDocuments() {
	ctx := context.Background()

	res, total, err := s.merchantService.MerchantDocumentQueryService().FindByActive(ctx, &requests.FindAllMerchantDocuments{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.NotNil(total)
	s.GreaterOrEqual(len(res), 1)
}

// ────────────────────────────────────────────────────────────────────────────────
// MerchantTransactionService tests
// ────────────────────────────────────────────────────────────────────────────────

func (s *MerchantServiceTestSuite) seedTransaction(cardNumber string, merchantID int32) {
	_, err := s.mongoDB.Collection("transactions").InsertOne(context.Background(), bson.M{
		"transaction_id":   time.Now().UnixNano() % 1000000000,
		"card_number":      cardNumber,
		"merchant_id":      merchantID,
		"amount":           150000,
		"payment_method":   "bank_transfer",
		"transaction_time": time.Now(),
		"status":           "success",
	})
	s.Require().NoError(err)
}

func (s *MerchantServiceTestSuite) Test15_FindAllTransactions() {
	ctx := context.Background()
	s.Require().NotZero(s.merchantID)

	// Seed card + transaction
	cardNumber := fmt.Sprintf("%016d", time.Now().UnixNano()%1e16)
	_, err := s.mongoDB.Collection("cards").InsertOne(ctx, bson.M{
		"user_id":       s.userID,
		"card_number":   cardNumber,
		"card_type":     "debit",
		"cvv":           "123",
		"card_provider": "visa",
		"expire_date":   time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	s.Require().NoError(err)

	s.seedTransaction(cardNumber, int32(s.merchantID))

	res, total, err := s.merchantService.MerchantTransactionService().FindAllTransactions(ctx, &requests.FindAllMerchantTransactions{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.NotNil(total)
	s.GreaterOrEqual(len(res), 1)
}

func (s *MerchantServiceTestSuite) Test16_FindAllTransactionsByMerchant() {
	ctx := context.Background()
	s.Require().NotZero(s.merchantID)

	res, total, err := s.merchantService.MerchantTransactionService().FindAllTransactionsByMerchant(ctx, &requests.FindAllMerchantTransactionsById{
		MerchantID: s.merchantID,
		Page:       1,
		PageSize:   10,
		Search:     "",
	})
	s.NoError(err)
	s.NotNil(total)
	s.GreaterOrEqual(len(res), 1)
}

func (s *MerchantServiceTestSuite) Test17_FindAllTransactionsByApikey() {
	ctx := context.Background()
	s.Require().NotEmpty(s.apiKey)

	res, total, err := s.merchantService.MerchantTransactionService().FindAllTransactionsByApikey(ctx, &requests.FindAllMerchantTransactionsByApiKey{
		ApiKey:   s.apiKey,
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.NotNil(total)
	s.GreaterOrEqual(len(res), 1)
}

// ────────────────────────────────────────────────────────────────────────────────
// Trash / Restore / Delete
// ────────────────────────────────────────────────────────────────────────────────

func (s *MerchantServiceTestSuite) Test18_TrashAndRestoreDocument() {
	ctx := context.Background()
	s.Require().NotZero(s.documentID)

	trashed, err := s.merchantService.MerchantDocumentCommandService().TrashedMerchantDocument(ctx, s.documentID)
	s.NoError(err)
	s.NotNil(trashed)

	// Verify it appears in trashed list
	resTrashed, _, err := s.merchantService.MerchantDocumentQueryService().FindByTrashed(ctx, &requests.FindAllMerchantDocuments{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.GreaterOrEqual(len(resTrashed), 1)

	restored, err := s.merchantService.MerchantDocumentCommandService().RestoreMerchantDocument(ctx, s.documentID)
	s.NoError(err)
	s.NotNil(restored)
}

func (s *MerchantServiceTestSuite) Test19_TrashAndRestoreMerchant() {
	ctx := context.Background()
	s.Require().NotZero(s.merchantID)

	trashed, err := s.merchantService.MerchantCommandService().TrashedMerchant(ctx, s.merchantID)
	s.NoError(err)
	s.NotNil(trashed)

	// Verify it appears in trashed list
	resTrashed, _, err := s.merchantService.MerchantQueryService().FindByTrashed(ctx, &requests.FindAllMerchants{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.GreaterOrEqual(len(resTrashed), 1)

	restored, err := s.merchantService.MerchantCommandService().RestoreMerchant(ctx, s.merchantID)
	s.NoError(err)
	s.NotNil(restored)
}

func (s *MerchantServiceTestSuite) Test20_DeleteMerchantPermanent() {
	ctx := context.Background()

	// Create a standalone merchant to delete
	res, err := s.merchantService.MerchantCommandService().CreateMerchant(ctx, &requests.CreateMerchantRequest{
		UserID: s.userID,
		Name:   fmt.Sprintf("ToDelete %d", time.Now().UnixNano()),
	})
	s.Require().NoError(err)
	tmpID := int(res.MerchantID)

	_, err = s.merchantService.MerchantCommandService().TrashedMerchant(ctx, tmpID)
	s.Require().NoError(err)

	ok, err := s.merchantService.MerchantCommandService().DeleteMerchantPermanent(ctx, tmpID)
	s.NoError(err)
	s.True(ok)
}

func (s *MerchantServiceTestSuite) Test21_DeleteMerchantDocumentPermanent() {
	ctx := context.Background()

	// Create fresh merchant + document to delete
	m, err := s.merchantService.MerchantCommandService().CreateMerchant(ctx, &requests.CreateMerchantRequest{
		UserID: s.userID,
		Name:   fmt.Sprintf("DocDelete %d", time.Now().UnixNano()),
	})
	s.Require().NoError(err)

	doc, err := s.merchantService.MerchantDocumentCommandService().CreateMerchantDocument(ctx, &requests.CreateMerchantDocumentRequest{
		MerchantID:   int(m.MerchantID),
		DocumentType: "Delete Test",
		DocumentUrl:  "http://example.com/delete.pdf",
	})
	s.Require().NoError(err)
	tmpDocID := int(doc.DocumentID)

	_, err = s.merchantService.MerchantDocumentCommandService().TrashedMerchantDocument(ctx, tmpDocID)
	s.Require().NoError(err)

	ok, err := s.merchantService.MerchantDocumentCommandService().DeleteMerchantDocumentPermanent(ctx, tmpDocID)
	s.NoError(err)
	s.True(ok)
}

func (s *MerchantServiceTestSuite) Test22_BulkRestoreAndDeleteAll() {
	ctx := context.Background()

	ok, err := s.merchantService.MerchantCommandService().RestoreAllMerchant(ctx)
	s.NoError(err)
	s.True(ok)

	ok, err = s.merchantService.MerchantDocumentCommandService().RestoreAllMerchantDocument(ctx)
	s.NoError(err)
	s.True(ok)

	ok, err = s.merchantService.MerchantCommandService().DeleteAllMerchantPermanent(ctx)
	s.NoError(err)
	s.True(ok)

	ok, err = s.merchantService.MerchantDocumentCommandService().DeleteAllMerchantDocumentPermanent(ctx)
	s.NoError(err)
	s.True(ok)
}

func (s *MerchantServiceTestSuite) TestZZ_NotFound404() {
	ctx := context.Background()
	_, err := s.merchantService.MerchantQueryService().FindById(ctx, 999999999)
	s.Require().Error(err)
	s.Contains(err.Error(), "not found")

	_, err = s.merchantService.MerchantDocumentQueryService().FindById(ctx, 999999999)
	s.Require().Error(err)
	s.Contains(err.Error(), "not found")
}

func TestMerchantServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(MerchantServiceTestSuite))
}
