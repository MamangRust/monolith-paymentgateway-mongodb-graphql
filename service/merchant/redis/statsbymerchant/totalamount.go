package merchantstatsbymerchant

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type merchantStatsTotalAmountByMerchant struct {
	store *sharedcachehelpers.CacheStore
}

func NewMerchantStatsTotalAmountByMerchantCache(store *sharedcachehelpers.CacheStore) MerchantStatsTotalAmountByMerchantCache {
	return &merchantStatsTotalAmountByMerchant{store: store}
}

func (m *merchantStatsTotalAmountByMerchant) SetMonthlyTotalAmountByMerchantsCache(ctx context.Context, req *requests.MonthYearTotalAmountMerchant, data []*salemongo.MerchantMonthlyTotalAmountRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(merchantMonthlyTotalAmountByMerchantCacheKey, req.MerchantID, req.Year)

	sharedcachehelpers.SetToCache(ctx, m.store, key, &data, ttlDefault)
}

func (m *merchantStatsTotalAmountByMerchant) GetMonthlyTotalAmountByMerchantsCache(ctx context.Context, req *requests.MonthYearTotalAmountMerchant) ([]*salemongo.MerchantMonthlyTotalAmountRow, bool) {
	key := fmt.Sprintf(merchantMonthlyTotalAmountByMerchantCacheKey, req.MerchantID, req.Year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MerchantMonthlyTotalAmountRow](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (m *merchantStatsTotalAmountByMerchant) SetYearlyTotalAmountByMerchantsCache(ctx context.Context, req *requests.MonthYearTotalAmountMerchant, data []*salemongo.MerchantYearlyTotalAmountRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(merchantYearlyTotalAmountByMerchantCacheKey, req.MerchantID, req.Year)

	sharedcachehelpers.SetToCache(ctx, m.store, key, &data, ttlDefault)
}

func (m *merchantStatsTotalAmountByMerchant) GetYearlyTotalAmountByMerchantsCache(ctx context.Context, req *requests.MonthYearTotalAmountMerchant) ([]*salemongo.MerchantYearlyTotalAmountRow, bool) {
	key := fmt.Sprintf(merchantYearlyTotalAmountByMerchantCacheKey, req.MerchantID, req.Year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MerchantYearlyTotalAmountRow](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}
