package transaction_stats_bycard_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type transactionStatsByCardAmountCache struct {
	store *cache.CacheStore
}

func NewTransactionStatsByCardAmountCache(store *cache.CacheStore) TransactionStatsByCardAmountCache {
	return &transactionStatsByCardAmountCache{store: store}
}

func (t *transactionStatsByCardAmountCache) GetMonthlyAmountsByCardCache(ctx context.Context, req *model.FindByYearCardNumberTransactionInput) (*model.APIResponseTransactionMonthAmount, bool) {
	key := fmt.Sprintf(monthTransactionAmountByCardCacheKey, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTransactionMonthAmount](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transactionStatsByCardAmountCache) SetMonthlyAmountsByCardCache(ctx context.Context, req *model.FindByYearCardNumberTransactionInput, data *model.APIResponseTransactionMonthAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(monthTransactionAmountByCardCacheKey, req.CardNumber, req.Year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transactionStatsByCardAmountCache) GetYearlyAmountsByCardCache(ctx context.Context, req *model.FindByYearCardNumberTransactionInput) (*model.APIResponseTransactionYearAmount, bool) {
	key := fmt.Sprintf(yearTransactionAmountByCardCacheKey, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTransactionYearAmount](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transactionStatsByCardAmountCache) SetYearlyAmountsByCardCache(ctx context.Context, req *model.FindByYearCardNumberTransactionInput, data *model.APIResponseTransactionYearAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(yearTransactionAmountByCardCacheKey, req.CardNumber, req.Year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}
