package card_stats_bycard_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type cardStatsTransferByCardCache struct {
	store *cache.CacheStore
}

func NewCardStatsTransferByCardCache(store *cache.CacheStore) CardStatsTransferByCardCache {
	return &cardStatsTransferByCardCache{store: store}
}

func (c *cardStatsTransferByCardCache) GetMonthlyTransferBySenderCache(ctx context.Context, req *model.FindYearCardNumberInput) (*model.APIResponseMonthlyAmount, bool) {
	key := fmt.Sprintf(cacheKeyMonthlySenderByCard, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseMonthlyAmount](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *cardStatsTransferByCardCache) SetMonthlyTransferBySenderCache(ctx context.Context, req *model.FindYearCardNumberInput, data *model.APIResponseMonthlyAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cacheKeyMonthlySenderByCard, req.CardNumber, req.Year)
	cache.SetToCache(ctx, c.store, key, data, expirationCardStatistic)
}

func (c *cardStatsTransferByCardCache) GetYearlyTransferBySenderCache(ctx context.Context, req *model.FindYearCardNumberInput) (*model.APIResponseYearlyAmount, bool) {
	key := fmt.Sprintf(cacheKeyYearlySenderByCard, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseYearlyAmount](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *cardStatsTransferByCardCache) SetYearlyTransferBySenderCache(ctx context.Context, req *model.FindYearCardNumberInput, data *model.APIResponseYearlyAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cacheKeyYearlySenderByCard, req.CardNumber, req.Year)
	cache.SetToCache(ctx, c.store, key, data, expirationCardStatistic)
}

func (c *cardStatsTransferByCardCache) GetMonthlyTransferByReceiverCache(ctx context.Context, req *model.FindYearCardNumberInput) (*model.APIResponseMonthlyAmount, bool) {
	key := fmt.Sprintf(cacheKeyMonthlyReceiverByCard, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseMonthlyAmount](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *cardStatsTransferByCardCache) SetMonthlyTransferByReceiverCache(ctx context.Context, req *model.FindYearCardNumberInput, data *model.APIResponseMonthlyAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cacheKeyMonthlyReceiverByCard, req.CardNumber, req.Year)
	cache.SetToCache(ctx, c.store, key, data, expirationCardStatistic)
}

func (c *cardStatsTransferByCardCache) GetYearlyTransferByReceiverCache(ctx context.Context, req *model.FindYearCardNumberInput) (*model.APIResponseYearlyAmount, bool) {
	key := fmt.Sprintf(cacheKeyYearlyReceiverByCard, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseYearlyAmount](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *cardStatsTransferByCardCache) SetYearlyTransferByReceiverCache(ctx context.Context, req *model.FindYearCardNumberInput, data *model.APIResponseYearlyAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cacheKeyYearlyReceiverByCard, req.CardNumber, req.Year)
	cache.SetToCache(ctx, c.store, key, data, expirationCardStatistic)
}
