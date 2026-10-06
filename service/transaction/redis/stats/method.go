package transactionstatscache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type transactionStatsMethodCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewTransactionStatsMethodCache(store *sharedcachehelpers.CacheStore) TransactionStatsMethodCache {
	return &transactionStatsMethodCache{store: store}
}

func (t *transactionStatsMethodCache) GetMonthlyPaymentMethodsCache(ctx context.Context, year int) ([]*salemongo.TransactionMonthlyPaymentMethodRow, bool) {
	key := fmt.Sprintf(monthTransactionMethodCacheKey, year)
	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TransactionMonthlyPaymentMethodRow](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (t *transactionStatsMethodCache) SetMonthlyPaymentMethodsCache(ctx context.Context, year int, data []*salemongo.TransactionMonthlyPaymentMethodRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(monthTransactionMethodCacheKey, year)
	sharedcachehelpers.SetToCache(ctx, t.store, key, &data, ttlDefault)
}

func (t *transactionStatsMethodCache) GetYearlyPaymentMethodsCache(ctx context.Context, year int) ([]*salemongo.TransactionYearlyPaymentMethodRow, bool) {
	key := fmt.Sprintf(yearTransactionMethodCacheKey, year)
	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TransactionYearlyPaymentMethodRow](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (t *transactionStatsMethodCache) SetYearlyPaymentMethodsCache(ctx context.Context, year int, data []*salemongo.TransactionYearlyPaymentMethodRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(yearTransactionMethodCacheKey, year)
	sharedcachehelpers.SetToCache(ctx, t.store, key, &data, ttlDefault)
}
