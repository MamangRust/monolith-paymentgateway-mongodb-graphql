package saldo_stats_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type saldoStatsBalanceCache struct {
	store *cache.CacheStore
}

func NewSaldoStatsBalanceCache(store *cache.CacheStore) SaldoStatsBalanceCache {
	return &saldoStatsBalanceCache{store: store}
}

func (c *saldoStatsBalanceCache) GetMonthlySaldoBalanceCache(ctx context.Context, year int) (*model.APIResponseMonthSaldoBalances, bool) {
	key := fmt.Sprintf(saldoMonthBalanceCacheKey, year)
	result, found := cache.GetFromCache[model.APIResponseMonthSaldoBalances](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *saldoStatsBalanceCache) SetMonthlySaldoBalanceCache(ctx context.Context, year int, data *model.APIResponseMonthSaldoBalances) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(saldoMonthBalanceCacheKey, year)
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *saldoStatsBalanceCache) GetYearlySaldoBalanceCache(ctx context.Context, year int) (*model.APIResponseYearSaldoBalances, bool) {
	key := fmt.Sprintf(saldoYearlyBalanceCacheKey, year)
	result, found := cache.GetFromCache[model.APIResponseYearSaldoBalances](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *saldoStatsBalanceCache) SetYearlySaldoBalanceCache(ctx context.Context, year int, data *model.APIResponseYearSaldoBalances) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(saldoYearlyBalanceCacheKey, year)
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}
