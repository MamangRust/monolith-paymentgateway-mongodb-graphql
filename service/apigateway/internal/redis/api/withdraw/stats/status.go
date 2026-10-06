package withdraw_stats_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type withdrawStatsStatusCache struct {
	store *cache.CacheStore
}

func NewWithdrawStatsStatusCache(store *cache.CacheStore) WithdrawStatsStatusCache {
	return &withdrawStatsStatusCache{store: store}
}

func (w *withdrawStatsStatusCache) GetCachedMonthWithdrawStatusSuccessCache(ctx context.Context, req *model.FindMonthlyWithdrawStatusInput) (*model.APIResponseWithdrawMonthStatusSuccess, bool) {
	key := fmt.Sprintf(montWithdrawStatusSuccessKey, req.Month, req.Year)
	result, found := cache.GetFromCache[model.APIResponseWithdrawMonthStatusSuccess](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (w *withdrawStatsStatusCache) SetCachedMonthWithdrawStatusSuccessCache(ctx context.Context, req *model.FindMonthlyWithdrawStatusInput, data *model.APIResponseWithdrawMonthStatusSuccess) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(montWithdrawStatusSuccessKey, req.Month, req.Year)
	cache.SetToCache(ctx, w.store, key, data, ttlDefault)
}

func (w *withdrawStatsStatusCache) GetCachedYearlyWithdrawStatusSuccessCache(ctx context.Context, year int) (*model.APIResponseWithdrawYearStatusSuccess, bool) {
	key := fmt.Sprintf(yearWithdrawStatusSuccessKey, year)
	result, found := cache.GetFromCache[model.APIResponseWithdrawYearStatusSuccess](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (w *withdrawStatsStatusCache) SetCachedYearlyWithdrawStatusSuccessCache(ctx context.Context, year int, data *model.APIResponseWithdrawYearStatusSuccess) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(yearWithdrawStatusSuccessKey, year)
	cache.SetToCache(ctx, w.store, key, data, ttlDefault)
}

func (w *withdrawStatsStatusCache) GetCachedMonthWithdrawStatusFailedCache(ctx context.Context, req *model.FindMonthlyWithdrawStatusInput) (*model.APIResponseWithdrawMonthStatusFailed, bool) {
	key := fmt.Sprintf(montWithdrawStatusFailedKey, req.Month, req.Year)
	result, found := cache.GetFromCache[model.APIResponseWithdrawMonthStatusFailed](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (w *withdrawStatsStatusCache) SetCachedMonthWithdrawStatusFailedCache(ctx context.Context, req *model.FindMonthlyWithdrawStatusInput, data *model.APIResponseWithdrawMonthStatusFailed) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(montWithdrawStatusFailedKey, req.Month, req.Year)
	cache.SetToCache(ctx, w.store, key, data, ttlDefault)
}

func (w *withdrawStatsStatusCache) GetCachedYearlyWithdrawStatusFailedCache(ctx context.Context, year int) (*model.APIResponseWithdrawYearStatusFailed, bool) {
	key := fmt.Sprintf(yearWithdrawStatusFailedKey, year)
	result, found := cache.GetFromCache[model.APIResponseWithdrawYearStatusFailed](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (w *withdrawStatsStatusCache) SetCachedYearlyWithdrawStatusFailedCache(ctx context.Context, year int, data *model.APIResponseWithdrawYearStatusFailed) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(yearWithdrawStatusFailedKey, year)
	cache.SetToCache(ctx, w.store, key, data, ttlDefault)
}
