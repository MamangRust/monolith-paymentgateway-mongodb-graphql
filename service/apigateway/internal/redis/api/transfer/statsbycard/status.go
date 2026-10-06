package transfer_stats_bycard_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-shared/cache"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
)

type transferStatsByCardStatusCache struct {
	store *cache.CacheStore
}

func NewTransferStatsByCardStatusCache(store *cache.CacheStore) TransferStatsByCardStatusCache {
	return &transferStatsByCardStatusCache{store: store}
}

func (t *transferStatsByCardStatusCache) GetMonthTransferStatusSuccessByCard(ctx context.Context, req *model.FindMonthlyTransferStatusCardNumberInput) (*model.APIResponseTransferMonthStatusSuccess, bool) {
	key := fmt.Sprintf(transferMonthTransferStatusSuccessByCardKey, req.CardNumber, req.Month, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTransferMonthStatusSuccess](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transferStatsByCardStatusCache) SetMonthTransferStatusSuccessByCard(ctx context.Context, req *model.FindMonthlyTransferStatusCardNumberInput, data *model.APIResponseTransferMonthStatusSuccess) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(transferMonthTransferStatusSuccessByCardKey, req.CardNumber, req.Month, req.Year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transferStatsByCardStatusCache) GetYearlyTransferStatusSuccessByCard(ctx context.Context, req *model.FindYearTransferStatusCardNumberInput) (*model.APIResponseTransferYearStatusSuccess, bool) {
	key := fmt.Sprintf(transferYearTransferStatusSuccessByCardKey, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTransferYearStatusSuccess](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transferStatsByCardStatusCache) SetYearlyTransferStatusSuccessByCard(ctx context.Context, req *model.FindYearTransferStatusCardNumberInput, data *model.APIResponseTransferYearStatusSuccess) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(transferYearTransferStatusSuccessByCardKey, req.CardNumber, req.Year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transferStatsByCardStatusCache) GetMonthTransferStatusFailedByCard(ctx context.Context, req *model.FindMonthlyTransferStatusCardNumberInput) (*model.APIResponseTransferMonthStatusFailed, bool) {
	key := fmt.Sprintf(transferMonthTransferStatusFailedByCardKey, req.CardNumber, req.Month, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTransferMonthStatusFailed](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transferStatsByCardStatusCache) SetMonthTransferStatusFailedByCard(ctx context.Context, req *model.FindMonthlyTransferStatusCardNumberInput, data *model.APIResponseTransferMonthStatusFailed) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(transferMonthTransferStatusFailedByCardKey, req.CardNumber, req.Month, req.Year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transferStatsByCardStatusCache) GetYearlyTransferStatusFailedByCard(ctx context.Context, req *model.FindYearTransferStatusCardNumberInput) (*model.APIResponseTransferYearStatusFailed, bool) {
	key := fmt.Sprintf(transferYearTransferStatusFailedByCardKey, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTransferYearStatusFailed](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transferStatsByCardStatusCache) SetYearlyTransferStatusFailedByCard(ctx context.Context, req *model.FindYearTransferStatusCardNumberInput, data *model.APIResponseTransferYearStatusFailed) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(transferYearTransferStatusFailedByCardKey, req.CardNumber, req.Year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}
