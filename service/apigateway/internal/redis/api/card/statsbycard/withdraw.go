package card_stats_bycard_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type cardStatsWithdrawByCardCache struct {
	store *cache.CacheStore
}

func NewCardStatsWithdrawByCardCache(store *cache.CacheStore) CardStatsWithdrawByCardCache {
	return &cardStatsWithdrawByCardCache{store: store}
}

func (c *cardStatsWithdrawByCardCache) GetMonthlyWithdrawByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput) (*model.APIResponseMonthlyAmount, bool) {
	key := fmt.Sprintf(cacheKeyMonthlyWithdrawByCard, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseMonthlyAmount](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *cardStatsWithdrawByCardCache) SetMonthlyWithdrawByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput, data *model.APIResponseMonthlyAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cacheKeyMonthlyWithdrawByCard, req.CardNumber, req.Year)
	cache.SetToCache(ctx, c.store, key, data, expirationCardStatistic)
}

func (c *cardStatsWithdrawByCardCache) GetYearlyWithdrawByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput) (*model.APIResponseYearlyAmount, bool) {
	key := fmt.Sprintf(cacheKeyYearlyWithdrawByCard, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseYearlyAmount](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *cardStatsWithdrawByCardCache) SetYearlyWithdrawByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput, data *model.APIResponseYearlyAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cacheKeyYearlyWithdrawByCard, req.CardNumber, req.Year)
	cache.SetToCache(ctx, c.store, key, data, expirationCardStatistic)
}
