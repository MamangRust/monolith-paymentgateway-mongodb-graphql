package transactionstatsbycarcache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type transactionStatsByCardAmountCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewTransactionStatsByCardAmountCache(store *sharedcachehelpers.CacheStore) TransactionStatsByCardAmountCache {
	return &transactionStatsByCardAmountCache{store: store}
}

func (t *transactionStatsByCardAmountCache) GetMonthlyAmountsByCardCache(ctx context.Context, req *requests.MonthYearPaymentMethod) ([]*salemongo.TransactionMonthlyAmountByCardRow, bool) {
	key := fmt.Sprintf(monthTransactionAmountByCardCacheKey, req.CardNumber, req.Year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TransactionMonthlyAmountByCardRow](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (t *transactionStatsByCardAmountCache) SetMonthlyAmountsByCardCache(ctx context.Context, req *requests.MonthYearPaymentMethod, data []*salemongo.TransactionMonthlyAmountByCardRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(monthTransactionAmountByCardCacheKey, req.CardNumber, req.Year)
	sharedcachehelpers.SetToCache(ctx, t.store, key, &data, ttlDefault)
}

func (t *transactionStatsByCardAmountCache) GetYearlyAmountsByCardCache(ctx context.Context, req *requests.MonthYearPaymentMethod) ([]*salemongo.TransactionYearlyAmountByCardRow, bool) {
	key := fmt.Sprintf(yearTransactionAmountByCardCacheKey, req.CardNumber, req.Year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TransactionYearlyAmountByCardRow](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (t *transactionStatsByCardAmountCache) SetYearlyAmountsByCardCache(ctx context.Context, req *requests.MonthYearPaymentMethod, data []*salemongo.TransactionYearlyAmountByCardRow) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(yearTransactionAmountByCardCacheKey, req.CardNumber, req.Year)
	sharedcachehelpers.SetToCache(ctx, t.store, key, &data, ttlDefault)
}
