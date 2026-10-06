package card_dashboard_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type cardDashboardCache struct {
	store *cache.CacheStore
}

func NewCardDashboardCache(store *cache.CacheStore) CardDashboardTotalCache {
	return &cardDashboardCache{store: store}
}

func (c *cardDashboardCache) GetDashboardCardCache(ctx context.Context) (*model.APIResponseDashboardCard, bool) {
	result, found := cache.GetFromCache[model.APIResponseDashboardCard](ctx, c.store, cacheKeyDashboardDefault)

	if !found || result == nil {
		return nil, false
	}

	return result, true
}

func (c *cardDashboardCache) SetDashboardCardCache(ctx context.Context, data *model.APIResponseDashboardCard) {
	if data == nil {
		return
	}

	cache.SetToCache(ctx, c.store, cacheKeyDashboardDefault, data, ttlDashboardDefault)
}

func (c *cardDashboardCache) DeleteDashboardCardCache(ctx context.Context) {
	cache.DeleteFromCache(ctx, c.store, cacheKeyDashboardDefault)
}
