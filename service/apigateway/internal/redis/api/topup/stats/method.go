package topup_stats_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type topupStatsMethodCache struct {
	store *cache.CacheStore
}

func NewTopupStatsMethodCache(store *cache.CacheStore) TopupStatsMethodCache {
	return &topupStatsMethodCache{store: store}
}

func (c *topupStatsMethodCache) GetMonthlyTopupMethodsCache(ctx context.Context, year int) (*model.APIResponseTopupMonthMethod, bool) {
	key := fmt.Sprintf(monthTopupMethodCacheKey, year)
	result, found := cache.GetFromCache[model.APIResponseTopupMonthMethod](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *topupStatsMethodCache) SetMonthlyTopupMethodsCache(ctx context.Context, year int, data *model.APIResponseTopupMonthMethod) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(monthTopupMethodCacheKey, year)
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *topupStatsMethodCache) GetYearlyTopupMethodsCache(ctx context.Context, year int) (*model.APIResponseTopupYearMethod, bool) {
	key := fmt.Sprintf(yearTopupMethodCacheKey, year)
	result, found := cache.GetFromCache[model.APIResponseTopupYearMethod](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *topupStatsMethodCache) SetYearlyTopupMethodsCache(ctx context.Context, year int, data *model.APIResponseTopupYearMethod) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(yearTopupMethodCacheKey, year)
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}
