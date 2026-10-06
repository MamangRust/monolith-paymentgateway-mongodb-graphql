package transferstatscache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type transferStatsStatusCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewTransferStatsStatusCache(store *sharedcachehelpers.CacheStore) TransferStatsStatusCache {
	return &transferStatsStatusCache{store: store}
}

func (t *transferStatsStatusCache) GetCachedMonthTransferStatusSuccess(ctx context.Context, req *requests.MonthStatusTransfer) ([]*salemongo.TransferMonthlyStatusSuccessRow, bool) {
	key := fmt.Sprintf(transferMonthTransferStatusSuccessKey, req.Month, req.Year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TransferMonthlyStatusSuccessRow](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (t *transferStatsStatusCache) SetCachedMonthTransferStatusSuccess(ctx context.Context, req *requests.MonthStatusTransfer, data []*salemongo.TransferMonthlyStatusSuccessRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(transferMonthTransferStatusSuccessKey, req.Month, req.Year)
	sharedcachehelpers.SetToCache(ctx, t.store, key, &data, ttlDefault)
}

func (t *transferStatsStatusCache) GetCachedYearlyTransferStatusSuccess(ctx context.Context, year int) ([]*salemongo.TransferYearlyStatusSuccessRow, bool) {
	key := fmt.Sprintf(transferYearTransferStatusSuccessKey, year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TransferYearlyStatusSuccessRow](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (t *transferStatsStatusCache) SetCachedYearlyTransferStatusSuccess(ctx context.Context, year int, data []*salemongo.TransferYearlyStatusSuccessRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(transferYearTransferStatusSuccessKey, year)
	sharedcachehelpers.SetToCache(ctx, t.store, key, &data, ttlDefault)
}

func (t *transferStatsStatusCache) GetCachedMonthTransferStatusFailed(ctx context.Context, req *requests.MonthStatusTransfer) ([]*salemongo.TransferMonthlyStatusFailedRow, bool) {
	key := fmt.Sprintf(transferMonthTransferStatusFailedKey, req.Month, req.Year)
	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TransferMonthlyStatusFailedRow](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (t *transferStatsStatusCache) SetCachedMonthTransferStatusFailed(ctx context.Context, req *requests.MonthStatusTransfer, data []*salemongo.TransferMonthlyStatusFailedRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(transferMonthTransferStatusFailedKey, req.Month, req.Year)
	sharedcachehelpers.SetToCache(ctx, t.store, key, &data, ttlDefault)
}

func (t *transferStatsStatusCache) GetCachedYearlyTransferStatusFailed(ctx context.Context, year int) ([]*salemongo.TransferYearlyStatusFailedRow, bool) {
	key := fmt.Sprintf(transferYearTransferStatusFailedKey, year)
	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TransferYearlyStatusFailedRow](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (t *transferStatsStatusCache) SetCachedYearlyTransferStatusFailed(ctx context.Context, year int, data []*salemongo.TransferYearlyStatusFailedRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(transferYearTransferStatusFailedKey, year)
	sharedcachehelpers.SetToCache(ctx, t.store, key, &data, ttlDefault)
}
