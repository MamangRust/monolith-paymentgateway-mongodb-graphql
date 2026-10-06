package card_stats_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type cardStatsWithdrawCache struct {
	store *cache.CacheStore
}

func NewCardStatsWithdrawCache(store *cache.CacheStore) CardStatsWithdrawCache {
	return &cardStatsWithdrawCache{store: store}
}

func (c *cardStatsWithdrawCache) GetMonthlyWithdrawCache(ctx context.Context, year int) (*model.APIResponseMonthlyAmount, bool) {
	key := fmt.Sprintf(cacheKeyMonthlyWithdrawAmount, year)
	result, found := cache.GetFromCache[model.APIResponseMonthlyAmount](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *cardStatsWithdrawCache) SetMonthlyWithdrawCache(ctx context.Context, year int, data *model.APIResponseMonthlyAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cacheKeyMonthlyWithdrawAmount, year)
	cache.SetToCache(ctx, c.store, key, data, ttlStatistic)
}

func (c *cardStatsWithdrawCache) GetYearlyWithdrawCache(ctx context.Context, year int) (*model.APIResponseYearlyAmount, bool) {
	key := fmt.Sprintf(cacheKeyYearlyWithdrawAmount, year)
	result, found := cache.GetFromCache[model.APIResponseYearlyAmount](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *cardStatsWithdrawCache) SetYearlyWithdrawCache(ctx context.Context, year int, data *model.APIResponseYearlyAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cacheKeyYearlyWithdrawAmount, year)
	cache.SetToCache(ctx, c.store, key, data, ttlStatistic)
}
