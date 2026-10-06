package transaction_stats_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type transactionStatsMethodCache struct {
	store *cache.CacheStore
}

func NewTransactionStatsMethodCache(store *cache.CacheStore) TransactionStatsMethodCache {
	return &transactionStatsMethodCache{store: store}
}

func (t *transactionStatsMethodCache) GetMonthlyPaymentMethodsCache(ctx context.Context, year int) (*model.APIResponseTransactionMonthMethod, bool) {
	key := fmt.Sprintf(monthTransactionMethodCacheKey, year)
	result, found := cache.GetFromCache[model.APIResponseTransactionMonthMethod](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transactionStatsMethodCache) SetMonthlyPaymentMethodsCache(ctx context.Context, year int, data *model.APIResponseTransactionMonthMethod) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(monthTransactionMethodCacheKey, year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transactionStatsMethodCache) GetYearlyPaymentMethodsCache(ctx context.Context, year int) (*model.APIResponseTransactionYearMethod, bool) {
	key := fmt.Sprintf(yearTransactionMethodCacheKey, year)
	result, found := cache.GetFromCache[model.APIResponseTransactionYearMethod](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transactionStatsMethodCache) SetYearlyPaymentMethodsCache(ctx context.Context, year int, data *model.APIResponseTransactionYearMethod) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(yearTransactionMethodCacheKey, year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}
