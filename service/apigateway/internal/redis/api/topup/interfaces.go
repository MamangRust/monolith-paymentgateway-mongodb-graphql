package topup_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
)

type TopupQueryCache interface {
	GetCachedTopupsCache(ctx context.Context, req *model.FindAllTopupInput) (*model.APIResponsePaginationTopup, bool)
	SetCachedTopupsCache(ctx context.Context, req *model.FindAllTopupInput, data *model.APIResponsePaginationTopup)

	GetCacheTopupByCardCache(ctx context.Context, req *model.FindAllTopupByCardNumberInput) (*model.APIResponsePaginationTopup, bool)
	SetCacheTopupByCardCache(ctx context.Context, req *model.FindAllTopupByCardNumberInput, data *model.APIResponsePaginationTopup)

	GetCachedTopupActiveCache(ctx context.Context, req *model.FindAllTopupInput) (*model.APIResponsePaginationTopupDeleteAt, bool)
	SetCachedTopupActiveCache(ctx context.Context, req *model.FindAllTopupInput, data *model.APIResponsePaginationTopupDeleteAt)

	GetCachedTopupTrashedCache(ctx context.Context, req *model.FindAllTopupInput) (*model.APIResponsePaginationTopupDeleteAt, bool)
	SetCachedTopupTrashedCache(ctx context.Context, req *model.FindAllTopupInput, data *model.APIResponsePaginationTopupDeleteAt)

	GetCachedTopupCache(ctx context.Context, id int) (*model.APIResponseTopup, bool)
	SetCachedTopupCache(ctx context.Context, data *model.APIResponseTopup)
}

type TopupCommandCache interface {
	DeleteCachedTopupCache(ctx context.Context, id int)
}
