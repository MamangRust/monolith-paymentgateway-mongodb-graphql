package topup_stats_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type topupStatsAmountCache struct {
	store *cache.CacheStore
}

func NewTopupStatsAmountCache(store *cache.CacheStore) TopupStatsAmountCache {
	return &topupStatsAmountCache{store: store}
}

func (c *topupStatsAmountCache) GetMonthlyTopupAmountsCache(ctx context.Context, year int) (*model.APIResponseTopupMonthAmount, bool) {
	key := fmt.Sprintf(monthTopupAmountCacheKey, year)
	result, found := cache.GetFromCache[model.APIResponseTopupMonthAmount](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *topupStatsAmountCache) SetMonthlyTopupAmountsCache(ctx context.Context, year int, data *model.APIResponseTopupMonthAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(monthTopupAmountCacheKey, year)
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *topupStatsAmountCache) GetYearlyTopupAmountsCache(ctx context.Context, year int) (*model.APIResponseTopupYearAmount, bool) {
	key := fmt.Sprintf(yearTopupAmountCacheKey, year)
	result, found := cache.GetFromCache[model.APIResponseTopupYearAmount](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *topupStatsAmountCache) SetYearlyTopupAmountsCache(ctx context.Context, year int, data *model.APIResponseTopupYearAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(yearTopupAmountCacheKey, year)
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}
