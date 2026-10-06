package withdrawstatscache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type withdrawStatsAmountCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewWithdrawStatsAmountCache(store *sharedcachehelpers.CacheStore) WithdrawStatsAmountCache {
	return &withdrawStatsAmountCache{store: store}
}

func (w *withdrawStatsAmountCache) GetCachedMonthlyWithdraws(ctx context.Context, year int) ([]*salemongo.WithdrawMonthlyAmountRow, bool) {
	key := fmt.Sprintf(montWithdrawAmountKey, year)
	result, found := cache.GetFromCache[[]*salemongo.WithdrawMonthlyAmountRow](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (w *withdrawStatsAmountCache) SetCachedMonthlyWithdraws(ctx context.Context, year int, data []*salemongo.WithdrawMonthlyAmountRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(montWithdrawAmountKey, year)
	cache.SetToCache(ctx, w.store, key, &data, ttlDefault)
}

func (w *withdrawStatsAmountCache) GetCachedYearlyWithdraws(ctx context.Context, year int) ([]*salemongo.WithdrawYearlyAmountRow, bool) {
	key := fmt.Sprintf(yearWithdrawAmountKey, year)
	result, found := cache.GetFromCache[[]*salemongo.WithdrawYearlyAmountRow](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (w *withdrawStatsAmountCache) SetCachedYearlyWithdraws(ctx context.Context, year int, data []*salemongo.WithdrawYearlyAmountRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(yearWithdrawAmountKey, year)
	cache.SetToCache(ctx, w.store, key, &data, ttlDefault)
}
