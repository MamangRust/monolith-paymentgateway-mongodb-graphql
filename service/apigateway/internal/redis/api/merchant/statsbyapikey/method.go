package merchant_stats_byapikey_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type merchantStatsMethodByApiKeyCache struct {
	store *cache.CacheStore
}

func NewMerchantStatsMethodByApiKeyCache(store *cache.CacheStore) MerchantStatsMethodByApiKeyCache {
	return &merchantStatsMethodByApiKeyCache{store: store}
}

func (m *merchantStatsMethodByApiKeyCache) GetMonthlyPaymentMethodByApikeysCache(ctx context.Context, req *model.FindYearMerchantByApikeyInput) (*model.APIResponseMerchantMonthlyPaymentMethod, bool) {
	key := fmt.Sprintf(merchantMonthlyPaymentMethodByApikeyCacheKey, req.APIKey, req.Year)
	result, found := cache.GetFromCache[model.APIResponseMerchantMonthlyPaymentMethod](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (m *merchantStatsMethodByApiKeyCache) SetMonthlyPaymentMethodByApikeysCache(ctx context.Context, req *model.FindYearMerchantByApikeyInput, data *model.APIResponseMerchantMonthlyPaymentMethod) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantMonthlyPaymentMethodByApikeyCacheKey, req.APIKey, req.Year)
	cache.SetToCache(ctx, m.store, key, data, ttlDefault)
}

func (m *merchantStatsMethodByApiKeyCache) GetYearlyPaymentMethodByApikeysCache(ctx context.Context, req *model.FindYearMerchantByApikeyInput) (*model.APIResponseMerchantYearlyPaymentMethod, bool) {
	key := fmt.Sprintf(merchantYearlyPaymentMethodByApikeyCacheKey, req.APIKey, req.Year)
	result, found := cache.GetFromCache[model.APIResponseMerchantYearlyPaymentMethod](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (m *merchantStatsMethodByApiKeyCache) SetYearlyPaymentMethodByApikeysCache(ctx context.Context, req *model.FindYearMerchantByApikeyInput, data *model.APIResponseMerchantYearlyPaymentMethod) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantYearlyPaymentMethodByApikeyCacheKey, req.APIKey, req.Year)
	cache.SetToCache(ctx, m.store, key, data, ttlDefault)
}
