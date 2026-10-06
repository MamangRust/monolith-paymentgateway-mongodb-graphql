package transaction_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
)

type TransactionQueryCache interface {
	GetCachedTransactionsCache(ctx context.Context, req *model.FindAllTransactionInput) (*model.APIResponsePaginationTransaction, bool)
	SetCachedTransactionsCache(ctx context.Context, req *model.FindAllTransactionInput, data *model.APIResponsePaginationTransaction)

	GetCachedTransactionByCardNumberCache(ctx context.Context, req *model.FindAllTransactionCardNumberInput) (*model.APIResponsePaginationTransaction, bool)
	SetCachedTransactionByCardNumberCache(ctx context.Context, req *model.FindAllTransactionCardNumberInput, data *model.APIResponsePaginationTransaction)

	GetCachedTransactionActiveCache(ctx context.Context, req *model.FindAllTransactionInput) (*model.APIResponsePaginationTransactionDeleteAt, bool)
	SetCachedTransactionActiveCache(ctx context.Context, req *model.FindAllTransactionInput, data *model.APIResponsePaginationTransactionDeleteAt)

	GetCachedTransactionTrashedCache(ctx context.Context, req *model.FindAllTransactionInput) (*model.APIResponsePaginationTransactionDeleteAt, bool)
	SetCachedTransactionTrashedCache(ctx context.Context, req *model.FindAllTransactionInput, data *model.APIResponsePaginationTransactionDeleteAt)

	GetCachedTransactionByMerchantIdCache(ctx context.Context, merchant_id int) (*model.APIResponseTransactions, bool)
	SetCachedTransactionByMerchantIdCache(ctx context.Context, merchant_id int, data *model.APIResponseTransactions)

	GetCachedTransactionCache(ctx context.Context, id int) (*model.APIResponseTransaction, bool)
	SetCachedTransactionCache(ctx context.Context, data *model.APIResponseTransaction)
}

type TransactionCommandCache interface {
	DeleteTransactionCache(ctx context.Context, id int)
}
