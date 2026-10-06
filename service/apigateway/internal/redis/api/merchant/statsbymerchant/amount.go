package merchant_stats_bymerchant_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type merchantStatsAmountByMerchant struct {
	store *cache.CacheStore
}

func NewMerchantStatsAmountByMerchantCache(store *cache.CacheStore) MerchantStatsAmountByMerchantCache {
	return &merchantStatsAmountByMerchant{store: store}
}

func (m *merchantStatsAmountByMerchant) GetMonthlyAmountByMerchantsCache(ctx context.Context, req *model.FindYearMerchantByIDInput) (*model.APIResponseMerchantMonthlyAmount, bool) {
	key := fmt.Sprintf(merchantMonthlyAmountByMerchantCacheKey, req.MerchantID, req.Year)
	result, found := cache.GetFromCache[model.APIResponseMerchantMonthlyAmount](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (m *merchantStatsAmountByMerchant) SetMonthlyAmountByMerchantsCache(ctx context.Context, req *model.FindYearMerchantByIDInput, data *model.APIResponseMerchantMonthlyAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantMonthlyAmountByMerchantCacheKey, req.MerchantID, req.Year)
	cache.SetToCache(ctx, m.store, key, data, ttlDefault)
}

func (m *merchantStatsAmountByMerchant) GetYearlyAmountByMerchantsCache(ctx context.Context, req *model.FindYearMerchantByIDInput) (*model.APIResponseMerchantYearlyAmount, bool) {
	key := fmt.Sprintf(merchantYearlyAmountByMerchantCacheKey, req.MerchantID, req.Year)
	result, found := cache.GetFromCache[model.APIResponseMerchantYearlyAmount](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (m *merchantStatsAmountByMerchant) SetYearlyAmountByMerchantsCache(ctx context.Context, req *model.FindYearMerchantByIDInput, data *model.APIResponseMerchantYearlyAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantYearlyAmountByMerchantCacheKey, req.MerchantID, req.Year)
	cache.SetToCache(ctx, m.store, key, data, ttlDefault)
}
