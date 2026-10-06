package merchant_stats_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type merchantStatsAmountCache struct {
	store *cache.CacheStore
}

func NewMerchantStatsAmountCache(store *cache.CacheStore) MerchantStatsAmountCache {
	return &merchantStatsAmountCache{store: store}
}

func (s *merchantStatsAmountCache) GetMonthlyAmountMerchantCache(ctx context.Context, year int) (*model.APIResponseMerchantMonthlyAmount, bool) {
	key := fmt.Sprintf(merchantMonthlyAmountCacheKey, year)
	result, found := cache.GetFromCache[model.APIResponseMerchantMonthlyAmount](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (s *merchantStatsAmountCache) SetMonthlyAmountMerchantCache(ctx context.Context, year int, data *model.APIResponseMerchantMonthlyAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantMonthlyAmountCacheKey, year)
	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}

func (s *merchantStatsAmountCache) GetYearlyAmountMerchantCache(ctx context.Context, year int) (*model.APIResponseMerchantYearlyAmount, bool) {
	key := fmt.Sprintf(MerchantYearlyAmountCacheKey, year)
	result, found := cache.GetFromCache[model.APIResponseMerchantYearlyAmount](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (s *merchantStatsAmountCache) SetYearlyAmountMerchantCache(ctx context.Context, year int, data *model.APIResponseMerchantYearlyAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(MerchantYearlyAmountCacheKey, year)
	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}
