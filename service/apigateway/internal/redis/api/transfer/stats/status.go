package transfer_stats_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type transferStatsStatusCache struct {
	store *cache.CacheStore
}

func NewTransferStatsStatusCache(store *cache.CacheStore) TransferStatsStatusCache {
	return &transferStatsStatusCache{store: store}
}

func (t *transferStatsStatusCache) GetCachedMonthTransferStatusSuccess(ctx context.Context, req *model.FindMonthlyTransferStatusInput) (*model.APIResponseTransferMonthStatusSuccess, bool) {
	key := fmt.Sprintf(transferMonthTransferStatusSuccessKey, req.Month, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTransferMonthStatusSuccess](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transferStatsStatusCache) SetCachedMonthTransferStatusSuccess(ctx context.Context, req *model.FindMonthlyTransferStatusInput, data *model.APIResponseTransferMonthStatusSuccess) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(transferMonthTransferStatusSuccessKey, req.Month, req.Year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transferStatsStatusCache) GetCachedYearlyTransferStatusSuccess(ctx context.Context, year int) (*model.APIResponseTransferYearStatusSuccess, bool) {
	key := fmt.Sprintf(transferYearTransferStatusSuccessKey, year)
	result, found := cache.GetFromCache[model.APIResponseTransferYearStatusSuccess](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transferStatsStatusCache) SetCachedYearlyTransferStatusSuccess(ctx context.Context, year int, data *model.APIResponseTransferYearStatusSuccess) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(transferYearTransferStatusSuccessKey, year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transferStatsStatusCache) GetCachedMonthTransferStatusFailed(ctx context.Context, req *model.FindMonthlyTransferStatusInput) (*model.APIResponseTransferMonthStatusFailed, bool) {
	key := fmt.Sprintf(transferMonthTransferStatusFailedKey, req.Month, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTransferMonthStatusFailed](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transferStatsStatusCache) SetCachedMonthTransferStatusFailed(ctx context.Context, req *model.FindMonthlyTransferStatusInput, data *model.APIResponseTransferMonthStatusFailed) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(transferMonthTransferStatusFailedKey, req.Month, req.Year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transferStatsStatusCache) GetCachedYearlyTransferStatusFailed(ctx context.Context, year int) (*model.APIResponseTransferYearStatusFailed, bool) {
	key := fmt.Sprintf(transferYearTransferStatusFailedKey, year)
	result, found := cache.GetFromCache[model.APIResponseTransferYearStatusFailed](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transferStatsStatusCache) SetCachedYearlyTransferStatusFailed(ctx context.Context, year int, data *model.APIResponseTransferYearStatusFailed) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(transferYearTransferStatusFailedKey, year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}
