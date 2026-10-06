package card_stats_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type cardStatsTransactionCache struct {
	store *cache.CacheStore
}

func NewCardStatsTransactionCache(store *cache.CacheStore) CardStatsTransactionCache {
	return &cardStatsTransactionCache{store: store}
}

func (c *cardStatsTransactionCache) GetMonthlyTransactionCache(ctx context.Context, year int) (*model.APIResponseMonthlyAmount, bool) {
	key := fmt.Sprintf(cacheKeyMonthlyTransactionAmount, year)
	result, found := cache.GetFromCache[model.APIResponseMonthlyAmount](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *cardStatsTransactionCache) SetMonthlyTransactionCache(ctx context.Context, year int, data *model.APIResponseMonthlyAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cacheKeyMonthlyTransactionAmount, year)
	cache.SetToCache(ctx, c.store, key, data, ttlStatistic)
}

func (c *cardStatsTransactionCache) GetYearlyTransactionCache(ctx context.Context, year int) (*model.APIResponseYearlyAmount, bool) {
	key := fmt.Sprintf(cacheKeyYearlyTransactionAmount, year)
	result, found := cache.GetFromCache[model.APIResponseYearlyAmount](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *cardStatsTransactionCache) SetYearlyTransactionCache(ctx context.Context, year int, data *model.APIResponseYearlyAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cacheKeyYearlyTransactionAmount, year)
	cache.SetToCache(ctx, c.store, key, data, ttlStatistic)
}
