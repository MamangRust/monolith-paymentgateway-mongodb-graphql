package transaction_stats_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type transactionStatsStatusCache struct {
	store *cache.CacheStore
}

func NewTransactionStatsStatusCache(store *cache.CacheStore) TransactionStatsStatusCache {
	return &transactionStatsStatusCache{store: store}
}

func (t *transactionStatsStatusCache) GetMonthTransactionStatusSuccessCache(ctx context.Context, req *requests.MonthStatusTransaction) (*model.APIResponseTransactionMonthStatusSuccess, bool) {
	key := fmt.Sprintf(monthTransactionStatusSuccessCacheKey, req.Month, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTransactionMonthStatusSuccess](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transactionStatsStatusCache) SetMonthTransactionStatusSuccessCache(ctx context.Context, req *requests.MonthStatusTransaction, data *model.APIResponseTransactionMonthStatusSuccess) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(monthTransactionStatusSuccessCacheKey, req.Month, req.Year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transactionStatsStatusCache) GetYearTransactionStatusSuccessCache(ctx context.Context, year int) (*model.APIResponseTransactionYearStatusSuccess, bool) {
	key := fmt.Sprintf(yearTransactionStatusSuccessCacheKey, year)
	result, found := cache.GetFromCache[model.APIResponseTransactionYearStatusSuccess](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transactionStatsStatusCache) SetYearTransactionStatusSuccessCache(ctx context.Context, year int, data *model.APIResponseTransactionYearStatusSuccess) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(yearTransactionStatusSuccessCacheKey, year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transactionStatsStatusCache) GetMonthTransactionStatusFailedCache(ctx context.Context, req *requests.MonthStatusTransaction) (*model.APIResponseTransactionMonthStatusFailed, bool) {
	key := fmt.Sprintf(monthTransactionStatusFailedCacheKey, req.Month, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTransactionMonthStatusFailed](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transactionStatsStatusCache) SetMonthTransactionStatusFailedCache(ctx context.Context, req *requests.MonthStatusTransaction, data *model.APIResponseTransactionMonthStatusFailed) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(monthTransactionStatusFailedCacheKey, req.Month, req.Year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transactionStatsStatusCache) GetYearTransactionStatusFailedCache(ctx context.Context, year int) (*model.APIResponseTransactionYearStatusFailed, bool) {
	key := fmt.Sprintf(yearTransactionStatusFailedCacheKey, year)
	result, found := cache.GetFromCache[model.APIResponseTransactionYearStatusFailed](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transactionStatsStatusCache) SetYearTransactionStatusFailedCache(ctx context.Context, year int, data *model.APIResponseTransactionYearStatusFailed) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(yearTransactionStatusFailedCacheKey, year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}
