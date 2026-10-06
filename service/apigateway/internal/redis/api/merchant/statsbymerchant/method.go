package merchant_stats_bymerchant_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type merchantStatsMethodByMerchant struct {
	store *cache.CacheStore
}

func NewMerchantStatsMethodByMerchantCache(store *cache.CacheStore) MerchantStatsMethodByMerchantCache {
	return &merchantStatsMethodByMerchant{store: store}
}

func (m *merchantStatsMethodByMerchant) GetMonthlyPaymentMethodByMerchantsCache(ctx context.Context, req *model.FindYearMerchantByIDInput) (*model.APIResponseMerchantMonthlyPaymentMethod, bool) {
	key := fmt.Sprintf(merchantMonthlyPaymentMethodByMerchantCacheKey, req.MerchantID, req.Year)
	result, found := cache.GetFromCache[model.APIResponseMerchantMonthlyPaymentMethod](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (m *merchantStatsMethodByMerchant) SetMonthlyPaymentMethodByMerchantsCache(ctx context.Context, req *model.FindYearMerchantByIDInput, data *model.APIResponseMerchantMonthlyPaymentMethod) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantMonthlyPaymentMethodByMerchantCacheKey, req.MerchantID, req.Year)
	cache.SetToCache(ctx, m.store, key, data, ttlDefault)
}

func (m *merchantStatsMethodByMerchant) GetYearlyPaymentMethodByMerchantsCache(ctx context.Context, req *model.FindYearMerchantByIDInput) (*model.APIResponseMerchantYearlyPaymentMethod, bool) {
	key := fmt.Sprintf(merchantYearlyPaymentMethodByMerchantCacheKey, req.MerchantID, req.Year)
	result, found := cache.GetFromCache[model.APIResponseMerchantYearlyPaymentMethod](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (m *merchantStatsMethodByMerchant) SetYearlyPaymentMethodByMerchantsCache(ctx context.Context, req *model.FindYearMerchantByIDInput, data *model.APIResponseMerchantYearlyPaymentMethod) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantYearlyPaymentMethodByMerchantCacheKey, req.MerchantID, req.Year)
	cache.SetToCache(ctx, m.store, key, data, ttlDefault)
}
