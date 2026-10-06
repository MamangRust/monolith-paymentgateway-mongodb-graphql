package merchant_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
)

type MerchantQueryCache interface {
	GetCachedMerchants(ctx context.Context, req *model.FindAllMerchantInput) (*model.APIResponsePaginationMerchant, bool)
	SetCachedMerchants(ctx context.Context, req *model.FindAllMerchantInput, data *model.APIResponsePaginationMerchant)

	GetCachedMerchantActive(ctx context.Context, req *model.FindAllMerchantInput) (*model.APIResponsePaginationMerchantDeleteAt, bool)
	SetCachedMerchantActive(ctx context.Context, req *model.FindAllMerchantInput, data *model.APIResponsePaginationMerchantDeleteAt)

	GetCachedMerchantTrashed(ctx context.Context, req *model.FindAllMerchantInput) (*model.APIResponsePaginationMerchantDeleteAt, bool)
	SetCachedMerchantTrashed(ctx context.Context, req *model.FindAllMerchantInput, data *model.APIResponsePaginationMerchantDeleteAt)

	GetCachedMerchant(ctx context.Context, id int) (*model.APIResponseMerchant, bool)
	SetCachedMerchant(ctx context.Context, data *model.APIResponseMerchant)

	GetCachedMerchantsByUserId(ctx context.Context, userId int) (*model.APIResponsesMerchant, bool)
	SetCachedMerchantsByUserId(ctx context.Context, userId int, data *model.APIResponsesMerchant)

	GetCachedMerchantByApiKey(ctx context.Context, apiKey string) (*model.APIResponseMerchant, bool)
	SetCachedMerchantByApiKey(ctx context.Context, apiKey string, data *model.APIResponseMerchant)
}

type MerchantTransactionCache interface {
	GetCacheAllMerchantTransactions(ctx context.Context, req *model.FindAllMerchantInput) (*model.APIResponsePaginationMerchantTransaction, bool)
	SetCacheAllMerchantTransactions(ctx context.Context, req *model.FindAllMerchantInput, data *model.APIResponsePaginationMerchantTransaction)

	GetCacheMerchantTransactions(ctx context.Context, req *model.FindAllMerchantTransactionInput) (*model.APIResponsePaginationMerchantTransaction, bool)
	SetCacheMerchantTransactions(ctx context.Context, req *model.FindAllMerchantTransactionInput, data *model.APIResponsePaginationMerchantTransaction)

	GetCacheMerchantTransactionApikey(ctx context.Context, req *model.FindAllMerchantApikeyInput) (*model.APIResponsePaginationMerchantTransaction, bool)
	SetCacheMerchantTransactionApikey(ctx context.Context, req *model.FindAllMerchantApikeyInput, data *model.APIResponsePaginationMerchantTransaction)
}

type MerchantCommandCache interface {
	DeleteCachedMerchant(ctx context.Context, id int)
}
