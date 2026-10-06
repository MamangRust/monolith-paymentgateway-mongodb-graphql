package card_stats_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type cardStatsTransferCache struct {
	store *cache.CacheStore
}

func NewCardStatsTransferCache(store *cache.CacheStore) CardStatsTransferCache {
	return &cardStatsTransferCache{store: store}
}

func (c *cardStatsTransferCache) GetMonthlyTransferSenderCache(ctx context.Context, year int) (*model.APIResponseMonthlyAmount, bool) {
	key := fmt.Sprintf(cacheKeyMonthlyTransferSender, year)
	result, found := cache.GetFromCache[model.APIResponseMonthlyAmount](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *cardStatsTransferCache) SetMonthlyTransferSenderCache(ctx context.Context, year int, data *model.APIResponseMonthlyAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cacheKeyMonthlyTransferSender, year)
	cache.SetToCache(ctx, c.store, key, data, ttlStatistic)
}

func (c *cardStatsTransferCache) GetYearlyTransferSenderCache(ctx context.Context, year int) (*model.APIResponseYearlyAmount, bool) {
	key := fmt.Sprintf(cacheKeyYearlyTransferSender, year)
	result, found := cache.GetFromCache[model.APIResponseYearlyAmount](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *cardStatsTransferCache) SetYearlyTransferSenderCache(ctx context.Context, year int, data *model.APIResponseYearlyAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cacheKeyYearlyTransferSender, year)
	cache.SetToCache(ctx, c.store, key, data, ttlStatistic)
}

func (c *cardStatsTransferCache) GetMonthlyTransferReceiverCache(ctx context.Context, year int) (*model.APIResponseMonthlyAmount, bool) {
	key := fmt.Sprintf(cacheKeyMonthlyTransferReceiver, year)
	result, found := cache.GetFromCache[model.APIResponseMonthlyAmount](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *cardStatsTransferCache) SetMonthlyTransferReceiverCache(ctx context.Context, year int, data *model.APIResponseMonthlyAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cacheKeyMonthlyTransferReceiver, year)
	cache.SetToCache(ctx, c.store, key, data, ttlStatistic)
}

func (c *cardStatsTransferCache) GetYearlyTransferReceiverCache(ctx context.Context, year int) (*model.APIResponseYearlyAmount, bool) {
	key := fmt.Sprintf(cacheKeyYearlyTransferReceiver, year)
	result, found := cache.GetFromCache[model.APIResponseYearlyAmount](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *cardStatsTransferCache) SetYearlyTransferReceiverCache(ctx context.Context, year int, data *model.APIResponseYearlyAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cacheKeyYearlyTransferReceiver, year)
	cache.SetToCache(ctx, c.store, key, data, ttlStatistic)
}
