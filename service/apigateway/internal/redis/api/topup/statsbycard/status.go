package topup_stats_bycard_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type topupStatsStatusByCardCache struct {
	store *cache.CacheStore
}

func NewTopupStatsStatusByCardCache(store *cache.CacheStore) TopupStatsStatusByCardCache {
	return &topupStatsStatusByCardCache{store: store}
}

func (s *topupStatsStatusByCardCache) GetMonthTopupStatusSuccessByCardNumberCache(ctx context.Context, req *model.FindMonthlyTopupStatusCardNumberInput) (*model.APIResponseTopupMonthStatusSuccess, bool) {
	key := fmt.Sprintf(monthTopupStatusSuccessByCardCacheKey, req.CardNumber, req.Month, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTopupMonthStatusSuccess](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (s *topupStatsStatusByCardCache) SetMonthTopupStatusSuccessByCardNumberCache(ctx context.Context, req *model.FindMonthlyTopupStatusCardNumberInput, data *model.APIResponseTopupMonthStatusSuccess) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(monthTopupStatusSuccessByCardCacheKey, req.CardNumber, req.Month, req.Year)
	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}

func (s *topupStatsStatusByCardCache) GetYearlyTopupStatusSuccessByCardNumberCache(ctx context.Context, req *model.FindYearTopupStatusCardNumberInput) (*model.APIResponseTopupYearStatusSuccess, bool) {
	key := fmt.Sprintf(yearTopupStatusSuccessByCardCacheKey, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTopupYearStatusSuccess](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (s *topupStatsStatusByCardCache) SetYearlyTopupStatusSuccessByCardNumberCache(ctx context.Context, req *model.FindYearTopupStatusCardNumberInput, data *model.APIResponseTopupYearStatusSuccess) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(yearTopupStatusSuccessByCardCacheKey, req.CardNumber, req.Year)
	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}

func (s *topupStatsStatusByCardCache) GetMonthTopupStatusFailedByCardNumberCache(ctx context.Context, req *model.FindMonthlyTopupStatusCardNumberInput) (*model.APIResponseTopupMonthStatusFailed, bool) {
	key := fmt.Sprintf(monthTopupStatusFailedByCardCacheKey, req.CardNumber, req.Month, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTopupMonthStatusFailed](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (s *topupStatsStatusByCardCache) SetMonthTopupStatusFailedByCardNumberCache(ctx context.Context, req *model.FindMonthlyTopupStatusCardNumberInput, data *model.APIResponseTopupMonthStatusFailed) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(monthTopupStatusFailedByCardCacheKey, req.CardNumber, req.Month, req.Year)
	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}

func (s *topupStatsStatusByCardCache) GetYearlyTopupStatusFailedByCardNumberCache(ctx context.Context, req *model.FindYearTopupStatusCardNumberInput) (*model.APIResponseTopupYearStatusFailed, bool) {
	key := fmt.Sprintf(yearTopupStatusFailedByCardCacheKey, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTopupYearStatusFailed](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (s *topupStatsStatusByCardCache) SetYearlyTopupStatusFailedByCardNumberCache(ctx context.Context, req *model.FindYearTopupStatusCardNumberInput, data *model.APIResponseTopupYearStatusFailed) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(yearTopupStatusFailedByCardCacheKey, req.CardNumber, req.Year)
	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}
