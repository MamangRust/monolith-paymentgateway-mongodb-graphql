package merchantstatscache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type merchantStatsAmountCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewMerchantStatsAmountCache(store *sharedcachehelpers.CacheStore) MerchantStatsAmountCache {
	return &merchantStatsAmountCache{store: store}
}

func (s *merchantStatsAmountCache) GetMonthlyAmountMerchantCache(ctx context.Context, year int) ([]*salemongo.MerchantMonthlyAmountRow, bool) {
	key := fmt.Sprintf(merchantMonthlyAmountCacheKey, year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MerchantMonthlyAmountRow](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (s *merchantStatsAmountCache) SetMonthlyAmountMerchantCache(ctx context.Context, year int, data []*salemongo.MerchantMonthlyAmountRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(merchantMonthlyAmountCacheKey, year)

	sharedcachehelpers.SetToCache(ctx, s.store, key, &data, ttlDefault)
}

func (s *merchantStatsAmountCache) GetYearlyAmountMerchantCache(ctx context.Context, year int) ([]*salemongo.MerchantYearlyAmountRow, bool) {
	key := fmt.Sprintf(MerchantYearlyAmountCacheKey, year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MerchantYearlyAmountRow](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (s *merchantStatsAmountCache) SetYearlyAmountMerchantCache(ctx context.Context, year int, data []*salemongo.MerchantYearlyAmountRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(MerchantYearlyAmountCacheKey, year)

	sharedcachehelpers.SetToCache(ctx, s.store, key, &data, ttlDefault)
}
