package card_stats_bycard_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type cardStatsBalanceByCardCache struct {
	store *cache.CacheStore
}

func NewCardStatsBalanceByCardCache(store *cache.CacheStore) CardStatsBalanceByCardCache {
	return &cardStatsBalanceByCardCache{store: store}
}

func (c *cardStatsBalanceByCardCache) GetMonthlyBalanceByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput) (*model.APIResponseMonthlyBalance, bool) {
	key := fmt.Sprintf(cacheKeyMonthlyBalanceByCard, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseMonthlyBalance](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *cardStatsBalanceByCardCache) SetMonthlyBalanceByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput, data *model.APIResponseMonthlyBalance) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cacheKeyMonthlyBalanceByCard, req.CardNumber, req.Year)
	cache.SetToCache(ctx, c.store, key, data, expirationCardStatistic)
}

func (c *cardStatsBalanceByCardCache) GetYearlyBalanceByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput) (*model.APIResponseYearlyBalance, bool) {
	key := fmt.Sprintf(cacheKeyYearlyBalanceByCard, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseYearlyBalance](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *cardStatsBalanceByCardCache) SetYearlyBalanceByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput, data *model.APIResponseYearlyBalance) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cacheKeyYearlyBalanceByCard, req.CardNumber, req.Year)
	cache.SetToCache(ctx, c.store, key, data, expirationCardStatistic)
}
