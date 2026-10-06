package merchantstatscache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type merchantStatsTotalAmountCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewMerchantStatsTotalAmountCache(store *sharedcachehelpers.CacheStore) MerchantStatsTotalAmountCache {
	return &merchantStatsTotalAmountCache{store: store}
}

func (s *merchantStatsTotalAmountCache) GetMonthlyTotalAmountMerchantCache(ctx context.Context, year int) ([]*salemongo.MerchantMonthlyTotalAmountRow, bool) {
	key := fmt.Sprintf(merchantMonthlyTotalAmountCacheKey, year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MerchantMonthlyTotalAmountRow](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (s *merchantStatsTotalAmountCache) SetMonthlyTotalAmountMerchantCache(ctx context.Context, year int, data []*salemongo.MerchantMonthlyTotalAmountRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(merchantMonthlyTotalAmountCacheKey, year)

	sharedcachehelpers.SetToCache(ctx, s.store, key, &data, ttlDefault)
}

func (s *merchantStatsTotalAmountCache) GetYearlyTotalAmountMerchantCache(ctx context.Context, year int) ([]*salemongo.MerchantYearlyTotalAmountRow, bool) {
	key := fmt.Sprintf(merchantYearlyTotalAmountCacheKey, year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MerchantYearlyTotalAmountRow](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (s *merchantStatsTotalAmountCache) SetYearlyTotalAmountMerchantCache(ctx context.Context, year int, data []*salemongo.MerchantYearlyTotalAmountRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(merchantYearlyTotalAmountCacheKey, year)

	sharedcachehelpers.SetToCache(ctx, s.store, key, &data, ttlDefault)
}
