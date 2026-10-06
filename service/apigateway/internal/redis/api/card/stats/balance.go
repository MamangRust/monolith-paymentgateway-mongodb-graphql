package card_stats_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type cardStatsBalanceCache struct {
	store *cache.CacheStore
}

func NewCardStatsBalanceCache(store *cache.CacheStore) CardStatsBalanceCache {
	return &cardStatsBalanceCache{store: store}
}

func (c *cardStatsBalanceCache) GetMonthlyBalanceCache(ctx context.Context, year int) (*model.APIResponseMonthlyBalance, bool) {
	key := fmt.Sprintf(cacheKeyMonthlyBalance, year)
	result, found := cache.GetFromCache[model.APIResponseMonthlyBalance](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *cardStatsBalanceCache) SetMonthlyBalanceCache(ctx context.Context, year int, data *model.APIResponseMonthlyBalance) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cacheKeyMonthlyBalance, year)
	cache.SetToCache(ctx, c.store, key, data, ttlStatistic)
}

func (c *cardStatsBalanceCache) GetYearlyBalanceCache(ctx context.Context, year int) (*model.APIResponseYearlyBalance, bool) {
	key := fmt.Sprintf(cacheKeyYearlyBalance, year)
	result, found := cache.GetFromCache[model.APIResponseYearlyBalance](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *cardStatsBalanceCache) SetYearlyBalanceCache(ctx context.Context, year int, data *model.APIResponseYearlyBalance) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cacheKeyYearlyBalance, year)
	cache.SetToCache(ctx, c.store, key, data, ttlStatistic)
}
