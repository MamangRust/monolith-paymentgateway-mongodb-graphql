package withdraw_stats_bycard_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type withdrawStatsByCardStatusCache struct {
	store *cache.CacheStore
}

func NewWithdrawStatsByCardStatusCache(store *cache.CacheStore) WithdrawStatsByCardStatusCache {
	return &withdrawStatsByCardStatusCache{store: store}
}

func (w *withdrawStatsByCardStatusCache) GetCachedMonthWithdrawStatusSuccessByCardNumber(ctx context.Context, req *model.FindMonthlyWithdrawStatusCardNumberInput) (*model.APIResponseWithdrawMonthStatusSuccess, bool) {
	key := fmt.Sprintf(monthWithdrawStatusSuccessByCardKey, req.CardNumber, req.Month, req.Year)
	result, found := cache.GetFromCache[model.APIResponseWithdrawMonthStatusSuccess](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (w *withdrawStatsByCardStatusCache) SetCachedMonthWithdrawStatusSuccessByCardNumber(ctx context.Context, req *model.FindMonthlyWithdrawStatusCardNumberInput, data *model.APIResponseWithdrawMonthStatusSuccess) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(monthWithdrawStatusSuccessByCardKey, req.CardNumber, req.Month, req.Year)
	cache.SetToCache(ctx, w.store, key, data, ttlDefault)
}

func (w *withdrawStatsByCardStatusCache) GetCachedYearlyWithdrawStatusSuccessByCardNumber(ctx context.Context, req *model.FindYearWithdrawStatusCardNumberInput) (*model.APIResponseWithdrawYearStatusSuccess, bool) {
	key := fmt.Sprintf(yearWithdrawStatusSuccessByCardKey, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseWithdrawYearStatusSuccess](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (w *withdrawStatsByCardStatusCache) SetCachedYearlyWithdrawStatusSuccessByCardNumber(ctx context.Context, req *model.FindYearWithdrawStatusCardNumberInput, data *model.APIResponseWithdrawYearStatusSuccess) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(yearWithdrawStatusSuccessByCardKey, req.CardNumber, req.Year)
	cache.SetToCache(ctx, w.store, key, data, ttlDefault)
}

func (w *withdrawStatsByCardStatusCache) GetCachedMonthWithdrawStatusFailedByCardNumber(ctx context.Context, req *model.FindMonthlyWithdrawStatusCardNumberInput) (*model.APIResponseWithdrawMonthStatusFailed, bool) {
	key := fmt.Sprintf(monthWithdrawStatusFailedByCardKey, req.CardNumber, req.Month, req.Year)
	result, found := cache.GetFromCache[model.APIResponseWithdrawMonthStatusFailed](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (w *withdrawStatsByCardStatusCache) SetCachedMonthWithdrawStatusFailedByCardNumber(ctx context.Context, req *model.FindMonthlyWithdrawStatusCardNumberInput, data *model.APIResponseWithdrawMonthStatusFailed) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(monthWithdrawStatusFailedByCardKey, req.CardNumber, req.Month, req.Year)
	cache.SetToCache(ctx, w.store, key, data, ttlDefault)
}

func (w *withdrawStatsByCardStatusCache) GetCachedYearlyWithdrawStatusFailedByCardNumber(ctx context.Context, req *model.FindYearWithdrawStatusCardNumberInput) (*model.APIResponseWithdrawYearStatusFailed, bool) {
	key := fmt.Sprintf(yearWithdrawStatusFailedByCardKey, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseWithdrawYearStatusFailed](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (w *withdrawStatsByCardStatusCache) SetCachedYearlyWithdrawStatusFailedByCardNumber(ctx context.Context, req *model.FindYearWithdrawStatusCardNumberInput, data *model.APIResponseWithdrawYearStatusFailed) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(yearWithdrawStatusFailedByCardKey, req.CardNumber, req.Year)
	cache.SetToCache(ctx, w.store, key, data, ttlDefault)
}
