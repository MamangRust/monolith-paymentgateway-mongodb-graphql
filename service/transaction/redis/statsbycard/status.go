package transactionstatsbycarcache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type transactionStatsByCardStatusCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewTransactionStatsByCardStatusCache(store *sharedcachehelpers.CacheStore) TransactionStatsByCardStatusCache {
	return &transactionStatsByCardStatusCache{store: store}
}

func (t *transactionStatsByCardStatusCache) GetMonthTransactionStatusSuccessByCardCache(ctx context.Context, req *requests.MonthStatusTransactionCardNumber) ([]*salemongo.TransactionMonthlyStatusSuccessByCardRow, bool) {
	key := fmt.Sprintf(monthTransactionStatusSuccessByCardCacheKey, req.CardNumber, req.Month, req.Year)
	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TransactionMonthlyStatusSuccessByCardRow](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (t *transactionStatsByCardStatusCache) SetMonthTransactionStatusSuccessByCardCache(ctx context.Context, req *requests.MonthStatusTransactionCardNumber, data []*salemongo.TransactionMonthlyStatusSuccessByCardRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(monthTransactionStatusSuccessByCardCacheKey, req.CardNumber, req.Month, req.Year)
	sharedcachehelpers.SetToCache(ctx, t.store, key, &data, ttlDefault)
}

func (t *transactionStatsByCardStatusCache) GetYearTransactionStatusSuccessByCardCache(ctx context.Context, req *requests.YearStatusTransactionCardNumber) ([]*salemongo.TransactionYearlyStatusSuccessByCardRow, bool) {
	key := fmt.Sprintf(yearTransactionStatusSuccessByCardCacheKey, req.CardNumber, req.Year)
	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TransactionYearlyStatusSuccessByCardRow](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (t *transactionStatsByCardStatusCache) SetYearTransactionStatusSuccessByCardCache(ctx context.Context, req *requests.YearStatusTransactionCardNumber, data []*salemongo.TransactionYearlyStatusSuccessByCardRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(yearTransactionStatusSuccessByCardCacheKey, req.CardNumber, req.Year)
	sharedcachehelpers.SetToCache(ctx, t.store, key, &data, ttlDefault)
}

func (t *transactionStatsByCardStatusCache) GetMonthTransactionStatusFailedByCardCache(ctx context.Context, req *requests.MonthStatusTransactionCardNumber) ([]*salemongo.TransactionMonthlyStatusFailedByCardRow, bool) {
	key := fmt.Sprintf(monthTransactionStatusFailedByCardCacheKey, req.CardNumber, req.Month, req.Year)
	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TransactionMonthlyStatusFailedByCardRow](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (t *transactionStatsByCardStatusCache) SetMonthTransactionStatusFailedByCardCache(ctx context.Context, req *requests.MonthStatusTransactionCardNumber, data []*salemongo.TransactionMonthlyStatusFailedByCardRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(monthTransactionStatusFailedByCardCacheKey, req.CardNumber, req.Month, req.Year)
	sharedcachehelpers.SetToCache(ctx, t.store, key, &data, ttlDefault)
}

func (t *transactionStatsByCardStatusCache) GetYearTransactionStatusFailedByCardCache(ctx context.Context, req *requests.YearStatusTransactionCardNumber) ([]*salemongo.TransactionYearlyStatusFailedByCardRow, bool) {
	key := fmt.Sprintf(yearTransactionStatusFailedByCardCacheKey, req.CardNumber, req.Year)
	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TransactionYearlyStatusFailedByCardRow](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (t *transactionStatsByCardStatusCache) SetYearTransactionStatusFailedByCardCache(ctx context.Context, req *requests.YearStatusTransactionCardNumber, data []*salemongo.TransactionYearlyStatusFailedByCardRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(yearTransactionStatusFailedByCardCacheKey, req.CardNumber, req.Year)
	sharedcachehelpers.SetToCache(ctx, t.store, key, &data, ttlDefault)
}
