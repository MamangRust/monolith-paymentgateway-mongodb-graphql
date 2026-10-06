package service

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-transaction/repository"
)

// TransactionQueryService handles queries related to transactions.
type TransactionQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllTransactions) ([]*repository.TransactionQueryResult, *int, error)
	FindAllByCardNumber(ctx context.Context, req *requests.FindAllTransactionCardNumber) ([]*repository.TransactionByCardResult, *int, error)
	FindById(ctx context.Context, transactionID int) (*salemongo.Transaction, error)
	FindByActive(ctx context.Context, req *requests.FindAllTransactions) ([]*repository.TransactionQueryResult, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllTransactions) ([]*repository.TransactionQueryResult, *int, error)
	FindTransactionByMerchantId(ctx context.Context, merchant_id int) ([]*salemongo.Transaction, error)
}

type TransactionCommandService interface {
	Create(ctx context.Context, apiKey string, request *requests.CreateTransactionRequest) (*salemongo.Transaction, error)
	Update(ctx context.Context, apiKey string, request *requests.UpdateTransactionRequest) (*salemongo.Transaction, error)
	TrashedTransaction(ctx context.Context, transaction_id int) (*salemongo.Transaction, error)
	RestoreTransaction(ctx context.Context, transaction_id int) (*salemongo.Transaction, error)
	DeleteTransactionPermanent(ctx context.Context, transaction_id int) (bool, error)

	RestoreAllTransaction(ctx context.Context) (bool, error)
	DeleteAllTransactionPermanent(ctx context.Context) (bool, error)
	RecoverStuckOperations(ctx context.Context, olderThan time.Duration, maxRows int32) error
	StartRecoveryWorker(ctx context.Context, interval, olderThan time.Duration, maxRows int32)

	AuthorizeTransaction(ctx context.Context, apiKey string, request *requests.CreateTransactionRequest) (*salemongo.Transaction, error)
	CaptureTransaction(ctx context.Context, apiKey string, transactionID int) (*salemongo.Transaction, error)
	VoidTransaction(ctx context.Context, apiKey string, transactionID int) (*salemongo.Transaction, error)
	RefundTransaction(ctx context.Context, apiKey string, transactionID int) (*salemongo.Transaction, error)
}
