package transfer_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type transferQueryCache struct {
	store *cache.CacheStore
}

func NewTransferQueryCache(store *cache.CacheStore) TransferQueryCache {
	return &transferQueryCache{store: store}
}

func (c *transferQueryCache) GetCachedTransfersCache(ctx context.Context, req *model.FindAllTransferInput) (*model.APIResponsePaginationTransfer, bool) {
	key := fmt.Sprintf(transferAllCacheKey, *req.Page, *req.PageSize, safeString(req.Search))

	result, found := cache.GetFromCache[model.APIResponsePaginationTransfer](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *transferQueryCache) GetCachedTransferActiveCache(ctx context.Context, req *model.FindAllTransferInput) (*model.APIResponsePaginationTransferDeleteAt, bool) {
	key := fmt.Sprintf(transferActiveCacheKey, *req.Page, *req.PageSize, safeString(req.Search))

	result, found := cache.GetFromCache[model.APIResponsePaginationTransferDeleteAt](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return result, true
}

func (c *transferQueryCache) GetCachedTransferTrashedCache(ctx context.Context, req *model.FindAllTransferInput) (*model.APIResponsePaginationTransferDeleteAt, bool) {
	key := fmt.Sprintf(transferTrashedCacheKey, *req.Page, *req.PageSize, safeString(req.Search))

	result, found := cache.GetFromCache[model.APIResponsePaginationTransferDeleteAt](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return result, true
}

func (c *transferQueryCache) GetCachedTransferCache(ctx context.Context, id int) (*model.APIResponseTransfer, bool) {
	key := fmt.Sprintf(transferByIdCacheKey, id)
	result, found := cache.GetFromCache[model.APIResponseTransfer](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return result, true
}

func (c *transferQueryCache) GetCachedTransferByFrom(ctx context.Context, from string) (*model.APIResponseTransfers, bool) {
	key := fmt.Sprintf(transferByFromCacheKey, from)
	result, found := cache.GetFromCache[model.APIResponseTransfers](ctx, c.store, key)
	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *transferQueryCache) GetCachedTransferByTo(ctx context.Context, to string) (*model.APIResponseTransfers, bool) {
	key := fmt.Sprintf(transferByToCacheKey, to)
	result, found := cache.GetFromCache[model.APIResponseTransfers](ctx, c.store, key)
	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *transferQueryCache) SetCachedTransfersCache(ctx context.Context, req *model.FindAllTransferInput, data *model.APIResponsePaginationTransfer) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(transferAllCacheKey, *req.Page, *req.PageSize, safeString(req.Search))

	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *transferQueryCache) SetCachedTransferActiveCache(ctx context.Context, req *model.FindAllTransferInput, data *model.APIResponsePaginationTransferDeleteAt) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(transferActiveCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *transferQueryCache) SetCachedTransferTrashedCache(ctx context.Context, req *model.FindAllTransferInput, data *model.APIResponsePaginationTransferDeleteAt) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(transferTrashedCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *transferQueryCache) SetCachedTransferCache(ctx context.Context, data *model.APIResponseTransfer) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(transferByIdCacheKey, data.Data.ID)
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *transferQueryCache) SetCachedTransferByFrom(ctx context.Context, from string, data *model.APIResponseTransfers) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(transferByFromCacheKey, from)
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *transferQueryCache) SetCachedTransferByTo(ctx context.Context, to string, data *model.APIResponseTransfers) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(transferByToCacheKey, to)
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
