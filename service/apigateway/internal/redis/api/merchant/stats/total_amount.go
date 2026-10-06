package merchant_stats_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type merchantStatsTotalAmountCache struct {
	store *cache.CacheStore
}

func NewMerchantStatsTotalAmountCache(store *cache.CacheStore) MerchantStatsTotalAmountCache {
	return &merchantStatsTotalAmountCache{store: store}
}

func (s *merchantStatsTotalAmountCache) GetMonthlyTotalAmountMerchantCache(ctx context.Context, year int) (*model.APIResponseMerchantMonthlyTotalAmount, bool) {
	key := fmt.Sprintf(merchantMonthlyTotalAmountCacheKey, year)
	result, found := cache.GetFromCache[model.APIResponseMerchantMonthlyTotalAmount](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (s *merchantStatsTotalAmountCache) SetMonthlyTotalAmountMerchantCache(ctx context.Context, year int, data *model.APIResponseMerchantMonthlyTotalAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantMonthlyTotalAmountCacheKey, year)
	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}

func (s *merchantStatsTotalAmountCache) GetYearlyTotalAmountMerchantCache(ctx context.Context, year int) (*model.APIResponseMerchantYearlyTotalAmount, bool) {
	key := fmt.Sprintf(merchantYearlyTotalAmountCacheKey, year)
	result, found := cache.GetFromCache[model.APIResponseMerchantYearlyTotalAmount](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (s *merchantStatsTotalAmountCache) SetYearlyTotalAmountMerchantCache(ctx context.Context, year int, data *model.APIResponseMerchantYearlyTotalAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantYearlyTotalAmountCacheKey, year)
	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}
