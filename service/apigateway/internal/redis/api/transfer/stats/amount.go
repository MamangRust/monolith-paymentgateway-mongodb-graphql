package transfer_stats_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type transferStatsAmountCache struct {
	store *cache.CacheStore
}

func NewTransferStatsAmountCache(store *cache.CacheStore) TransferStatsAmountCache {
	return &transferStatsAmountCache{store: store}
}

func (t *transferStatsAmountCache) GetCachedMonthTransferAmounts(ctx context.Context, year int) (*model.APIResponseTransferMonthAmount, bool) {
	key := fmt.Sprintf(transferMonthTransferAmountKey, year)
	result, found := cache.GetFromCache[model.APIResponseTransferMonthAmount](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transferStatsAmountCache) SetCachedMonthTransferAmounts(ctx context.Context, year int, data *model.APIResponseTransferMonthAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(transferMonthTransferAmountKey, year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transferStatsAmountCache) GetCachedYearlyTransferAmounts(ctx context.Context, year int) (*model.APIResponseTransferYearAmount, bool) {
	key := fmt.Sprintf(transferYearTransferAmountKey, year)
	result, found := cache.GetFromCache[model.APIResponseTransferYearAmount](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transferStatsAmountCache) SetCachedYearlyTransferAmounts(ctx context.Context, year int, data *model.APIResponseTransferYearAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(transferYearTransferAmountKey, year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}
