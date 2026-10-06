package topup_stats_bycard_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type topupStatsMethodByCardCache struct {
	store *cache.CacheStore
}

func NewTopupStatsMethodByCardCache(store *cache.CacheStore) TopupStatsMethodByCardCache {
	return &topupStatsMethodByCardCache{store: store}
}

func (s *topupStatsMethodByCardCache) GetMonthlyTopupMethodsByCardNumberCache(ctx context.Context, req *model.FindYearTopupCardNumberInput) (*model.APIResponseTopupMonthMethod, bool) {
	key := fmt.Sprintf(monthTopupMethodByCardCacheKey, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTopupMonthMethod](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (s *topupStatsMethodByCardCache) SetMonthlyTopupMethodsByCardNumberCache(ctx context.Context, req *model.FindYearTopupCardNumberInput, data *model.APIResponseTopupMonthMethod) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(monthTopupMethodByCardCacheKey, req.CardNumber, req.Year)
	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}

func (s *topupStatsMethodByCardCache) GetYearlyTopupMethodsByCardNumberCache(ctx context.Context, req *model.FindYearTopupCardNumberInput) (*model.APIResponseTopupYearMethod, bool) {
	key := fmt.Sprintf(yearTopupMethodByCardCacheKey, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTopupYearMethod](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (s *topupStatsMethodByCardCache) SetYearlyTopupMethodsByCardNumberCache(ctx context.Context, req *model.FindYearTopupCardNumberInput, data *model.APIResponseTopupYearMethod) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(yearTopupMethodByCardCacheKey, req.CardNumber, req.Year)
	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}
