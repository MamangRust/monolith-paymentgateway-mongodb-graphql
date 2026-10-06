package transactionstatscache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type transactionStatsStatusCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewTransactionStatsStatusCache(store *sharedcachehelpers.CacheStore) TransactionStatsStatusCache {
	return &transactionStatsStatusCache{store: store}
}

func (t *transactionStatsStatusCache) GetMonthTransactionStatusSuccessCache(ctx context.Context, req *requests.MonthStatusTransaction) ([]*salemongo.TransactionMonthlyStatusSuccessRow, bool) {
	key := fmt.Sprintf(monthTransactionStatusSuccessCacheKey, req.Month, req.Year)
	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TransactionMonthlyStatusSuccessRow](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (t *transactionStatsStatusCache) SetMonthTransactionStatusSuccessCache(ctx context.Context, req *requests.MonthStatusTransaction, data []*salemongo.TransactionMonthlyStatusSuccessRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(monthTransactionStatusSuccessCacheKey, req.Month, req.Year)
	sharedcachehelpers.SetToCache(ctx, t.store, key, &data, ttlDefault)
}

func (t *transactionStatsStatusCache) GetYearTransactionStatusSuccessCache(ctx context.Context, year int) ([]*salemongo.TransactionYearlyStatusSuccessRow, bool) {
	key := fmt.Sprintf(yearTransactionStatusSuccessCacheKey, year)
	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TransactionYearlyStatusSuccessRow](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (t *transactionStatsStatusCache) SetYearTransactionStatusSuccessCache(ctx context.Context, year int, data []*salemongo.TransactionYearlyStatusSuccessRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(yearTransactionStatusSuccessCacheKey, year)
	sharedcachehelpers.SetToCache(ctx, t.store, key, &data, ttlDefault)
}

func (t *transactionStatsStatusCache) GetMonthTransactionStatusFailedCache(ctx context.Context, req *requests.MonthStatusTransaction) ([]*salemongo.TransactionMonthlyStatusFailedRow, bool) {
	key := fmt.Sprintf(monthTransactionStatusFailedCacheKey, req.Month, req.Year)
	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TransactionMonthlyStatusFailedRow](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (t *transactionStatsStatusCache) SetMonthTransactionStatusFailedCache(ctx context.Context, req *requests.MonthStatusTransaction, data []*salemongo.TransactionMonthlyStatusFailedRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(monthTransactionStatusFailedCacheKey, req.Month, req.Year)
	sharedcachehelpers.SetToCache(ctx, t.store, key, &data, ttlDefault)
}

func (t *transactionStatsStatusCache) GetYearTransactionStatusFailedCache(ctx context.Context, year int) ([]*salemongo.TransactionYearlyStatusFailedRow, bool) {
	key := fmt.Sprintf(yearTransactionStatusFailedCacheKey, year)
	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TransactionYearlyStatusFailedRow](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (t *transactionStatsStatusCache) SetYearTransactionStatusFailedCache(ctx context.Context, year int, data []*salemongo.TransactionYearlyStatusFailedRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(yearTransactionStatusFailedCacheKey, year)
	sharedcachehelpers.SetToCache(ctx, t.store, key, &data, ttlDefault)
}
