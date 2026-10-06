package saldo_stats_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type saldoStatsTotalCache struct {
	store *cache.CacheStore
}

func NewSaldoStatsTotalCache(store *cache.CacheStore) SaldoStatsTotalCache {
	return &saldoStatsTotalCache{store: store}
}

func (c *saldoStatsTotalCache) GetMonthlyTotalSaldoBalanceCache(ctx context.Context, req *model.FindMonthlySaldoTotalBalanceInput) (*model.APIResponseMonthTotalSaldo, bool) {
	key := fmt.Sprintf(saldoMonthTotalBalanceCacheKey, req.Month, req.Year)
	result, found := cache.GetFromCache[model.APIResponseMonthTotalSaldo](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *saldoStatsTotalCache) SetMonthlyTotalSaldoCache(ctx context.Context, req *model.FindMonthlySaldoTotalBalanceInput, data *model.APIResponseMonthTotalSaldo) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(saldoMonthTotalBalanceCacheKey, req.Month, req.Year)
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *saldoStatsTotalCache) GetYearTotalSaldoBalanceCache(ctx context.Context, year int) (*model.APIResponseYearTotalSaldo, bool) {
	key := fmt.Sprintf(saldoYearTotalBalanceCacheKey, year)
	result, found := cache.GetFromCache[model.APIResponseYearTotalSaldo](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *saldoStatsTotalCache) SetYearTotalSaldoBalanceCache(ctx context.Context, year int, data *model.APIResponseYearTotalSaldo) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(saldoYearTotalBalanceCacheKey, year)
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}
