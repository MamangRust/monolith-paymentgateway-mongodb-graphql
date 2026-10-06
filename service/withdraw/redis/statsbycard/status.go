package withdrawstatsbycardcache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type withdrawStatsByCardStatusCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewWithdrawStatsStatusCache(store *sharedcachehelpers.CacheStore) WithdrawStatsByCardStatusCache {
	return &withdrawStatsByCardStatusCache{store: store}
}

func (w *withdrawStatsByCardStatusCache) GetCachedMonthWithdrawStatusSuccessByCardNumber(ctx context.Context, req *requests.MonthStatusWithdrawCardNumber) ([]*salemongo.WithdrawMonthlyStatusSuccessByCardRow, bool) {
	key := fmt.Sprintf(monthWithdrawStatusSuccessByCardKey, req.CardNumber, req.Month, req.Year)
	result, found := cache.GetFromCache[[]*salemongo.WithdrawMonthlyStatusSuccessByCardRow](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (w *withdrawStatsByCardStatusCache) SetCachedMonthWithdrawStatusSuccessByCardNumber(ctx context.Context, req *requests.MonthStatusWithdrawCardNumber, data []*salemongo.WithdrawMonthlyStatusSuccessByCardRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(monthWithdrawStatusSuccessByCardKey, req.CardNumber, req.Month, req.Year)
	cache.SetToCache(ctx, w.store, key, &data, ttlDefault)
}

func (w *withdrawStatsByCardStatusCache) GetCachedYearlyWithdrawStatusSuccessByCardNumber(ctx context.Context, req *requests.YearStatusWithdrawCardNumber) ([]*salemongo.WithdrawYearlyStatusSuccessByCardRow, bool) {
	key := fmt.Sprintf(yearWithdrawStatusSuccessByCardKey, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[[]*salemongo.WithdrawYearlyStatusSuccessByCardRow](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (w *withdrawStatsByCardStatusCache) SetCachedYearlyWithdrawStatusSuccessByCardNumber(ctx context.Context, req *requests.YearStatusWithdrawCardNumber, data []*salemongo.WithdrawYearlyStatusSuccessByCardRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(yearWithdrawStatusSuccessByCardKey, req.CardNumber, req.Year)
	cache.SetToCache(ctx, w.store, key, &data, ttlDefault)
}

func (w *withdrawStatsByCardStatusCache) GetCachedMonthWithdrawStatusFailedByCardNumber(ctx context.Context, req *requests.MonthStatusWithdrawCardNumber) ([]*salemongo.WithdrawMonthlyStatusFailedByCardRow, bool) {
	key := fmt.Sprintf(monthWithdrawStatusFailedByCardKey, req.CardNumber, req.Month, req.Year)
	result, found := cache.GetFromCache[[]*salemongo.WithdrawMonthlyStatusFailedByCardRow](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (w *withdrawStatsByCardStatusCache) SetCachedMonthWithdrawStatusFailedByCardNumber(ctx context.Context, req *requests.MonthStatusWithdrawCardNumber, data []*salemongo.WithdrawMonthlyStatusFailedByCardRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(monthWithdrawStatusFailedByCardKey, req.CardNumber, req.Month, req.Year)
	cache.SetToCache(ctx, w.store, key, &data, ttlDefault)
}

func (w *withdrawStatsByCardStatusCache) GetCachedYearlyWithdrawStatusFailedByCardNumber(ctx context.Context, req *requests.YearStatusWithdrawCardNumber) ([]*salemongo.WithdrawYearlyStatusFailedByCardRow, bool) {
	key := fmt.Sprintf(yearWithdrawStatusFailedByCardKey, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[[]*salemongo.WithdrawYearlyStatusFailedByCardRow](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (w *withdrawStatsByCardStatusCache) SetCachedYearlyWithdrawStatusFailedByCardNumber(ctx context.Context, req *requests.YearStatusWithdrawCardNumber, data []*salemongo.WithdrawYearlyStatusFailedByCardRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(yearWithdrawStatusFailedByCardKey, req.CardNumber, req.Year)
	cache.SetToCache(ctx, w.store, key, &data, ttlDefault)
}
