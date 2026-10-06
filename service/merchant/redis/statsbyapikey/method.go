package merchantstatsapikey

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type merchantStatsMethodByApiKeyCache struct {
	store *cache.CacheStore
}

func NewMerchantStatsMethodByApiKeyCache(store *cache.CacheStore) MerchantStatsMethodByApiKeyCache {
	return &merchantStatsMethodByApiKeyCache{store: store}
}

func (m *merchantStatsMethodByApiKeyCache) GetMonthlyPaymentMethodByApikeysCache(ctx context.Context, req *requests.MonthYearPaymentMethodApiKey) ([]*salemongo.MerchantMonthlyPaymentMethodRow, bool) {
	key := fmt.Sprintf(merchantMonthlyPaymentMethodByApikeyCacheKey, req.Apikey, req.Year)

	result, found := cache.GetFromCache[[]*salemongo.MerchantMonthlyPaymentMethodRow](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (m *merchantStatsMethodByApiKeyCache) SetMonthlyPaymentMethodByApikeysCache(ctx context.Context, req *requests.MonthYearPaymentMethodApiKey, data []*salemongo.MerchantMonthlyPaymentMethodRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(merchantMonthlyPaymentMethodByApikeyCacheKey, req.Apikey, req.Year)

	cache.SetToCache(ctx, m.store, key, &data, ttlDefault)
}

func (m *merchantStatsMethodByApiKeyCache) GetYearlyPaymentMethodByApikeysCache(ctx context.Context, req *requests.MonthYearPaymentMethodApiKey) ([]*salemongo.MerchantYearlyPaymentMethodRow, bool) {
	key := fmt.Sprintf(merchantYearlyPaymentMethodByApikeyCacheKey, req.Apikey, req.Year)

	result, found := cache.GetFromCache[[]*salemongo.MerchantYearlyPaymentMethodRow](ctx, m.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (m *merchantStatsMethodByApiKeyCache) SetYearlyPaymentMethodByApikeysCache(ctx context.Context, req *requests.MonthYearPaymentMethodApiKey, data []*salemongo.MerchantYearlyPaymentMethodRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(merchantYearlyPaymentMethodByApikeyCacheKey, req.Apikey, req.Year)

	cache.SetToCache(ctx, m.store, key, &data, ttlDefault)
}
