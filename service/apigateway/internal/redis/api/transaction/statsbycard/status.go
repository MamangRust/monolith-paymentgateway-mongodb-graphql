package transaction_stats_bycard_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type transactionStatsByCardStatusCache struct {
	store *cache.CacheStore
}

func NewTransactionStatsByCardStatusCache(store *cache.CacheStore) TransactionStatsByCardStatusCache {
	return &transactionStatsByCardStatusCache{store: store}
}

func (t *transactionStatsByCardStatusCache) GetMonthTransactionStatusSuccessByCardCache(ctx context.Context, req *model.FindMonthlyTransactionStatusCardNumberInput) (*model.APIResponseTransactionMonthStatusSuccess, bool) {
	key := fmt.Sprintf(monthTransactionStatusSuccessByCardCacheKey, req.CardNumber, req.Month, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTransactionMonthStatusSuccess](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transactionStatsByCardStatusCache) SetMonthTransactionStatusSuccessByCardCache(ctx context.Context, req *model.FindMonthlyTransactionStatusCardNumberInput, data *model.APIResponseTransactionMonthStatusSuccess) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(monthTransactionStatusSuccessByCardCacheKey, req.CardNumber, req.Month, req.Year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transactionStatsByCardStatusCache) GetYearTransactionStatusSuccessByCardCache(ctx context.Context, req *model.FindYearTransactionStatusCardNumberInput) (*model.APIResponseTransactionYearStatusSuccess, bool) {
	key := fmt.Sprintf(yearTransactionStatusSuccessByCardCacheKey, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTransactionYearStatusSuccess](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transactionStatsByCardStatusCache) SetYearTransactionStatusSuccessByCardCache(ctx context.Context, req *model.FindYearTransactionStatusCardNumberInput, data *model.APIResponseTransactionYearStatusSuccess) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(yearTransactionStatusSuccessByCardCacheKey, req.CardNumber, req.Year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transactionStatsByCardStatusCache) GetMonthTransactionStatusFailedByCardCache(ctx context.Context, req *model.FindMonthlyTransactionStatusCardNumberInput) (*model.APIResponseTransactionMonthStatusFailed, bool) {
	key := fmt.Sprintf(monthTransactionStatusFailedByCardCacheKey, req.CardNumber, req.Month, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTransactionMonthStatusFailed](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transactionStatsByCardStatusCache) SetMonthTransactionStatusFailedByCardCache(ctx context.Context, req *model.FindMonthlyTransactionStatusCardNumberInput, data *model.APIResponseTransactionMonthStatusFailed) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(monthTransactionStatusFailedByCardCacheKey, req.CardNumber, req.Month, req.Year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transactionStatsByCardStatusCache) GetYearTransactionStatusFailedByCardCache(ctx context.Context, req *model.FindYearTransactionStatusCardNumberInput) (*model.APIResponseTransactionYearStatusFailed, bool) {
	key := fmt.Sprintf(yearTransactionStatusFailedByCardCacheKey, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTransactionYearStatusFailed](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transactionStatsByCardStatusCache) SetYearTransactionStatusFailedByCardCache(ctx context.Context, req *model.FindYearTransactionStatusCardNumberInput, data *model.APIResponseTransactionYearStatusFailed) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(yearTransactionStatusFailedByCardCacheKey, req.CardNumber, req.Year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}
