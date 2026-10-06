package merchant_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type merchantTransactionCache struct {
	store *cache.CacheStore
}

func NewMerchantTransactionCache(store *cache.CacheStore) MerchantTransactionCache {
	return &merchantTransactionCache{store: store}
}

func (m *merchantTransactionCache) GetCacheAllMerchantTransactions(ctx context.Context, req *model.FindAllMerchantInput) (*model.APIResponsePaginationMerchantTransaction, bool) {
	key := fmt.Sprintf(merchantTransactionsCacheKey, safeString(req.Search), *req.Page, *req.PageSize)
	result, found := cache.GetFromCache[model.APIResponsePaginationMerchantTransaction](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (m *merchantTransactionCache) GetCacheMerchantTransactions(ctx context.Context, req *model.FindAllMerchantTransactionInput) (*model.APIResponsePaginationMerchantTransaction, bool) {
	key := fmt.Sprintf(merchantTransactionCacheKey, req.MerchantID, safeString(req.Search), *req.Page, *req.PageSize)
	result, found := cache.GetFromCache[model.APIResponsePaginationMerchantTransaction](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (m *merchantTransactionCache) GetCacheMerchantTransactionApikey(ctx context.Context, req *model.FindAllMerchantApikeyInput) (*model.APIResponsePaginationMerchantTransaction, bool) {
	key := fmt.Sprintf(merchantTransactionApikeyCacheKey, req.APIKey, safeString(req.Search), *req.Page, *req.PageSize)
	result, found := cache.GetFromCache[model.APIResponsePaginationMerchantTransaction](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (m *merchantTransactionCache) SetCacheAllMerchantTransactions(ctx context.Context, req *model.FindAllMerchantInput, data *model.APIResponsePaginationMerchantTransaction) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantTransactionsCacheKey, safeString(req.Search), *req.Page, *req.PageSize)
	cache.SetToCache(ctx, m.store, key, data, ttlDefault)
}

func (m *merchantTransactionCache) SetCacheMerchantTransactions(ctx context.Context, req *model.FindAllMerchantTransactionInput, data *model.APIResponsePaginationMerchantTransaction) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantTransactionCacheKey, req.MerchantID, safeString(req.Search), *req.Page, *req.PageSize)
	cache.SetToCache(ctx, m.store, key, data, ttlDefault)
}

func (m *merchantTransactionCache) SetCacheMerchantTransactionApikey(ctx context.Context, req *model.FindAllMerchantApikeyInput, data *model.APIResponsePaginationMerchantTransaction) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantTransactionApikeyCacheKey, req.APIKey, safeString(req.Search), *req.Page, *req.PageSize)
	cache.SetToCache(ctx, m.store, key, data, ttlDefault)
}
