package saldostatscache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type saldoStatsTotalCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewSaldoStatsTotalCache(store *sharedcachehelpers.CacheStore) SaldoStatsTotalCache {
	return &saldoStatsTotalCache{store: store}
}

func (c *saldoStatsTotalCache) GetMonthlyTotalSaldoBalanceCache(ctx context.Context, req *requests.MonthTotalSaldoBalance) ([]*salemongo.MonthlyTotalSaldoBalanceRow, bool) {
	key := fmt.Sprintf(saldoMonthTotalBalanceCacheKey, req.Month, req.Year)
	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MonthlyTotalSaldoBalanceRow](ctx, c.store, key)
	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (c *saldoStatsTotalCache) SetMonthlyTotalSaldoCache(ctx context.Context, req *requests.MonthTotalSaldoBalance, data []*salemongo.MonthlyTotalSaldoBalanceRow) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(saldoMonthTotalBalanceCacheKey, req.Month, req.Year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, ttlDefault)
}

func (c *saldoStatsTotalCache) GetYearTotalSaldoBalanceCache(ctx context.Context, year int) ([]*salemongo.YearlyTotalSaldoBalancesRow, bool) {
	key := fmt.Sprintf(saldoYearTotalBalanceCacheKey, year)
	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.YearlyTotalSaldoBalancesRow](ctx, c.store, key)
	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (c *saldoStatsTotalCache) SetYearTotalSaldoBalanceCache(ctx context.Context, year int, data []*salemongo.YearlyTotalSaldoBalancesRow) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(saldoYearTotalBalanceCacheKey, year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, ttlDefault)
}
