package transferstatsbycardcache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type transferStatsByCardAmountCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewTransferStatsByCardAmountCache(store *sharedcachehelpers.CacheStore) TransferStatsByCardAmountCache {
	return &transferStatsByCardAmountCache{store: store}
}

func (t *transferStatsByCardAmountCache) GetMonthlyTransferAmountsBySenderCard(ctx context.Context, req *requests.MonthYearCardNumber) ([]*salemongo.TransferMonthlyAmountBySenderCardRow, bool) {
	key := fmt.Sprintf(transferMonthTransferAmountBySenderCardKey, req.CardNumber, req.Year)
	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TransferMonthlyAmountBySenderCardRow](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (t *transferStatsByCardAmountCache) SetMonthlyTransferAmountsBySenderCard(ctx context.Context, req *requests.MonthYearCardNumber, data []*salemongo.TransferMonthlyAmountBySenderCardRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(transferMonthTransferAmountBySenderCardKey, req.CardNumber, req.Year)
	sharedcachehelpers.SetToCache(ctx, t.store, key, &data, ttlDefault)
}

func (t *transferStatsByCardAmountCache) GetMonthlyTransferAmountsByReceiverCard(ctx context.Context, req *requests.MonthYearCardNumber) ([]*salemongo.TransferMonthlyAmountByReceiverCardRow, bool) {
	key := fmt.Sprintf(transferMonthTransferAmountByReceiverCardKey, req.CardNumber, req.Year)
	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TransferMonthlyAmountByReceiverCardRow](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (t *transferStatsByCardAmountCache) SetMonthlyTransferAmountsByReceiverCard(ctx context.Context, req *requests.MonthYearCardNumber, data []*salemongo.TransferMonthlyAmountByReceiverCardRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(transferMonthTransferAmountByReceiverCardKey, req.CardNumber, req.Year)
	sharedcachehelpers.SetToCache(ctx, t.store, key, &data, ttlDefault)
}

func (t *transferStatsByCardAmountCache) GetYearlyTransferAmountsBySenderCard(ctx context.Context, req *requests.MonthYearCardNumber) ([]*salemongo.TransferYearlyAmountBySenderCardRow, bool) {
	key := fmt.Sprintf(transferYearTransferAmountBySenderCardKey, req.CardNumber, req.Year)
	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TransferYearlyAmountBySenderCardRow](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (t *transferStatsByCardAmountCache) SetYearlyTransferAmountsBySenderCard(ctx context.Context, req *requests.MonthYearCardNumber, data []*salemongo.TransferYearlyAmountBySenderCardRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(transferYearTransferAmountBySenderCardKey, req.CardNumber, req.Year)
	sharedcachehelpers.SetToCache(ctx, t.store, key, &data, ttlDefault)
}

func (t *transferStatsByCardAmountCache) GetYearlyTransferAmountsByReceiverCard(ctx context.Context, req *requests.MonthYearCardNumber) ([]*salemongo.TransferYearlyAmountByReceiverCardRow, bool) {
	key := fmt.Sprintf(transferYearTransferAmountByReceiverCardKey, req.CardNumber, req.Year)
	result, found := sharedcachehelpers.GetFromCache[[]*salemongo.TransferYearlyAmountByReceiverCardRow](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (t *transferStatsByCardAmountCache) SetYearlyTransferAmountsByReceiverCard(ctx context.Context, req *requests.MonthYearCardNumber, data []*salemongo.TransferYearlyAmountByReceiverCardRow) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(transferYearTransferAmountByReceiverCardKey, req.CardNumber, req.Year)
	sharedcachehelpers.SetToCache(ctx, t.store, key, &data, ttlDefault)
}
