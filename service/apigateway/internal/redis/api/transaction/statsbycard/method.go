package transaction_stats_bycard_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type transactionStatsByCardMethodCache struct {
	store *cache.CacheStore
}

func NewTransactionStatsByCardMethodCache(store *cache.CacheStore) TransactionStatsByCardMethodCache {
	return &transactionStatsByCardMethodCache{store: store}
}

func (t *transactionStatsByCardMethodCache) GetMonthlyPaymentMethodsByCardCache(ctx context.Context, req *model.FindByYearCardNumberTransactionInput) (*model.APIResponseTransactionMonthMethod, bool) {
	key := fmt.Sprintf(monthTransactionMethodByCardCacheKey, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTransactionMonthMethod](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transactionStatsByCardMethodCache) SetMonthlyPaymentMethodsByCardCache(ctx context.Context, req *model.FindByYearCardNumberTransactionInput, data *model.APIResponseTransactionMonthMethod) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(monthTransactionMethodByCardCacheKey, req.CardNumber, req.Year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transactionStatsByCardMethodCache) GetYearlyPaymentMethodsByCardCache(ctx context.Context, req *model.FindByYearCardNumberTransactionInput) (*model.APIResponseTransactionYearMethod, bool) {
	key := fmt.Sprintf(yearTransactionMethodByCardCacheKey, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTransactionYearMethod](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transactionStatsByCardMethodCache) SetYearlyPaymentMethodsByCardCache(ctx context.Context, req *model.FindByYearCardNumberTransactionInput, data *model.APIResponseTransactionYearMethod) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(yearTransactionMethodByCardCacheKey, req.CardNumber, req.Year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}
