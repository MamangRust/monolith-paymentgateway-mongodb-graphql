package transaction_test

import (
	"context"
	"fmt"
	models "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"testing"
	"time"

	card_repo "github.com/MamangRust/monolith-payment-gateway-card/repository"
	merchant_repo "github.com/MamangRust/monolith-payment-gateway-merchant/repository"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/MamangRust/monolith-payment-gateway-transaction/repository"
	user_repo "github.com/MamangRust/monolith-payment-gateway-user/repository"

	"github.com/stretchr/testify/suite"
)

type TransactionRepositoryTestSuite struct {
	suite.Suite
	ts           *tests.TestSuite
	repo         repository.Repositories
	cardRepo     *card_repo.Repositories
	merchantRepo merchant_repo.Repositories
	userRepo     user_repo.Repositories
	userID       int
	merchantID   int
}

func (s *TransactionRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	mongoDB, err := s.ts.MongoDB()
	if err != nil {
		s.Require().NoError(err)
	}
	s.userRepo = user_repo.NewRepositories(mongoDB, nil, nil)
	s.cardRepo = card_repo.NewRepositories(mongoDB, nil)
	s.merchantRepo = merchant_repo.NewRepositories(mongoDB, nil)
	s.repo = repository.NewRepositoriesWithAdapters(mongoDB, nil, nil, nil)

	// Create user
	user, err := s.userRepo.UserCommand().CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Transaction",
		LastName:  "Tester",
		Email:     fmt.Sprintf("transaction.tester-%d@example.com", time.Now().UnixNano()),
		Password:  "password123",
	})
	s.Require().NoError(err)
	s.userID = int(user.UserID)

	// Create merchant
	merchant, err := s.merchantRepo.CreateMerchant(context.Background(), &requests.CreateMerchantRequest{
		Name:   "Test Merchant",
		UserID: s.userID,
	})
	s.Require().NoError(err)
	s.merchantID = int(merchant.MerchantID)
}

func (s *TransactionRepositoryTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *TransactionRepositoryTestSuite) createSeedTransaction() (*models.Transaction, error) {
	card, err := s.cardRepo.CardCommand.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID:       s.userID,
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(5, 0, 0),
		CVV:          "123",
		CardProvider: "Visa",
	})
	if err != nil {
		return nil, err
	}

	mid := s.merchantID
	return s.repo.CreateTransaction(context.Background(), &requests.CreateTransactionRequest{
		CardNumber:      card.CardNumber,
		Amount:          100000,
		PaymentMethod:   "debit",
		MerchantID:      &mid,
		TransactionTime: time.Now(),
	})
}

func (s *TransactionRepositoryTestSuite) TestCreateTransaction() {
	ctx := context.Background()

	card, _ := s.cardRepo.CardCommand.CreateCard(ctx, &requests.CreateCardRequest{
		UserID:       s.userID,
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(5, 0, 0),
		CVV:          "123",
		CardProvider: "Visa",
	})

	mid := s.merchantID
	req := &requests.CreateTransactionRequest{
		CardNumber:      card.CardNumber,
		Amount:          100000,
		PaymentMethod:   "debit",
		MerchantID:      &mid,
		TransactionTime: time.Now(),
	}

	res, err := s.repo.CreateTransaction(ctx, req)
	s.NoError(err)
	s.NotNil(res)
}

func (s *TransactionRepositoryTestSuite) TestFindAllTransactions() {
	_, err := s.createSeedTransaction()
	s.Require().NoError(err)
	ctx := context.Background()

	res, err := s.repo.FindAllTransactions(ctx, &requests.FindAllTransactions{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *TransactionRepositoryTestSuite) TestFindById() {
	transaction, err := s.createSeedTransaction()
	s.Require().NoError(err)
	ctx := context.Background()

	found, err := s.repo.FindById(ctx, int(transaction.TransactionID))
	s.NoError(err)
	s.NotNil(found)
	s.Equal(transaction.TransactionID, found.TransactionID)
}

func (s *TransactionRepositoryTestSuite) TestFindByActive() {
	_, err := s.createSeedTransaction()
	s.Require().NoError(err)
	ctx := context.Background()

	res, err := s.repo.FindByActive(ctx, &requests.FindAllTransactions{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *TransactionRepositoryTestSuite) TestFindByTrashed() {
	transaction, err := s.createSeedTransaction()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.TrashedTransaction(ctx, int(transaction.TransactionID))
	s.Require().NoError(err)

	res, err := s.repo.FindByTrashed(ctx, &requests.FindAllTransactions{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *TransactionRepositoryTestSuite) TestUpdateTransaction() {
	transaction, err := s.createSeedTransaction()
	s.Require().NoError(err)
	ctx := context.Background()

	id := int(transaction.TransactionID)
	mid := s.merchantID
	req := &requests.UpdateTransactionRequest{
		TransactionID:   &id,
		CardNumber:      transaction.CardNumber,
		Amount:          200000,
		PaymentMethod:   "credit",
		MerchantID:      &mid,
		TransactionTime: time.Now(),
	}

	res, err := s.repo.UpdateTransaction(ctx, req)
	s.NoError(err)
	s.NotNil(res)
}

func (s *TransactionRepositoryTestSuite) TestTrashTransaction() {
	transaction, err := s.createSeedTransaction()
	s.Require().NoError(err)
	ctx := context.Background()

	trashed, err := s.repo.TrashedTransaction(ctx, int(transaction.TransactionID))
	s.NoError(err)
	s.NotNil(trashed)
}

func (s *TransactionRepositoryTestSuite) TestRestoreTransaction() {
	transaction, err := s.createSeedTransaction()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.TrashedTransaction(ctx, int(transaction.TransactionID))
	s.Require().NoError(err)

	restored, err := s.repo.RestoreTransaction(ctx, int(transaction.TransactionID))
	s.NoError(err)
	s.NotNil(restored)
}

func (s *TransactionRepositoryTestSuite) TestDeleteTransactionPermanent() {
	transaction, err := s.createSeedTransaction()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.TrashedTransaction(ctx, int(transaction.TransactionID))
	s.Require().NoError(err)

	success, err := s.repo.DeleteTransactionPermanent(ctx, int(transaction.TransactionID))
	s.NoError(err)
	s.True(success)
}

func (s *TransactionRepositoryTestSuite) TestRestoreAllTransaction() {
	transaction, err := s.createSeedTransaction()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.TrashedTransaction(ctx, int(transaction.TransactionID))
	s.Require().NoError(err)

	success, err := s.repo.RestoreAllTransaction(ctx)
	s.NoError(err)
	s.True(success)
}

func (s *TransactionRepositoryTestSuite) TestDeleteAllTransactionPermanent() {
	_, err := s.createSeedTransaction()
	s.Require().NoError(err)
	ctx := context.Background()

	success, err := s.repo.DeleteAllTransactionPermanent(ctx)
	s.NoError(err)
	s.True(success)
}

func TestTransactionRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransactionRepositoryTestSuite))
}
