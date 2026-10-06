package merchant_stats_byapikey_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type merchantStatsTotalAmountByApiKeyCache struct {
	store *cache.CacheStore
}

func NewMerchantStatsTotalAmountByApiKeyCache(store *cache.CacheStore) MerchantStatsTotalAmountByApiKeyCache {
	return &merchantStatsTotalAmountByApiKeyCache{store: store}
}

func (m *merchantStatsTotalAmountByApiKeyCache) GetMonthlyTotalAmountByApikeysCache(ctx context.Context, req *model.FindYearMerchantByApikeyInput) (*model.APIResponseMerchantMonthlyTotalAmount, bool) {
	key := fmt.Sprintf(merchantMonthlyTotalAmountByApikeyCacheKey, req.APIKey, req.Year)
	result, found := cache.GetFromCache[model.APIResponseMerchantMonthlyTotalAmount](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (m *merchantStatsTotalAmountByApiKeyCache) SetMonthlyTotalAmountByApikeysCache(ctx context.Context, req *model.FindYearMerchantByApikeyInput, data *model.APIResponseMerchantMonthlyTotalAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantMonthlyTotalAmountByApikeyCacheKey, req.APIKey, req.Year)
	cache.SetToCache(ctx, m.store, key, data, ttlDefault)
}

func (m *merchantStatsTotalAmountByApiKeyCache) GetYearlyTotalAmountByApikeysCache(ctx context.Context, req *model.FindYearMerchantByApikeyInput) (*model.APIResponseMerchantYearlyTotalAmount, bool) {
	key := fmt.Sprintf(merchantYearlyTotalAmountByApikeyCacheKey, req.APIKey, req.Year)
	result, found := cache.GetFromCache[model.APIResponseMerchantYearlyTotalAmount](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (m *merchantStatsTotalAmountByApiKeyCache) SetYearlyTotalAmountByApikeysCache(ctx context.Context, req *model.FindYearMerchantByApikeyInput, data *model.APIResponseMerchantYearlyTotalAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantYearlyTotalAmountByApikeyCacheKey, req.APIKey, req.Year)
	cache.SetToCache(ctx, m.store, key, data, ttlDefault)
}
