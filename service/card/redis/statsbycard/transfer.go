package cardstatsbycardmencache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type cardStatsTransferByCardCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewCardStatsTransferByCardCache(store *sharedcachehelpers.CacheStore) CardStatsTransferByCardCache {
	return &cardStatsTransferByCardCache{store: store}
}

func (c *cardStatsTransferByCardCache) GetMonthlyTransferBySenderCache(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyTransferSentRow, bool) {
	key := fmt.Sprintf(cacheKeyMonthlySenderByCard, req.CardNumber, req.Year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MonthlyTransferSentRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *cardStatsTransferByCardCache) SetMonthlyTransferBySenderCache(ctx context.Context, req *requests.MonthYearCardNumberCard, data []*salemongo.MonthlyTransferSentRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(cacheKeyMonthlySenderByCard, req.CardNumber, req.Year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, expirationCardStatistic)
}

func (c *cardStatsTransferByCardCache) GetYearlyTransferBySenderCache(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyTransferSentRow, bool) {
	key := fmt.Sprintf(cacheKeyYearlySenderByCard, req.CardNumber, req.Year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.YearlyTransferSentRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *cardStatsTransferByCardCache) SetYearlyTransferBySenderCache(ctx context.Context, req *requests.MonthYearCardNumberCard, data []*salemongo.YearlyTransferSentRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(cacheKeyYearlySenderByCard, req.CardNumber, req.Year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, expirationCardStatistic)
}

func (c *cardStatsTransferByCardCache) GetMonthlyTransferByReceiverCache(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyTransferReceivedRow, bool) {
	key := fmt.Sprintf(cacheKeyMonthlyReceiverByCard, req.CardNumber, req.Year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.MonthlyTransferReceivedRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *cardStatsTransferByCardCache) SetMonthlyTransferByReceiverCache(ctx context.Context, req *requests.MonthYearCardNumberCard, data []*salemongo.MonthlyTransferReceivedRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(cacheKeyMonthlyReceiverByCard, req.CardNumber, req.Year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, expirationCardStatistic)
}

func (c *cardStatsTransferByCardCache) GetYearlyTransferByReceiverCache(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyTransferReceivedRow, bool) {
	key := fmt.Sprintf(cacheKeyYearlyReceiverByCard, req.CardNumber, req.Year)

	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.YearlyTransferReceivedRow](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *cardStatsTransferByCardCache) SetYearlyTransferByReceiverCache(ctx context.Context, req *requests.MonthYearCardNumberCard, data []*salemongo.YearlyTransferReceivedRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(cacheKeyYearlyReceiverByCard, req.CardNumber, req.Year)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, expirationCardStatistic)
}
