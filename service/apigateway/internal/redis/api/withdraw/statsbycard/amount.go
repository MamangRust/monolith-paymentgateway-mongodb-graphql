package withdraw_stats_bycard_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type withdrawStatsByCardAmountCache struct {
	store *cache.CacheStore
}

func NewWithdrawStatsByCardAmountCache(store *cache.CacheStore) WithdrawStatsByCardAmountCache {
	return &withdrawStatsByCardAmountCache{store: store}
}

func (w *withdrawStatsByCardAmountCache) GetCachedMonthlyWithdrawsByCardNumber(ctx context.Context, req *model.FindYearWithdrawCardNumberInput) (*model.APIResponseWithdrawMonthAmount, bool) {
	key := fmt.Sprintf(monthWithdrawAmountByCardKey, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseWithdrawMonthAmount](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (w *withdrawStatsByCardAmountCache) SetCachedMonthlyWithdrawsByCardNumber(ctx context.Context, req *model.FindYearWithdrawCardNumberInput, data *model.APIResponseWithdrawMonthAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(monthWithdrawAmountByCardKey, req.CardNumber, req.Year)
	cache.SetToCache(ctx, w.store, key, data, ttlDefault)
}

func (w *withdrawStatsByCardAmountCache) GetCachedYearlyWithdrawsByCardNumber(ctx context.Context, req *model.FindYearWithdrawCardNumberInput) (*model.APIResponseWithdrawYearAmount, bool) {
	key := fmt.Sprintf(yearWithdrawAmountByCardKey, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseWithdrawYearAmount](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (w *withdrawStatsByCardAmountCache) SetCachedYearlyWithdrawsByCardNumber(ctx context.Context, req *model.FindYearWithdrawCardNumberInput, data *model.APIResponseWithdrawYearAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(yearWithdrawAmountByCardKey, req.CardNumber, req.Year)
	cache.SetToCache(ctx, w.store, key, data, ttlDefault)
}
