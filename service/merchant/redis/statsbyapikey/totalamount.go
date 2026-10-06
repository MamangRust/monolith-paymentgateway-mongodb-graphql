package merchantstatsapikey

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type merchantStatsTotalAmountByApiKeyCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewMerchantStatsTotalAmountByApiKeyCache(store *sharedcachehelpers.CacheStore) MerchantStatsTotalAmountByApiKeyCache {
	return &merchantStatsTotalAmountByApiKeyCache{store: store}
}

func (m *merchantStatsTotalAmountByApiKeyCache) GetMonthlyTotalAmountByApikeysCache(ctx context.Context, req *requests.MonthYearTotalAmountApiKey) ([]*salemongo.MerchantMonthlyTotalAmountRow, bool) {
	key := fmt.Sprintf(merchantMonthlyTotalAmountByApikeyCacheKey, req.Apikey, req.Year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MerchantMonthlyTotalAmountRow](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (m *merchantStatsTotalAmountByApiKeyCache) SetMonthlyTotalAmountByApikeysCache(ctx context.Context, req *requests.MonthYearTotalAmountApiKey, data []*salemongo.MerchantMonthlyTotalAmountRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(merchantMonthlyTotalAmountByApikeyCacheKey, req.Apikey, req.Year)

	sharedcachehelpers.SetToCache(ctx, m.store, key, &data, ttlDefault)
}

func (m *merchantStatsTotalAmountByApiKeyCache) GetYearlyTotalAmountByApikeysCache(ctx context.Context, req *requests.MonthYearTotalAmountApiKey) ([]*salemongo.MerchantYearlyTotalAmountRow, bool) {
	key := fmt.Sprintf(merchantYearlyTotalAmountByApikeyCacheKey, req.Apikey, req.Year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MerchantYearlyTotalAmountRow](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (m *merchantStatsTotalAmountByApiKeyCache) SetYearlyTotalAmountByApikeysCache(ctx context.Context, req *requests.MonthYearTotalAmountApiKey, data []*salemongo.MerchantYearlyTotalAmountRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(merchantYearlyTotalAmountByApikeyCacheKey, req.Apikey, req.Year)

	sharedcachehelpers.SetToCache(ctx, m.store, key, &data, ttlDefault)
}
