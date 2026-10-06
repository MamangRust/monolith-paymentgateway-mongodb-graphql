package cardstatsmencache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type cardStatsTopupCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewCardStatsTopupCache(store *sharedcachehelpers.CacheStore) CardStatsTopupCache {
	return &cardStatsTopupCache{store: store}
}

func (c *cardStatsTopupCache) GetMonthlyTopupCache(ctx context.Context, year int) ([]*salemongo.MonthlyTopupRow, bool) {
	key := fmt.Sprintf(cacheKeyMonthlyTopupAmount, year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MonthlyTopupRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *cardStatsTopupCache) SetMonthlyTopupCache(ctx context.Context, year int, data []*salemongo.MonthlyTopupRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(cacheKeyMonthlyTopupAmount, year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, ttlStatistic)
}

func (c *cardStatsTopupCache) GetYearlyTopupCache(ctx context.Context, year int) ([]*salemongo.YearlyTopupRow, bool) {
	key := fmt.Sprintf(cacheKeyYearlyTopupAmount, year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.YearlyTopupRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *cardStatsTopupCache) SetYearlyTopupCache(ctx context.Context, year int, data []*salemongo.YearlyTopupRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(cacheKeyYearlyTopupAmount, year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, ttlStatistic)
}
