package merchantstatsbymerchant

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type merchantStatsMethodByMerchant struct {
	store *sharedcachehelpers.CacheStore
}

func NewMerchantStatsMethodByMerchantCache(store *sharedcachehelpers.CacheStore) MerchantStatsMethodByMerchantCache {
	return &merchantStatsMethodByMerchant{store: store}
}

func (m *merchantStatsMethodByMerchant) GetMonthlyPaymentMethodByMerchantsCache(ctx context.Context, req *requests.MonthYearPaymentMethodMerchant) ([]*salemongo.MerchantMonthlyPaymentMethodRow, bool) {
	key := fmt.Sprintf(merchantMonthlyPaymentMethodByMerchantCacheKey, req.MerchantID, req.Year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MerchantMonthlyPaymentMethodRow](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (m *merchantStatsMethodByMerchant) SetMonthlyPaymentMethodByMerchantsCache(ctx context.Context, req *requests.MonthYearPaymentMethodMerchant, data []*salemongo.MerchantMonthlyPaymentMethodRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(merchantMonthlyPaymentMethodByMerchantCacheKey, req.MerchantID, req.Year)

	sharedcachehelpers.SetToCache(ctx, m.store, key, &data, ttlDefault)
}

func (m *merchantStatsMethodByMerchant) GetYearlyPaymentMethodByMerchantsCache(ctx context.Context, req *requests.MonthYearPaymentMethodMerchant) ([]*salemongo.MerchantYearlyPaymentMethodRow, bool) {
	key := fmt.Sprintf(merchantYearlyPaymentMethodByMerchantCacheKey, req.MerchantID, req.Year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MerchantYearlyPaymentMethodRow](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (m *merchantStatsMethodByMerchant) SetYearlyPaymentMethodByMerchantsCache(ctx context.Context, req *requests.MonthYearPaymentMethodMerchant, data []*salemongo.MerchantYearlyPaymentMethodRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(merchantYearlyPaymentMethodByMerchantCacheKey, req.MerchantID, req.Year)

	sharedcachehelpers.SetToCache(ctx, m.store, key, &data, ttlDefault)
}
