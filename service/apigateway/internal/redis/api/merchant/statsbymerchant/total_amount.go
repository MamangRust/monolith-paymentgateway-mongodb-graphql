package merchant_stats_bymerchant_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type merchantStatsTotalAmountByMerchant struct {
	store *cache.CacheStore
}

func NewMerchantStatsTotalAmountByMerchantCache(store *cache.CacheStore) MerchantStatsTotalAmountByMerchantCache {
	return &merchantStatsTotalAmountByMerchant{store: store}
}

func (m *merchantStatsTotalAmountByMerchant) GetMonthlyTotalAmountByMerchantsCache(ctx context.Context, req *model.FindYearMerchantByIDInput) (*model.APIResponseMerchantMonthlyTotalAmount, bool) {
	key := fmt.Sprintf(merchantMonthlyTotalAmountByMerchantCacheKey, req.MerchantID, req.Year)
	result, found := cache.GetFromCache[model.APIResponseMerchantMonthlyTotalAmount](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (m *merchantStatsTotalAmountByMerchant) SetMonthlyTotalAmountByMerchantsCache(ctx context.Context, req *model.FindYearMerchantByIDInput, data *model.APIResponseMerchantMonthlyTotalAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantMonthlyTotalAmountByMerchantCacheKey, req.MerchantID, req.Year)
	cache.SetToCache(ctx, m.store, key, data, ttlDefault)
}

func (m *merchantStatsTotalAmountByMerchant) GetYearlyTotalAmountByMerchantsCache(ctx context.Context, req *model.FindYearMerchantByIDInput) (*model.APIResponseMerchantYearlyTotalAmount, bool) {
	key := fmt.Sprintf(merchantYearlyTotalAmountByMerchantCacheKey, req.MerchantID, req.Year)
	result, found := cache.GetFromCache[model.APIResponseMerchantYearlyTotalAmount](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (m *merchantStatsTotalAmountByMerchant) SetYearlyTotalAmountByMerchantsCache(ctx context.Context, req *model.FindYearMerchantByIDInput, data *model.APIResponseMerchantYearlyTotalAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantYearlyTotalAmountByMerchantCacheKey, req.MerchantID, req.Year)
	cache.SetToCache(ctx, m.store, key, data, ttlDefault)
}
