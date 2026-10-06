package merchantstatsapikey

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type merchantStatsAmountByApiKeyCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewMerchantStatsAmountByApiKeyCache(store *sharedcachehelpers.CacheStore) MerchantStatsAmountByApiKeyCache {
	return &merchantStatsAmountByApiKeyCache{store: store}
}

func (m *merchantStatsAmountByApiKeyCache) GetMonthlyAmountByApikeysCache(ctx context.Context, req *requests.MonthYearAmountApiKey) ([]*salemongo.MerchantMonthlyAmountRow, bool) {
	key := fmt.Sprintf(merchantMonthlyAmountByApikeyCacheKey, req.Apikey, req.Year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MerchantMonthlyAmountRow](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (m *merchantStatsAmountByApiKeyCache) SetMonthlyAmountByApikeysCache(ctx context.Context, req *requests.MonthYearAmountApiKey, data []*salemongo.MerchantMonthlyAmountRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(merchantMonthlyAmountByApikeyCacheKey, req.Apikey, req.Year)

	sharedcachehelpers.SetToCache(ctx, m.store, key, &data, ttlDefault)
}

func (m *merchantStatsAmountByApiKeyCache) GetYearlyAmountByApikeysCache(ctx context.Context, req *requests.MonthYearAmountApiKey) ([]*salemongo.MerchantYearlyAmountRow, bool) {
	key := fmt.Sprintf(merchantYearlyAmountByApikeyCacheKey, req.Apikey, req.Year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MerchantYearlyAmountRow](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (m *merchantStatsAmountByApiKeyCache) SetYearlyAmountByApikeysCache(ctx context.Context, req *requests.MonthYearAmountApiKey, data []*salemongo.MerchantYearlyAmountRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(merchantYearlyAmountByApikeyCacheKey, req.Apikey, req.Year)

	sharedcachehelpers.SetToCache(ctx, m.store, key, &data, ttlDefault)
}
