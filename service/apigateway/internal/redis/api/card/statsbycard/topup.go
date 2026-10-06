package card_stats_bycard_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type cardStatsTopupByCardCache struct {
	store *cache.CacheStore
}

func NewCardStatsTopupByCardCache(store *cache.CacheStore) CardStatsTopupByCardCache {
	return &cardStatsTopupByCardCache{store: store}
}

func (c *cardStatsTopupByCardCache) GetMonthlyTopupByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput) (*model.APIResponseMonthlyAmount, bool) {
	key := fmt.Sprintf(cacheKeyMonthlyTopupByCard, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseMonthlyAmount](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *cardStatsTopupByCardCache) SetMonthlyTopupByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput, data *model.APIResponseMonthlyAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cacheKeyMonthlyTopupByCard, req.CardNumber, req.Year)
	cache.SetToCache(ctx, c.store, key, data, expirationCardStatistic)
}

func (c *cardStatsTopupByCardCache) GetYearlyTopupByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput) (*model.APIResponseYearlyAmount, bool) {
	key := fmt.Sprintf(cacheKeyYearlyTopupByCard, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseYearlyAmount](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *cardStatsTopupByCardCache) SetYearlyTopupByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput, data *model.APIResponseYearlyAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cacheKeyYearlyTopupByCard, req.CardNumber, req.Year)
	cache.SetToCache(ctx, c.store, key, data, expirationCardStatistic)
}
