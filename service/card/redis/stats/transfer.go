package cardstatsmencache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type cardStatsTransferCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewCardStatsTransferCache(store *sharedcachehelpers.CacheStore) CardStatsTransferCache {
	return &cardStatsTransferCache{store: store}
}

func (c *cardStatsTransferCache) GetMonthlyTransferSenderCache(ctx context.Context, year int) ([]*salemongo.MonthlyTransferSentRow, bool) {
	key := fmt.Sprintf(cacheKeyMonthlyTransferSender, year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MonthlyTransferSentRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *cardStatsTransferCache) SetMonthlyTransferSenderCache(ctx context.Context, year int, data []*salemongo.MonthlyTransferSentRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(cacheKeyMonthlyTransferSender, year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, ttlStatistic)
}

func (c *cardStatsTransferCache) GetYearlyTransferSenderCache(ctx context.Context, year int) ([]*salemongo.YearlyTransferSentRow, bool) {
	key := fmt.Sprintf(cacheKeyYearlyTransferSender, year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.YearlyTransferSentRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *cardStatsTransferCache) SetYearlyTransferSenderCache(ctx context.Context, year int, data []*salemongo.YearlyTransferSentRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(cacheKeyYearlyTransferSender, year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, ttlStatistic)
}

func (c *cardStatsTransferCache) GetMonthlyTransferReceiverCache(ctx context.Context, year int) ([]*salemongo.MonthlyTransferReceivedRow, bool) {
	key := fmt.Sprintf(cacheKeyMonthlyTransferReceiver, year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MonthlyTransferReceivedRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *cardStatsTransferCache) SetMonthlyTransferReceiverCache(ctx context.Context, year int, data []*salemongo.MonthlyTransferReceivedRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(cacheKeyMonthlyTransferReceiver, year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, ttlStatistic)
}

func (c *cardStatsTransferCache) GetYearlyTransferReceiverCache(ctx context.Context, year int) ([]*salemongo.YearlyTransferReceivedRow, bool) {
	key := fmt.Sprintf(cacheKeyYearlyTransferReceiver, year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.YearlyTransferReceivedRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *cardStatsTransferCache) SetYearlyTransferReceiverCache(ctx context.Context, year int, data []*salemongo.YearlyTransferReceivedRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(cacheKeyYearlyTransferReceiver, year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, ttlStatistic)
}
