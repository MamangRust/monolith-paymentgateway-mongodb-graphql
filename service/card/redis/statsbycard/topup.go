package cardstatsbycardmencache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type cardStatsTopupByCardCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewCardStatsTopupByCardCache(store *sharedcachehelpers.CacheStore) CardStatsTopupByCardCache {
	return &cardStatsTopupByCardCache{store: store}
}

func (c *cardStatsTopupByCardCache) GetMonthlyTopupByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyTopupRow, bool) {
	key := fmt.Sprintf(cacheKeyMonthlyTopupByCard, req.CardNumber, req.Year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MonthlyTopupRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *cardStatsTopupByCardCache) SetMonthlyTopupByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard, data []*salemongo.MonthlyTopupRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(cacheKeyMonthlyTopupByCard, req.CardNumber, req.Year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, expirationCardStatistic)
}

func (c *cardStatsTopupByCardCache) GetYearlyTopupByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyTopupRow, bool) {
	key := fmt.Sprintf(cacheKeyYearlyTopupByCard, req.CardNumber, req.Year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.YearlyTopupRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *cardStatsTopupByCardCache) SetYearlyTopupByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard, data []*salemongo.YearlyTopupRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(cacheKeyYearlyTopupByCard, req.CardNumber, req.Year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, expirationCardStatistic)
}
