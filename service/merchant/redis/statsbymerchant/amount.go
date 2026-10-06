package merchantstatsbymerchant

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type merchantStatsAmountByMerchant struct {
	store *sharedcachehelpers.CacheStore
}

func NewMerchantStatsAmountByMerchantCache(store *sharedcachehelpers.CacheStore) MerchantStatsAmountByMerchantCache {
	return &merchantStatsAmountByMerchant{store: store}
}

func (m *merchantStatsAmountByMerchant) GetMonthlyAmountByMerchantsCache(ctx context.Context, req *requests.MonthYearAmountMerchant) ([]*salemongo.MerchantMonthlyAmountRow, bool) {
	key := fmt.Sprintf(merchantMonthlyAmountByMerchantCacheKey, req.MerchantID, req.Year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MerchantMonthlyAmountRow](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (m *merchantStatsAmountByMerchant) SetMonthlyAmountByMerchantsCache(ctx context.Context, req *requests.MonthYearAmountMerchant, data []*salemongo.MerchantMonthlyAmountRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(merchantMonthlyAmountByMerchantCacheKey, req.MerchantID, req.Year)

	sharedcachehelpers.SetToCache(ctx, m.store, key, &data, ttlDefault)
}

func (m *merchantStatsAmountByMerchant) GetYearlyAmountByMerchantsCache(ctx context.Context, req *requests.MonthYearAmountMerchant) ([]*salemongo.MerchantYearlyAmountRow, bool) {
	key := fmt.Sprintf(merchantYearlyAmountByMerchantCacheKey, req.MerchantID, req.Year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MerchantYearlyAmountRow](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (m *merchantStatsAmountByMerchant) SetYearlyAmountByMerchantsCache(ctx context.Context, req *requests.MonthYearAmountMerchant, data []*salemongo.MerchantYearlyAmountRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(merchantYearlyAmountByMerchantCacheKey, req.MerchantID, req.Year)

	sharedcachehelpers.SetToCache(ctx, m.store, key, &data, ttlDefault)
}
