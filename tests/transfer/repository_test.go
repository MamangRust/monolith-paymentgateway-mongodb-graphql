package transfer_test

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
	"time"

	card_repo "github.com/MamangRust/monolith-payment-gateway-card/repository"
	saldo_repo "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/MamangRust/monolith-payment-gateway-transfer/repository"
	user_repo "github.com/MamangRust/monolith-payment-gateway-user/repository"

	"github.com/stretchr/testify/suite"
)

type TransferRepositoryTestSuite struct {
	suite.Suite
	ts       *tests.TestSuite
	mongoDB  *mongo.Database
	repo     repository.Repositories
	cardRepo *card_repo.Repositories
	userRepo user_repo.Repositories
	userID   int
	seedSeq  int
}

func (s *TransferRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	mongoDB, err := s.ts.MongoDB()
	if err != nil {
		s.Require().NoError(err)
	}
	s.mongoDB = mongoDB
	s.userRepo = user_repo.NewRepositories(mongoDB, nil, nil)
	s.cardRepo = card_repo.NewRepositories(mongoDB, nil)
	s.repo = repository.NewRepositoriesWithAdapters(mongoDB, nil, nil)

	// Create user
	user, err := s.userRepo.UserCommand().CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Transfer",
		LastName:  "Tester",
		Email:     fmt.Sprintf("transfer.tester-%d@example.com", time.Now().UnixNano()),
		Password:  "password123",
	})
	s.Require().NoError(err)
	s.userID = int(user.UserID)
}

func (s *TransferRepositoryTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *TransferRepositoryTestSuite) createSeedTransfer() (*repository.TransferAtomicResult, error) {
	fromCard, err := s.cardRepo.CardCommand.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID:       s.userID,
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(5, 0, 0),
		CVV:          "111",
		CardProvider: "Visa",
	})
	if err != nil {
		return nil, err
	}

	toCard, err := s.cardRepo.CardCommand.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID:       s.userID,
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(5, 0, 0),
		CVV:          "222",
		CardProvider: "MasterCard",
	})
	if err != nil {
		return nil, err
	}

	// Seed saldo for both cards
	saldoRepo := saldo_repo.NewSaldoCommandRepository(s.mongoDB.Collection("saldos"))
	if _, err := saldoRepo.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{CardNumber: fromCard.CardNumber, TotalBalance: 1000000}); err != nil {
		return nil, err
	}
	if _, err := saldoRepo.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{CardNumber: toCard.CardNumber, TotalBalance: 0}); err != nil {
		return nil, err
	}

	s.seedSeq++
	idemKey := fmt.Sprintf("seed-transfer-%d", s.seedSeq)
	return s.repo.CreateTransferAtomic(context.Background(), &requests.CreateTransferRequest{
		TransferFrom:   fromCard.CardNumber,
		TransferTo:     toCard.CardNumber,
		TransferAmount: 100000,
		IdempotencyKey: idemKey,
	})
}

func (s *TransferRepositoryTestSuite) TestCreateTransfer() {
	ctx := context.Background()

	fromCard, _ := s.cardRepo.CardCommand.CreateCard(ctx, &requests.CreateCardRequest{
		UserID:       s.userID,
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(5, 0, 0),
		CVV:          "111",
		CardProvider: "Visa",
	})

	toCard, _ := s.cardRepo.CardCommand.CreateCard(ctx, &requests.CreateCardRequest{
		UserID:       s.userID,
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(5, 0, 0),
		CVV:          "222",
		CardProvider: "MasterCard",
	})

	saldoRepo := saldo_repo.NewSaldoCommandRepository(s.mongoDB.Collection("saldos"))
	saldoRepo.CreateSaldo(ctx, &requests.CreateSaldoRequest{CardNumber: fromCard.CardNumber, TotalBalance: 1000000})
	saldoRepo.CreateSaldo(ctx, &requests.CreateSaldoRequest{CardNumber: toCard.CardNumber, TotalBalance: 0})

	s.seedSeq++
	req := &requests.CreateTransferRequest{
		TransferFrom:   fromCard.CardNumber,
		TransferTo:     toCard.CardNumber,
		TransferAmount: 100000,
		IdempotencyKey: fmt.Sprintf("test-create-%d", s.seedSeq),
	}

	res, err := s.repo.CreateTransferAtomic(ctx, req)
	s.NoError(err)
	s.NotNil(res)
}

func (s *TransferRepositoryTestSuite) TestFindAllTransfers() {
	_, err := s.createSeedTransfer()
	s.Require().NoError(err)
	ctx := context.Background()

	res, err := s.repo.FindAll(ctx, &requests.FindAllTransfers{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *TransferRepositoryTestSuite) TestFindById() {
	transfer, err := s.createSeedTransfer()
	s.Require().NoError(err)
	ctx := context.Background()

	found, err := s.repo.FindById(ctx, int(transfer.Row.TransferID))
	s.NoError(err)
	s.NotNil(found)
	s.Equal(transfer.Row.TransferID, found.TransferID)
}

func (s *TransferRepositoryTestSuite) TestFindByActive() {
	_, err := s.createSeedTransfer()
	s.Require().NoError(err)
	ctx := context.Background()

	res, err := s.repo.FindByActive(ctx, &requests.FindAllTransfers{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *TransferRepositoryTestSuite) TestFindByTrashed() {
	transfer, err := s.createSeedTransfer()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.TrashedTransfer(ctx, int(transfer.Row.TransferID))
	s.Require().NoError(err)

	res, err := s.repo.FindByTrashed(ctx, &requests.FindAllTransfers{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *TransferRepositoryTestSuite) TestUpdateTransfer() {
	transfer, err := s.createSeedTransfer()
	s.Require().NoError(err)
	ctx := context.Background()

	id := int(transfer.Row.TransferID)
	req := &requests.UpdateTransferRequest{
		TransferID:     &id,
		TransferFrom:   transfer.Row.TransferFrom,
		TransferTo:     transfer.Row.TransferTo,
		TransferAmount: 200000,
	}

	res, err := s.repo.UpdateTransfer(ctx, req)
	s.NoError(err)
	s.NotNil(res)
}

func (s *TransferRepositoryTestSuite) TestTrashTransfer() {
	transfer, err := s.createSeedTransfer()
	s.Require().NoError(err)
	ctx := context.Background()

	trashed, err := s.repo.TrashedTransfer(ctx, int(transfer.Row.TransferID))
	s.NoError(err)
	s.NotNil(trashed)
}

func (s *TransferRepositoryTestSuite) TestRestoreTransfer() {
	transfer, err := s.createSeedTransfer()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.TrashedTransfer(ctx, int(transfer.Row.TransferID))
	s.Require().NoError(err)

	restored, err := s.repo.RestoreTransfer(ctx, int(transfer.Row.TransferID))
	s.NoError(err)
	s.NotNil(restored)
}

func (s *TransferRepositoryTestSuite) TestDeleteTransferPermanent() {
	transfer, err := s.createSeedTransfer()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.TrashedTransfer(ctx, int(transfer.Row.TransferID))
	s.Require().NoError(err)

	success, err := s.repo.DeleteTransferPermanent(ctx, int(transfer.Row.TransferID))
	s.NoError(err)
	s.True(success)
}

func (s *TransferRepositoryTestSuite) TestRestoreAllTransfer() {
	transfer, err := s.createSeedTransfer()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.TrashedTransfer(ctx, int(transfer.Row.TransferID))
	s.Require().NoError(err)

	success, err := s.repo.RestoreAllTransfer(ctx)
	s.NoError(err)
	s.True(success)
}

func (s *TransferRepositoryTestSuite) TestDeleteAllTransferPermanent() {
	_, err := s.createSeedTransfer()
	s.Require().NoError(err)
	ctx := context.Background()

	success, err := s.repo.DeleteAllTransferPermanent(ctx)
	s.NoError(err)
	s.True(success)
}

func TestTransferRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransferRepositoryTestSuite))
}
