package mencache

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-transaction/repository"
)

type TransactionQueryCache interface {
	GetCachedTransactionsCache(ctx context.Context, req *requests.FindAllTransactions) ([]*repository.TransactionQueryResult, *int, bool)
	SetCachedTransactionsCache(ctx context.Context, req *requests.FindAllTransactions, data []*repository.TransactionQueryResult, total *int)

	GetCachedTransactionByCardNumberCache(ctx context.Context, req *requests.FindAllTransactionCardNumber) ([]*repository.TransactionByCardResult, *int, bool)
	SetCachedTransactionByCardNumberCache(ctx context.Context, req *requests.FindAllTransactionCardNumber, data []*repository.TransactionByCardResult, total *int)

	GetCachedTransactionActiveCache(ctx context.Context, req *requests.FindAllTransactions) ([]*repository.TransactionQueryResult, *int, bool)
	SetCachedTransactionActiveCache(ctx context.Context, req *requests.FindAllTransactions, data []*repository.TransactionQueryResult, total *int)

	GetCachedTransactionTrashedCache(ctx context.Context, req *requests.FindAllTransactions) ([]*repository.TransactionQueryResult, *int, bool)
	SetCachedTransactionTrashedCache(ctx context.Context, req *requests.FindAllTransactions, data []*repository.TransactionQueryResult, total *int)

	GetCachedTransactionByMerchantIdCache(ctx context.Context, merchant_id int) ([]*salemongo.Transaction, bool)
	SetCachedTransactionByMerchantIdCache(ctx context.Context, merchant_id int, data []*salemongo.Transaction)

	GetCachedTransactionCache(ctx context.Context, id int) (*salemongo.Transaction, bool)
	SetCachedTransactionCache(ctx context.Context, data *salemongo.Transaction)
}

type TransactionCommandCache interface {
	DeleteTransactionCache(ctx context.Context, id int)
}
