package merchant_test

import (
	"context"
	"fmt"
	models "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-merchant/repository"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	sharederrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	user_repo "github.com/MamangRust/monolith-payment-gateway-user/repository"

	"github.com/stretchr/testify/suite"
	"net/http"
)

type MerchantRepositoryTestSuite struct {
	suite.Suite
	ts       *tests.TestSuite
	mongoDB  *mongo.Database
	repo     repository.Repositories
	userRepo user_repo.Repositories
	userID   int
}

func (s *MerchantRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	mongoDB, err := s.ts.MongoDB()
	if err != nil {
		s.Require().NoError(err)
	}
	s.mongoDB = mongoDB
	s.repo = repository.NewRepositories(mongoDB, nil)
	s.userRepo = user_repo.NewRepositories(mongoDB, nil, nil)

	// Create a user for merchant ownership
	user, err := s.userRepo.UserCommand().CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Merchant",
		LastName:  "Owner",
		Email:     fmt.Sprintf("merchant.owner-%d-%d@example.com", time.Now().UnixNano(), time.Now().UnixNano()%10000),
		Password:  "password123",
	})
	s.Require().NoError(err)
	s.userID = int(user.UserID)
}

func (s *MerchantRepositoryTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *MerchantRepositoryTestSuite) createSeedMerchant() (*models.Merchant, error) {
	return s.repo.CreateMerchant(context.Background(), &requests.CreateMerchantRequest{
		UserID: s.userID,
		Name:   fmt.Sprintf("Test Merchant %d", time.Now().UnixNano()),
	})
}

func (s *MerchantRepositoryTestSuite) TestCreateMerchant() {
	ctx := context.Background()
	req := &requests.CreateMerchantRequest{
		UserID: s.userID,
		Name:   fmt.Sprintf("Test Merchant %d", time.Now().UnixNano()),
	}

	res, err := s.repo.CreateMerchant(ctx, req)
	s.NoError(err)
	s.NotNil(res)
}

func (s *MerchantRepositoryTestSuite) TestFindAllMerchants() {
	_, err := s.createSeedMerchant()
	s.Require().NoError(err)
	ctx := context.Background()

	res, err := s.repo.FindAllMerchants(ctx, &requests.FindAllMerchants{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *MerchantRepositoryTestSuite) TestFindByApiKey() {
	merchant, err := s.createSeedMerchant()
	s.Require().NoError(err)
	ctx := context.Background()

	found, err := s.repo.FindByApiKey(ctx, merchant.ApiKey)
	s.NoError(err)
	s.NotNil(found)
	s.Equal(merchant.ApiKey, found.ApiKey)
}

func (s *MerchantRepositoryTestSuite) TestFindByMerchantId() {
	merchant, err := s.createSeedMerchant()
	s.Require().NoError(err)
	ctx := context.Background()

	found, err := s.repo.FindByMerchantId(ctx, int(merchant.MerchantID))
	s.NoError(err)
	s.NotNil(found)
	s.Equal(merchant.MerchantID, found.MerchantID)
}

func (s *MerchantRepositoryTestSuite) TestUpdateMerchant() {
	merchant, err := s.createSeedMerchant()
	s.Require().NoError(err)
	ctx := context.Background()

	merchantID := int(merchant.MerchantID)
	req := &requests.UpdateMerchantRequest{
		MerchantID: &merchantID,
		Name:       "Updated Merchant",
		UserID:     s.userID,
		Status:     "active",
	}

	res, err := s.repo.UpdateMerchant(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal("Updated Merchant", res.Name)
}

func (s *MerchantRepositoryTestSuite) TestTrashMerchant() {
	merchant, err := s.createSeedMerchant()
	s.Require().NoError(err)
	ctx := context.Background()

	trashed, err := s.repo.TrashedMerchant(ctx, int(merchant.MerchantID))
	s.NoError(err)
	s.NotNil(trashed)
}

func (s *MerchantRepositoryTestSuite) TestRestoreMerchant() {
	merchant, err := s.createSeedMerchant()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.TrashedMerchant(ctx, int(merchant.MerchantID))
	s.Require().NoError(err)

	restored, err := s.repo.RestoreMerchant(ctx, int(merchant.MerchantID))
	s.NoError(err)
	s.NotNil(restored)
}

func (s *MerchantRepositoryTestSuite) TestDeleteMerchantPermanent() {
	merchant, err := s.createSeedMerchant()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.TrashedMerchant(ctx, int(merchant.MerchantID))
	s.Require().NoError(err)

	success, err := s.repo.DeleteMerchantPermanent(ctx, int(merchant.MerchantID))
	s.NoError(err)
	s.True(success)
}

func (s *MerchantRepositoryTestSuite) TestRestoreAllMerchant() {
	merchant, err := s.createSeedMerchant()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.TrashedMerchant(ctx, int(merchant.MerchantID))
	s.Require().NoError(err)

	success, err := s.repo.RestoreAllMerchant(ctx)
	s.NoError(err)
	s.True(success)
}

func (s *MerchantRepositoryTestSuite) TestDeleteAllMerchantPermanent() {
	_, err := s.createSeedMerchant()
	s.Require().NoError(err)
	ctx := context.Background()

	success, err := s.repo.DeleteAllMerchantPermanent(ctx)
	s.NoError(err)
	s.True(success)
}

func (s *MerchantRepositoryTestSuite) TestDuplicateApiKeyConflict() {
	ctx := context.Background()

	// Insert a merchant directly via GORM to control the api_key
	seed := models.Merchant{
		UserID: int32(s.userID),
		ApiKey: "duplicate-api-key-12345",
		Name:   "First Merchant",
	}
	var err error
	_, err = s.mongoDB.Collection("merchants").InsertOne(ctx, seed)
	s.Require().NoError(err)

	// Second insert with same api_key must violate unique constraint
	dup := models.Merchant{
		UserID: int32(s.userID),
		ApiKey: "duplicate-api-key-12345",
		Name:   "Second Merchant",
	}
	_, err = s.mongoDB.Collection("merchants").InsertOne(ctx, dup)
	s.Require().Error(err, "duplicate api_key must violate the unique constraint")

	appErr := sharederrors.ErrConstraintOrFailed(err, "Merchant", "create merchant")
	s.Require().NotNil(appErr)
	s.Equal(http.StatusConflict, appErr.Code, "23505 must map to 409")
	s.Contains(appErr.Message, "already exists")
}

func TestMerchantRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(MerchantRepositoryTestSuite))
}
