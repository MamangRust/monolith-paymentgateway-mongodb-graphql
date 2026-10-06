package topup_stats_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type topupStatsStatusCache struct {
	store *cache.CacheStore
}

func NewTopupStatsStatusCache(store *cache.CacheStore) TopupStatsStatusCache {
	return &topupStatsStatusCache{store: store}
}

func (c *topupStatsStatusCache) GetMonthTopupStatusSuccessCache(ctx context.Context, req *requests.MonthTopupStatus) (*model.APIResponseTopupMonthStatusSuccess, bool) {
	key := fmt.Sprintf(monthTopupStatusSuccessCacheKey, req.Month, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTopupMonthStatusSuccess](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *topupStatsStatusCache) SetMonthTopupStatusSuccessCache(ctx context.Context, req *requests.MonthTopupStatus, data *model.APIResponseTopupMonthStatusSuccess) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(monthTopupStatusSuccessCacheKey, req.Month, req.Year)
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *topupStatsStatusCache) GetYearlyTopupStatusSuccessCache(ctx context.Context, year int) (*model.APIResponseTopupYearStatusSuccess, bool) {
	key := fmt.Sprintf(yearTopupStatusSuccessCacheKey, year)
	result, found := cache.GetFromCache[model.APIResponseTopupYearStatusSuccess](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *topupStatsStatusCache) SetYearlyTopupStatusSuccessCache(ctx context.Context, year int, data *model.APIResponseTopupYearStatusSuccess) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(yearTopupStatusSuccessCacheKey, year)
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *topupStatsStatusCache) GetMonthTopupStatusFailedCache(ctx context.Context, req *requests.MonthTopupStatus) (*model.APIResponseTopupMonthStatusFailed, bool) {
	key := fmt.Sprintf(monthTopupStatusFailedCacheKey, req.Month, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTopupMonthStatusFailed](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *topupStatsStatusCache) SetMonthTopupStatusFailedCache(ctx context.Context, req *requests.MonthTopupStatus, data *model.APIResponseTopupMonthStatusFailed) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(monthTopupStatusFailedCacheKey, req.Month, req.Year)
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *topupStatsStatusCache) GetYearlyTopupStatusFailedCache(ctx context.Context, year int) (*model.APIResponseTopupYearStatusFailed, bool) {
	key := fmt.Sprintf(yearTopupStatusFailedCacheKey, year)
	result, found := cache.GetFromCache[model.APIResponseTopupYearStatusFailed](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *topupStatsStatusCache) SetYearlyTopupStatusFailedCache(ctx context.Context, year int, data *model.APIResponseTopupYearStatusFailed) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(yearTopupStatusFailedCacheKey, year)
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}
