package withdrawstatscache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type withdrawStatsStatusCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewWithdrawStatsStatusCache(store *sharedcachehelpers.CacheStore) WithdrawStatsStatusCache {
	return &withdrawStatsStatusCache{store: store}
}

func (w *withdrawStatsStatusCache) GetCachedMonthWithdrawStatusSuccessCache(ctx context.Context, req *requests.MonthStatusWithdraw) ([]*salemongo.WithdrawMonthlyStatusSuccessRow, bool) {
	key := fmt.Sprintf(montWithdrawStatusSuccessKey, req.Month, req.Year)

	result, found := cache.GetFromCache[[]*salemongo.WithdrawMonthlyStatusSuccessRow](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (w *withdrawStatsStatusCache) SetCachedMonthWithdrawStatusSuccessCache(ctx context.Context, req *requests.MonthStatusWithdraw, data []*salemongo.WithdrawMonthlyStatusSuccessRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(montWithdrawStatusSuccessKey, req.Month, req.Year)
	cache.SetToCache(ctx, w.store, key, &data, ttlDefault)
}

func (w *withdrawStatsStatusCache) GetCachedYearlyWithdrawStatusSuccessCache(ctx context.Context, year int) ([]*salemongo.WithdrawYearlyStatusSuccessRow, bool) {
	key := fmt.Sprintf(yearWithdrawStatusSuccessKey, year)
	result, found := cache.GetFromCache[[]*salemongo.WithdrawYearlyStatusSuccessRow](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (w *withdrawStatsStatusCache) SetCachedYearlyWithdrawStatusSuccessCache(ctx context.Context, year int, data []*salemongo.WithdrawYearlyStatusSuccessRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(yearWithdrawStatusSuccessKey, year)
	cache.SetToCache(ctx, w.store, key, &data, ttlDefault)
}

func (w *withdrawStatsStatusCache) GetCachedMonthWithdrawStatusFailedCache(ctx context.Context, req *requests.MonthStatusWithdraw) ([]*salemongo.WithdrawMonthlyStatusFailedRow, bool) {
	key := fmt.Sprintf(montWithdrawStatusFailedKey, req.Month, req.Year)
	result, found := cache.GetFromCache[[]*salemongo.WithdrawMonthlyStatusFailedRow](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (w *withdrawStatsStatusCache) SetCachedMonthWithdrawStatusFailedCache(ctx context.Context, req *requests.MonthStatusWithdraw, data []*salemongo.WithdrawMonthlyStatusFailedRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(montWithdrawStatusFailedKey, req.Month, req.Year)
	cache.SetToCache(ctx, w.store, key, &data, ttlDefault)
}

func (w *withdrawStatsStatusCache) GetCachedYearlyWithdrawStatusFailedCache(ctx context.Context, year int) ([]*salemongo.WithdrawYearlyStatusFailedRow, bool) {
	key := fmt.Sprintf(yearWithdrawStatusFailedKey, year)
	result, found := cache.GetFromCache[[]*salemongo.WithdrawYearlyStatusFailedRow](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (w *withdrawStatsStatusCache) SetCachedYearlyWithdrawStatusFailedCache(ctx context.Context, year int, data []*salemongo.WithdrawYearlyStatusFailedRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(yearWithdrawStatusFailedKey, year)
	cache.SetToCache(ctx, w.store, key, &data, ttlDefault)
}
