package cardstatsmencache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type cardStatsBalanceCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewCardStatsBalanceCache(store *sharedcachehelpers.CacheStore) CardStatsBalanceCache {
	return &cardStatsBalanceCache{store: store}
}

func (c *cardStatsBalanceCache) GetMonthlyBalanceCache(ctx context.Context, year int) ([]*salemongo.MonthlyBalanceRow, bool) {
	key := fmt.Sprintf(cacheKeyMonthlyBalance, year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MonthlyBalanceRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *cardStatsBalanceCache) SetMonthlyBalanceCache(ctx context.Context, year int, data []*salemongo.MonthlyBalanceRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(cacheKeyMonthlyBalance, year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, ttlStatistic)
}

func (c *cardStatsBalanceCache) GetYearlyBalanceCache(ctx context.Context, year int) ([]*salemongo.YearlyBalanceRow, bool) {
	key := fmt.Sprintf(cacheKeyYearlyBalance, year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.YearlyBalanceRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *cardStatsBalanceCache) SetYearlyBalanceCache(ctx context.Context, year int, data []*salemongo.YearlyBalanceRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(cacheKeyYearlyBalance, year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, ttlStatistic)
}
