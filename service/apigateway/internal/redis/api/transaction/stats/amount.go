package transaction_stats_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type transactionStatsAmountCache struct {
	store *cache.CacheStore
}

func NewTransactionStatsAmountCache(store *cache.CacheStore) TransactionStatsAmountCache {
	return &transactionStatsAmountCache{store: store}
}

func (t *transactionStatsAmountCache) GetMonthlyAmountsCache(ctx context.Context, year int) (*model.APIResponseTransactionMonthAmount, bool) {
	key := fmt.Sprintf(monthTransactionAmountCacheKey, year)
	result, found := cache.GetFromCache[model.APIResponseTransactionMonthAmount](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transactionStatsAmountCache) SetMonthlyAmountsCache(ctx context.Context, year int, data *model.APIResponseTransactionMonthAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(monthTransactionAmountCacheKey, year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transactionStatsAmountCache) GetYearlyAmountsCache(ctx context.Context, year int) (*model.APIResponseTransactionYearAmount, bool) {
	key := fmt.Sprintf(yearTransactionAmountCacheKey, year)
	result, found := cache.GetFromCache[model.APIResponseTransactionYearAmount](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transactionStatsAmountCache) SetYearlyAmountsCache(ctx context.Context, year int, data *model.APIResponseTransactionYearAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(yearTransactionAmountCacheKey, year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}
