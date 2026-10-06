package transactionstatsbycarcache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type transactionStatsByCardMethodCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewTransactionStatsByCardMethodCache(store *sharedcachehelpers.CacheStore) TransactionStatsByCardMethodCache {
	return &transactionStatsByCardMethodCache{store: store}
}

func (t *transactionStatsByCardMethodCache) GetMonthlyPaymentMethodsByCardCache(ctx context.Context, req *requests.MonthYearPaymentMethod) ([]*salemongo.TransactionMonthlyPaymentMethodByCardRow, bool) {
	key := fmt.Sprintf(monthTransactionMethodByCardCacheKey, req.CardNumber, req.Year)
	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TransactionMonthlyPaymentMethodByCardRow](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (t *transactionStatsByCardMethodCache) SetMonthlyPaymentMethodsByCardCache(ctx context.Context, req *requests.MonthYearPaymentMethod, data []*salemongo.TransactionMonthlyPaymentMethodByCardRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(monthTransactionMethodByCardCacheKey, req.CardNumber, req.Year)
	sharedcachehelpers.SetToCache(ctx, t.store, key, &data, ttlDefault)
}

func (t *transactionStatsByCardMethodCache) GetYearlyPaymentMethodsByCardCache(ctx context.Context, req *requests.MonthYearPaymentMethod) ([]*salemongo.TransactionYearlyPaymentMethodByCardRow, bool) {
	key := fmt.Sprintf(yearTransactionMethodByCardCacheKey, req.CardNumber, req.Year)
	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TransactionYearlyPaymentMethodByCardRow](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (t *transactionStatsByCardMethodCache) SetYearlyPaymentMethodsByCardCache(ctx context.Context, req *requests.MonthYearPaymentMethod, data []*salemongo.TransactionYearlyPaymentMethodByCardRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(yearTransactionMethodByCardCacheKey, req.CardNumber, req.Year)
	sharedcachehelpers.SetToCache(ctx, t.store, key, &data, ttlDefault)
}
