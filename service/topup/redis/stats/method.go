package topupstatscache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type topupStatsMethodCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewTopupStatsMethodCache(store *sharedcachehelpers.CacheStore) TopupStatsMethodCache {
	return &topupStatsMethodCache{store: store}
}

func (c *topupStatsMethodCache) GetMonthlyTopupMethodsCache(ctx context.Context, year int) ([]*salemongo.TopupMonthlyMethodRow, bool) {
	key := fmt.Sprintf(monthTopupMethodCacheKey, year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TopupMonthlyMethodRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (c *topupStatsMethodCache) SetMonthlyTopupMethodsCache(ctx context.Context, year int, data []*salemongo.TopupMonthlyMethodRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(monthTopupMethodCacheKey, year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, ttlDefault)
}

func (c *topupStatsMethodCache) GetYearlyTopupMethodsCache(ctx context.Context, year int) ([]*salemongo.TopupYearlyMethodRow, bool) {
	key := fmt.Sprintf(yearTopupMethodCacheKey, year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TopupYearlyMethodRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (c *topupStatsMethodCache) SetYearlyTopupMethodsCache(ctx context.Context, year int, data []*salemongo.TopupYearlyMethodRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(yearTopupMethodCacheKey, year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, ttlDefault)
}
