package withdraw_stats_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type withdrawStatsAmountCache struct {
	store *cache.CacheStore
}

func NewWithdrawStatsAmountCache(store *cache.CacheStore) WithdrawStatsAmountCache {
	return &withdrawStatsAmountCache{store: store}
}

func (w *withdrawStatsAmountCache) GetCachedMonthlyWithdraws(ctx context.Context, year int) (*model.APIResponseWithdrawMonthAmount, bool) {
	key := fmt.Sprintf(montWithdrawAmountKey, year)
	result, found := cache.GetFromCache[model.APIResponseWithdrawMonthAmount](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (w *withdrawStatsAmountCache) SetCachedMonthlyWithdraws(ctx context.Context, year int, data *model.APIResponseWithdrawMonthAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(montWithdrawAmountKey, year)
	cache.SetToCache(ctx, w.store, key, data, ttlDefault)
}

func (w *withdrawStatsAmountCache) GetCachedYearlyWithdraws(ctx context.Context, year int) (*model.APIResponseWithdrawYearAmount, bool) {
	key := fmt.Sprintf(yearWithdrawAmountKey, year)
	result, found := cache.GetFromCache[model.APIResponseWithdrawYearAmount](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (w *withdrawStatsAmountCache) SetCachedYearlyWithdraws(ctx context.Context, year int, data *model.APIResponseWithdrawYearAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(yearWithdrawAmountKey, year)
	cache.SetToCache(ctx, w.store, key, data, ttlDefault)
}
