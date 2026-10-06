package transactionstatscache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type transactionStatsAmountCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewTransactionStatsAmountCache(store *sharedcachehelpers.CacheStore) TransactionStatsAmountCache {
	return &transactionStatsAmountCache{store: store}
}

func (t *transactionStatsAmountCache) GetMonthlyAmountsCache(ctx context.Context, year int) ([]*salemongo.TransactionMonthlyAmountRow, bool) {
	key := fmt.Sprintf(monthTransactionAmountCacheKey, year)
	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TransactionMonthlyAmountRow](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (t *transactionStatsAmountCache) SetMonthlyAmountsCache(ctx context.Context, year int, data []*salemongo.TransactionMonthlyAmountRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(monthTransactionAmountCacheKey, year)
	sharedcachehelpers.SetToCache(ctx, t.store, key, &data, ttlDefault)
}

func (t *transactionStatsAmountCache) GetYearlyAmountsCache(ctx context.Context, year int) ([]*salemongo.TransactionYearlyAmountRow, bool) {
	key := fmt.Sprintf(yearTransactionAmountCacheKey, year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TransactionYearlyAmountRow](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (t *transactionStatsAmountCache) SetYearlyAmountsCache(ctx context.Context, year int, data []*salemongo.TransactionYearlyAmountRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(yearTransactionAmountCacheKey, year)
	sharedcachehelpers.SetToCache(ctx, t.store, key, &data, ttlDefault)
}
