package cardstatsmencache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type cardStatsWithdrawCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewCardStatsWithdrawCache(store *sharedcachehelpers.CacheStore) CardStatsWithdrawCache {
	return &cardStatsWithdrawCache{store: store}
}

func (c *cardStatsWithdrawCache) GetMonthlyWithdrawCache(ctx context.Context, year int) ([]*salemongo.MonthlyWithdrawRow, bool) {
	key := fmt.Sprintf(cacheKeyMonthlyWithdrawAmount, year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MonthlyWithdrawRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *cardStatsWithdrawCache) SetMonthlyWithdrawCache(ctx context.Context, year int, data []*salemongo.MonthlyWithdrawRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(cacheKeyMonthlyWithdrawAmount, year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, ttlStatistic)
}

func (c *cardStatsWithdrawCache) GetYearlyWithdrawCache(ctx context.Context, year int) ([]*salemongo.YearlyWithdrawRow, bool) {
	key := fmt.Sprintf(cacheKeyYearlyWithdrawAmount, year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.YearlyWithdrawRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *cardStatsWithdrawCache) SetYearlyWithdrawCache(ctx context.Context, year int, data []*salemongo.YearlyWithdrawRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(cacheKeyYearlyWithdrawAmount, year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, ttlStatistic)
}
