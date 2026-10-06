package merchant_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type merchantQueryCache struct {
	store *cache.CacheStore
}

func NewMerchantQueryCache(store *cache.CacheStore) MerchantQueryCache {
	return &merchantQueryCache{store: store}
}

func (m *merchantQueryCache) GetCachedMerchants(ctx context.Context, req *model.FindAllMerchantInput) (*model.APIResponsePaginationMerchant, bool) {
	key := fmt.Sprintf(merchantAllCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	result, found := cache.GetFromCache[model.APIResponsePaginationMerchant](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (m *merchantQueryCache) GetCachedMerchantActive(ctx context.Context, req *model.FindAllMerchantInput) (*model.APIResponsePaginationMerchantDeleteAt, bool) {
	key := fmt.Sprintf(merchantActiveCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	result, found := cache.GetFromCache[model.APIResponsePaginationMerchantDeleteAt](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (m *merchantQueryCache) GetCachedMerchantTrashed(ctx context.Context, req *model.FindAllMerchantInput) (*model.APIResponsePaginationMerchantDeleteAt, bool) {
	key := fmt.Sprintf(merchantTrashedCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	result, found := cache.GetFromCache[model.APIResponsePaginationMerchantDeleteAt](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (m *merchantQueryCache) GetCachedMerchant(ctx context.Context, id int) (*model.APIResponseMerchant, bool) {
	key := fmt.Sprintf(merchantByIdCacheKey, id)
	result, found := cache.GetFromCache[model.APIResponseMerchant](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (m *merchantQueryCache) GetCachedMerchantsByUserId(ctx context.Context, userId int) (*model.APIResponsesMerchant, bool) {
	key := fmt.Sprintf(merchantByUserIdCacheKey, userId)
	result, found := cache.GetFromCache[model.APIResponsesMerchant](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (m *merchantQueryCache) GetCachedMerchantByApiKey(ctx context.Context, apiKey string) (*model.APIResponseMerchant, bool) {
	key := fmt.Sprintf(merchantByApiKeyCacheKey, apiKey)
	result, found := cache.GetFromCache[model.APIResponseMerchant](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (m *merchantQueryCache) SetCachedMerchants(ctx context.Context, req *model.FindAllMerchantInput, data *model.APIResponsePaginationMerchant) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantAllCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	cache.SetToCache(ctx, m.store, key, data, ttlDefault)
}

func (m *merchantQueryCache) SetCachedMerchantActive(ctx context.Context, req *model.FindAllMerchantInput, data *model.APIResponsePaginationMerchantDeleteAt) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantActiveCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	cache.SetToCache(ctx, m.store, key, data, ttlDefault)
}

func (m *merchantQueryCache) SetCachedMerchantTrashed(ctx context.Context, req *model.FindAllMerchantInput, data *model.APIResponsePaginationMerchantDeleteAt) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantTrashedCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	cache.SetToCache(ctx, m.store, key, data, ttlDefault)
}

func (m *merchantQueryCache) SetCachedMerchant(ctx context.Context, data *model.APIResponseMerchant) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantByIdCacheKey, data.Data.ID)
	cache.SetToCache(ctx, m.store, key, data, ttlDefault)
}

func (m *merchantQueryCache) SetCachedMerchantsByUserId(ctx context.Context, userId int, data *model.APIResponsesMerchant) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantByUserIdCacheKey, userId)
	cache.SetToCache(ctx, m.store, key, data, ttlDefault)
}

func (m *merchantQueryCache) SetCachedMerchantByApiKey(ctx context.Context, apiKey string, data *model.APIResponseMerchant) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantByApiKeyCacheKey, apiKey)
	cache.SetToCache(ctx, m.store, key, data, ttlDefault)
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
