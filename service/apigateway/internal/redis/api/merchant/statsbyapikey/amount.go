package merchant_stats_byapikey_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type merchantStatsAmountByApiKeyCache struct {
	store *cache.CacheStore
}

func NewMerchantStatsAmountByApiKeyCache(store *cache.CacheStore) MerchantStatsAmountByApiKeyCache {
	return &merchantStatsAmountByApiKeyCache{store: store}
}

func (m *merchantStatsAmountByApiKeyCache) GetMonthlyAmountByApikeysCache(ctx context.Context, req *model.FindYearMerchantByApikeyInput) (*model.APIResponseMerchantMonthlyAmount, bool) {
	key := fmt.Sprintf(merchantMonthlyAmountByApikeyCacheKey, req.APIKey, req.Year)
	result, found := cache.GetFromCache[model.APIResponseMerchantMonthlyAmount](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (m *merchantStatsAmountByApiKeyCache) SetMonthlyAmountByApikeysCache(ctx context.Context, req *model.FindYearMerchantByApikeyInput, data *model.APIResponseMerchantMonthlyAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantMonthlyAmountByApikeyCacheKey, req.APIKey, req.Year)
	cache.SetToCache(ctx, m.store, key, data, ttlDefault)
}

func (m *merchantStatsAmountByApiKeyCache) GetYearlyAmountByApikeysCache(ctx context.Context, req *model.FindYearMerchantByApikeyInput) (*model.APIResponseMerchantYearlyAmount, bool) {
	key := fmt.Sprintf(merchantYearlyAmountByApikeyCacheKey, req.APIKey, req.Year)
	result, found := cache.GetFromCache[model.APIResponseMerchantYearlyAmount](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (m *merchantStatsAmountByApiKeyCache) SetYearlyAmountByApikeysCache(ctx context.Context, req *model.FindYearMerchantByApikeyInput, data *model.APIResponseMerchantYearlyAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantYearlyAmountByApikeyCacheKey, req.APIKey, req.Year)
	cache.SetToCache(ctx, m.store, key, data, ttlDefault)
}
