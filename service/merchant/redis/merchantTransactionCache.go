package mencache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type merchantTransactionCacheResponse struct {
	Data         []*salemongo.Transaction `json:"data"`
	TotalRecords *int                     `json:"total_records"`
}

type merchantTransactionCache struct {
	store *cache.CacheStore
}

func NewMerchantTransactionCache(store *cache.CacheStore) MerchantTransactionCache {
	return &merchantTransactionCache{store: store}
}

func (m *merchantTransactionCache) SetCacheAllMerchantTransactions(ctx context.Context, req *requests.FindAllMerchantTransactions, data []*salemongo.Transaction, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*salemongo.Transaction{}
	}
	key := fmt.Sprintf(merchantTransactionsCacheKey, req.Search, req.Page, req.PageSize)
	payload := &merchantTransactionCacheResponse{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, m.store, key, payload, ttlDefault)
}

func (m *merchantTransactionCache) GetCacheAllMerchantTransactions(ctx context.Context, req *requests.FindAllMerchantTransactions) ([]*salemongo.Transaction, *int, bool) {
	key := fmt.Sprintf(merchantTransactionsCacheKey, req.Search, req.Page, req.PageSize)
	result, found := cache.GetFromCache[merchantTransactionCacheResponse](ctx, m.store, key)
	if !found || result == nil {
		return nil, nil, false
	}
	return result.Data, result.TotalRecords, true
}

func (m *merchantTransactionCache) SetCacheMerchantTransactions(ctx context.Context, req *requests.FindAllMerchantTransactionsById, data []*salemongo.Transaction, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*salemongo.Transaction{}
	}
	key := fmt.Sprintf(merchantTransactionCacheKey, req.MerchantID, req.Search, req.Page, req.PageSize)
	payload := &merchantTransactionCacheResponse{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, m.store, key, payload, ttlDefault)
}

func (m *merchantTransactionCache) GetCacheMerchantTransactions(ctx context.Context, req *requests.FindAllMerchantTransactionsById) ([]*salemongo.Transaction, *int, bool) {
	key := fmt.Sprintf(merchantTransactionCacheKey, req.MerchantID, req.Search, req.Page, req.PageSize)
	result, found := cache.GetFromCache[merchantTransactionCacheResponse](ctx, m.store, key)
	if !found || result == nil {
		return nil, nil, false
	}
	return result.Data, result.TotalRecords, true
}

func (m *merchantTransactionCache) SetCacheMerchantTransactionApikey(ctx context.Context, req *requests.FindAllMerchantTransactionsByApiKey, data []*salemongo.Transaction, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*salemongo.Transaction{}
	}
	key := fmt.Sprintf(merchantTransactionApikeyCacheKey, req.ApiKey, req.Search, req.Page, req.PageSize)
	payload := &merchantTransactionCacheResponse{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, m.store, key, payload, ttlDefault)
}

func (m *merchantTransactionCache) GetCacheMerchantTransactionApikey(ctx context.Context, req *requests.FindAllMerchantTransactionsByApiKey) ([]*salemongo.Transaction, *int, bool) {
	key := fmt.Sprintf(merchantTransactionApikeyCacheKey, req.ApiKey, req.Search, req.Page, req.PageSize)
	result, found := cache.GetFromCache[merchantTransactionCacheResponse](ctx, m.store, key)
	if !found || result == nil {
		return nil, nil, false
	}
	return result.Data, result.TotalRecords, true
}
