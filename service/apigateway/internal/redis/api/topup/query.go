package topup_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type topupQueryCache struct {
	store *cache.CacheStore
}

func NewTopupQueryCache(store *cache.CacheStore) TopupQueryCache {
	return &topupQueryCache{store: store}
}

func (c *topupQueryCache) GetCachedTopupsCache(ctx context.Context, req *model.FindAllTopupInput) (*model.APIResponsePaginationTopup, bool) {
	key := fmt.Sprintf(topupAllCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	result, found := cache.GetFromCache[model.APIResponsePaginationTopup](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *topupQueryCache) GetCacheTopupByCardCache(ctx context.Context, req *model.FindAllTopupByCardNumberInput) (*model.APIResponsePaginationTopup, bool) {
	key := fmt.Sprintf(topupByCardCacheKey, req.CardNumber, *req.Page, *req.PageSize, safeString(req.Search))
	result, found := cache.GetFromCache[model.APIResponsePaginationTopup](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *topupQueryCache) GetCachedTopupActiveCache(ctx context.Context, req *model.FindAllTopupInput) (*model.APIResponsePaginationTopupDeleteAt, bool) {
	key := fmt.Sprintf(topupActiveCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	result, found := cache.GetFromCache[model.APIResponsePaginationTopupDeleteAt](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *topupQueryCache) GetCachedTopupTrashedCache(ctx context.Context, req *model.FindAllTopupInput) (*model.APIResponsePaginationTopupDeleteAt, bool) {
	key := fmt.Sprintf(topupTrashedCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	result, found := cache.GetFromCache[model.APIResponsePaginationTopupDeleteAt](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *topupQueryCache) GetCachedTopupCache(ctx context.Context, id int) (*model.APIResponseTopup, bool) {
	key := fmt.Sprintf(topupByIdCacheKey, id)
	result, found := cache.GetFromCache[model.APIResponseTopup](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *topupQueryCache) SetCachedTopupsCache(ctx context.Context, req *model.FindAllTopupInput, data *model.APIResponsePaginationTopup) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(topupAllCacheKey, *req.Page, *req.PageSize, safeString(req.Search))

	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *topupQueryCache) SetCacheTopupByCardCache(ctx context.Context, req *model.FindAllTopupByCardNumberInput, data *model.APIResponsePaginationTopup) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(topupByCardCacheKey, req.CardNumber, *req.Page, *req.PageSize, safeString(req.Search))
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *topupQueryCache) SetCachedTopupActiveCache(ctx context.Context, req *model.FindAllTopupInput, data *model.APIResponsePaginationTopupDeleteAt) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(topupActiveCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *topupQueryCache) SetCachedTopupTrashedCache(ctx context.Context, req *model.FindAllTopupInput, data *model.APIResponsePaginationTopupDeleteAt) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(topupTrashedCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *topupQueryCache) SetCachedTopupCache(ctx context.Context, data *model.APIResponseTopup) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(topupByIdCacheKey, data.Data.ID)
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
