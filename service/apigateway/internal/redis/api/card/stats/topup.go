package card_stats_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type cardStatsTopupCache struct {
	store *cache.CacheStore
}

func NewCardStatsTopupCache(store *cache.CacheStore) CardStatsTopupCache {
	return &cardStatsTopupCache{store: store}
}

func (c *cardStatsTopupCache) GetMonthlyTopupCache(ctx context.Context, year int) (*model.APIResponseMonthlyAmount, bool) {
	key := fmt.Sprintf(cacheKeyMonthlyTopupAmount, year)
	result, found := cache.GetFromCache[model.APIResponseMonthlyAmount](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *cardStatsTopupCache) SetMonthlyTopupCache(ctx context.Context, year int, data *model.APIResponseMonthlyAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cacheKeyMonthlyTopupAmount, year)
	cache.SetToCache(ctx, c.store, key, data, ttlStatistic)
}

func (c *cardStatsTopupCache) GetYearlyTopupCache(ctx context.Context, year int) (*model.APIResponseYearlyAmount, bool) {
	key := fmt.Sprintf(cacheKeyYearlyTopupAmount, year)
	result, found := cache.GetFromCache[model.APIResponseYearlyAmount](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *cardStatsTopupCache) SetYearlyTopupCache(ctx context.Context, year int, data *model.APIResponseYearlyAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cacheKeyYearlyTopupAmount, year)
	cache.SetToCache(ctx, c.store, key, data, ttlStatistic)
}
