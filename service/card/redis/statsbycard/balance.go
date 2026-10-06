package cardstatsbycardmencache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type cardStatsBalanceByCardCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewCardStatsBalanceByCardCache(store *sharedcachehelpers.CacheStore) CardStatsBalanceByCardCache {
	return &cardStatsBalanceByCardCache{store: store}
}

func (c *cardStatsBalanceByCardCache) GetMonthlyBalanceByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyBalanceRow, bool) {
	key := fmt.Sprintf(cacheKeyMonthlyBalanceByCard, req.CardNumber, req.Year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MonthlyBalanceRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *cardStatsBalanceByCardCache) SetMonthlyBalanceByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard, data []*salemongo.MonthlyBalanceRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(cacheKeyMonthlyBalanceByCard, req.CardNumber, req.Year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, expirationCardStatistic)
}

func (c *cardStatsBalanceByCardCache) GetYearlyBalanceByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyBalanceRow, bool) {
	key := fmt.Sprintf(cacheKeyYearlyBalanceByCard, req.CardNumber, req.Year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.YearlyBalanceRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *cardStatsBalanceByCardCache) SetYearlyBalanceByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard, data []*salemongo.YearlyBalanceRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(cacheKeyYearlyBalanceByCard, req.CardNumber, req.Year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, expirationCardStatistic)
}
