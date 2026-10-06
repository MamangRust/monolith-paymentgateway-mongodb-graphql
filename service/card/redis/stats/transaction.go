package cardstatsmencache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type cardStatsTransactionCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewCardStatsTransactionCache(store *sharedcachehelpers.CacheStore) CardStatsTransactionCache {
	return &cardStatsTransactionCache{store: store}
}

func (c *cardStatsTransactionCache) GetMonthlyTransactionCache(ctx context.Context, year int) ([]*salemongo.MonthlyTransactionRow, bool) {
	key := fmt.Sprintf(cacheKeyMonthlyTransactionAmount, year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MonthlyTransactionRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *cardStatsTransactionCache) SetMonthlyTransactionCache(ctx context.Context, year int, data []*salemongo.MonthlyTransactionRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(cacheKeyMonthlyTransactionAmount, year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, ttlStatistic)
}

func (c *cardStatsTransactionCache) GetYearlyTransactionCache(ctx context.Context, year int) ([]*salemongo.YearlyTransactionRow, bool) {
	key := fmt.Sprintf(cacheKeyYearlyTransactionAmount, year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.YearlyTransactionRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *cardStatsTransactionCache) SetYearlyTransactionCache(ctx context.Context, year int, data []*salemongo.YearlyTransactionRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(cacheKeyYearlyTransactionAmount, year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, ttlStatistic)
}
