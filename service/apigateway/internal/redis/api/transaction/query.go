package transaction_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type transactionQueryCache struct {
	store *cache.CacheStore
}

func NewTransactionQueryCache(store *cache.CacheStore) TransactionQueryCache {
	return &transactionQueryCache{store: store}
}

func (t *transactionQueryCache) GetCachedTransactionsCache(ctx context.Context, req *model.FindAllTransactionInput) (*model.APIResponsePaginationTransaction, bool) {
	key := fmt.Sprintf(transactionAllCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	result, found := cache.GetFromCache[model.APIResponsePaginationTransaction](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transactionQueryCache) GetCachedTransactionByCardNumberCache(ctx context.Context, req *model.FindAllTransactionCardNumberInput) (*model.APIResponsePaginationTransaction, bool) {
	key := fmt.Sprintf(transactionByCardCacheKey, req.CardNumber, *req.Page, *req.PageSize, safeString(req.Search))
	result, found := cache.GetFromCache[model.APIResponsePaginationTransaction](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transactionQueryCache) GetCachedTransactionActiveCache(ctx context.Context, req *model.FindAllTransactionInput) (*model.APIResponsePaginationTransactionDeleteAt, bool) {
	key := fmt.Sprintf(transactionActiveCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	result, found := cache.GetFromCache[model.APIResponsePaginationTransactionDeleteAt](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transactionQueryCache) GetCachedTransactionTrashedCache(ctx context.Context, req *model.FindAllTransactionInput) (*model.APIResponsePaginationTransactionDeleteAt, bool) {
	key := fmt.Sprintf(transactionTrashedCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	result, found := cache.GetFromCache[model.APIResponsePaginationTransactionDeleteAt](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transactionQueryCache) GetCachedTransactionByMerchantIdCache(ctx context.Context, merchantId int) (*model.APIResponseTransactions, bool) {
	key := fmt.Sprintf(transactionByMerchantIdCacheKey, merchantId)
	result, found := cache.GetFromCache[model.APIResponseTransactions](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transactionQueryCache) GetCachedTransactionCache(ctx context.Context, transactionId int) (*model.APIResponseTransaction, bool) {
	key := fmt.Sprintf(transactionByIdCacheKey, transactionId)
	result, found := cache.GetFromCache[model.APIResponseTransaction](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transactionQueryCache) SetCachedTransactionsCache(ctx context.Context, req *model.FindAllTransactionInput, data *model.APIResponsePaginationTransaction) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(transactionAllCacheKey, *req.Page, *req.PageSize, safeString(req.Search))

	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transactionQueryCache) SetCachedTransactionByCardNumberCache(ctx context.Context, req *model.FindAllTransactionCardNumberInput, data *model.APIResponsePaginationTransaction) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(transactionByCardCacheKey, req.CardNumber, *req.Page, *req.PageSize, safeString(req.Search))
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transactionQueryCache) SetCachedTransactionActiveCache(ctx context.Context, req *model.FindAllTransactionInput, data *model.APIResponsePaginationTransactionDeleteAt) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(transactionActiveCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transactionQueryCache) SetCachedTransactionTrashedCache(ctx context.Context, req *model.FindAllTransactionInput, data *model.APIResponsePaginationTransactionDeleteAt) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(transactionTrashedCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transactionQueryCache) SetCachedTransactionByMerchantIdCache(ctx context.Context, merchantId int, data *model.APIResponseTransactions) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(transactionByMerchantIdCacheKey, merchantId)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transactionQueryCache) SetCachedTransactionCache(ctx context.Context, data *model.APIResponseTransaction) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(transactionByIdCacheKey, data.Data.ID)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
