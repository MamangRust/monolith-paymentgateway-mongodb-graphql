package merchant_stats_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type merchantStatsMethodCache struct {
	store *cache.CacheStore
}

func NewMerchantStatsMethodCache(store *cache.CacheStore) MerchantStatsMethodCache {
	return &merchantStatsMethodCache{store: store}
}

func (s *merchantStatsMethodCache) GetMonthlyPaymentMethodsMerchantCache(ctx context.Context, year int) (*model.APIResponseMerchantMonthlyPaymentMethod, bool) {
	key := fmt.Sprintf(merchantMonthlyPaymentMethodCacheKey, year)
	result, found := cache.GetFromCache[model.APIResponseMerchantMonthlyPaymentMethod](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (s *merchantStatsMethodCache) SetMonthlyPaymentMethodsMerchantCache(ctx context.Context, year int, data *model.APIResponseMerchantMonthlyPaymentMethod) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantMonthlyPaymentMethodCacheKey, year)
	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}

func (s *merchantStatsMethodCache) GetYearlyPaymentMethodMerchantCache(ctx context.Context, year int) (*model.APIResponseMerchantYearlyPaymentMethod, bool) {
	key := fmt.Sprintf(merchantYearlyPaymentMethodCacheKey, year)
	result, found := cache.GetFromCache[model.APIResponseMerchantYearlyPaymentMethod](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (s *merchantStatsMethodCache) SetYearlyPaymentMethodMerchantCache(ctx context.Context, year int, data *model.APIResponseMerchantYearlyPaymentMethod) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantYearlyPaymentMethodCacheKey, year)
	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}
