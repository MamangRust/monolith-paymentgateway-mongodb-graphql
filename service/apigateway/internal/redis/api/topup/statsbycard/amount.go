package topup_stats_bycard_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type topupStatsAmountByCardCache struct {
	store *cache.CacheStore
}

func NewTopupStatsAmountByCardCache(store *cache.CacheStore) TopupStatsAmountByCardCache {
	return &topupStatsAmountByCardCache{store: store}
}

func (s *topupStatsAmountByCardCache) GetMonthlyTopupAmountsByCardNumberCache(ctx context.Context, req *model.FindYearTopupCardNumberInput) (*model.APIResponseTopupMonthAmount, bool) {
	key := fmt.Sprintf(monthTopupAmountByCardCacheKey, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTopupMonthAmount](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (s *topupStatsAmountByCardCache) SetMonthlyTopupAmountsByCardNumberCache(ctx context.Context, req *model.FindYearTopupCardNumberInput, data *model.APIResponseTopupMonthAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(monthTopupAmountByCardCacheKey, req.CardNumber, req.Year)
	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}

func (s *topupStatsAmountByCardCache) GetYearlyTopupAmountsByCardNumberCache(ctx context.Context, req *model.FindYearTopupCardNumberInput) (*model.APIResponseTopupYearAmount, bool) {
	key := fmt.Sprintf(yearTopupAmountByCardCacheKey, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTopupYearAmount](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (s *topupStatsAmountByCardCache) SetYearlyTopupAmountsByCardNumberCache(ctx context.Context, req *model.FindYearTopupCardNumberInput, data *model.APIResponseTopupYearAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(yearTopupAmountByCardCacheKey, req.CardNumber, req.Year)
	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}
