package saldostatscache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type saldoStatsBalanceCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewSaldoStatsBalanceCache(store *sharedcachehelpers.CacheStore) SaldoStatsBalanceCache {
	return &saldoStatsBalanceCache{store: store}
}

func (c *saldoStatsBalanceCache) GetMonthlySaldoBalanceCache(ctx context.Context, year int) ([]*salemongo.MonthlySaldoBalanceRow, bool) {
	key := fmt.Sprintf(saldoMonthBalanceCacheKey, year)
	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MonthlySaldoBalanceRow](ctx, c.store, key)
	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (c *saldoStatsBalanceCache) SetMonthlySaldoBalanceCache(ctx context.Context, year int, data []*salemongo.MonthlySaldoBalanceRow) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(saldoMonthBalanceCacheKey, year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, ttlDefault)
}

func (c *saldoStatsBalanceCache) GetYearlySaldoBalanceCache(ctx context.Context, year int) ([]*salemongo.YearlySaldoBalanceRow, bool) {
	key := fmt.Sprintf(saldoYearlyBalanceCacheKey, year)
	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.YearlySaldoBalanceRow](ctx, c.store, key)
	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (c *saldoStatsBalanceCache) SetYearlySaldoBalanceCache(ctx context.Context, year int, data []*salemongo.YearlySaldoBalanceRow) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(saldoYearlyBalanceCacheKey, year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, ttlDefault)
}
