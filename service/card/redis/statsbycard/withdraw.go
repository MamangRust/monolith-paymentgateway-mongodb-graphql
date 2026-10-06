package cardstatsbycardmencache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type cardStatsWithdrawByCardCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewCardStatsWithdrawByCardCache(store *sharedcachehelpers.CacheStore) CardStatsWithdrawByCardCache {
	return &cardStatsWithdrawByCardCache{store: store}
}

func (c *cardStatsWithdrawByCardCache) GetMonthlyWithdrawByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyWithdrawRow, bool) {
	key := fmt.Sprintf(cacheKeyMonthlyWithdrawByCard, req.CardNumber, req.Year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MonthlyWithdrawRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *cardStatsWithdrawByCardCache) SetMonthlyWithdrawByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard, data []*salemongo.MonthlyWithdrawRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(cacheKeyMonthlyWithdrawByCard, req.CardNumber, req.Year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, expirationCardStatistic)
}

func (c *cardStatsWithdrawByCardCache) GetYearlyWithdrawByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyWithdrawRow, bool) {
	key := fmt.Sprintf(cacheKeyYearlyWithdrawByCard, req.CardNumber, req.Year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.YearlyWithdrawRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *cardStatsWithdrawByCardCache) SetYearlyWithdrawByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard, data []*salemongo.YearlyWithdrawRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(cacheKeyYearlyWithdrawByCard, req.CardNumber, req.Year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, expirationCardStatistic)
}
