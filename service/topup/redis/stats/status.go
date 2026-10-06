package topupstatscache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type topupStatsStatusCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewTopupStatsStatusCache(store *sharedcachehelpers.CacheStore) TopupStatsStatusCache {
	return &topupStatsStatusCache{store: store}
}

func (c *topupStatsStatusCache) GetMonthTopupStatusSuccessCache(ctx context.Context, req *requests.MonthTopupStatus) ([]*salemongo.TopupMonthlyStatusRow, bool) {
	key := fmt.Sprintf(monthTopupStatusSuccessCacheKey, req.Month, req.Year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TopupMonthlyStatusRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *topupStatsStatusCache) SetMonthTopupStatusSuccessCache(ctx context.Context, req *requests.MonthTopupStatus, data []*salemongo.TopupMonthlyStatusRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(monthTopupStatusSuccessCacheKey, req.Month, req.Year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, ttlDefault)
}

func (c *topupStatsStatusCache) GetYearlyTopupStatusSuccessCache(ctx context.Context, year int) ([]*salemongo.TopupYearlyStatusRow, bool) {
	key := fmt.Sprintf(yearTopupStatusSuccessCacheKey, year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TopupYearlyStatusRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *topupStatsStatusCache) SetYearlyTopupStatusSuccessCache(ctx context.Context, year int, data []*salemongo.TopupYearlyStatusRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(yearTopupStatusSuccessCacheKey, year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, ttlDefault)
}

func (c *topupStatsStatusCache) GetMonthTopupStatusFailedCache(ctx context.Context, req *requests.MonthTopupStatus) ([]*salemongo.TopupMonthlyStatusRow, bool) {
	key := fmt.Sprintf(monthTopupStatusFailedCacheKey, req.Month, req.Year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TopupMonthlyStatusRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *topupStatsStatusCache) SetMonthTopupStatusFailedCache(ctx context.Context, req *requests.MonthTopupStatus, data []*salemongo.TopupMonthlyStatusRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(monthTopupStatusFailedCacheKey, req.Month, req.Year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, ttlDefault)
}

func (c *topupStatsStatusCache) GetYearlyTopupStatusFailedCache(ctx context.Context, year int) ([]*salemongo.TopupYearlyStatusRow, bool) {
	key := fmt.Sprintf(yearTopupStatusFailedCacheKey, year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TopupYearlyStatusRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (c *topupStatsStatusCache) SetYearlyTopupStatusFailedCache(ctx context.Context, year int, data []*salemongo.TopupYearlyStatusRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(yearTopupStatusFailedCacheKey, year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, ttlDefault)
}
