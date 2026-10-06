package transfer_stats_bycard_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type transferStatsByCardAmountCache struct {
	store *cache.CacheStore
}

func NewTransferStatsByCardAmountCache(store *cache.CacheStore) TransferStatsByCardAmountCache {
	return &transferStatsByCardAmountCache{store: store}
}

func (t *transferStatsByCardAmountCache) GetMonthlyTransferAmountsBySenderCard(ctx context.Context, req *model.FindByCardNumberTransferInput) (*model.APIResponseTransferMonthAmount, bool) {
	key := fmt.Sprintf(transferMonthTransferAmountBySenderCardKey, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTransferMonthAmount](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transferStatsByCardAmountCache) SetMonthlyTransferAmountsBySenderCard(ctx context.Context, req *model.FindByCardNumberTransferInput, data *model.APIResponseTransferMonthAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(transferMonthTransferAmountBySenderCardKey, req.CardNumber, req.Year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transferStatsByCardAmountCache) GetMonthlyTransferAmountsByReceiverCard(ctx context.Context, req *model.FindByCardNumberTransferInput) (*model.APIResponseTransferMonthAmount, bool) {
	key := fmt.Sprintf(transferMonthTransferAmountByReceiverCardKey, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTransferMonthAmount](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transferStatsByCardAmountCache) SetMonthlyTransferAmountsByReceiverCard(ctx context.Context, req *model.FindByCardNumberTransferInput, data *model.APIResponseTransferMonthAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(transferMonthTransferAmountByReceiverCardKey, req.CardNumber, req.Year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transferStatsByCardAmountCache) GetYearlyTransferAmountsBySenderCard(ctx context.Context, req *model.FindByCardNumberTransferInput) (*model.APIResponseTransferYearAmount, bool) {
	key := fmt.Sprintf(transferYearTransferAmountBySenderCardKey, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTransferYearAmount](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transferStatsByCardAmountCache) SetYearlyTransferAmountsBySenderCard(ctx context.Context, req *model.FindByCardNumberTransferInput, data *model.APIResponseTransferYearAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(transferYearTransferAmountBySenderCardKey, req.CardNumber, req.Year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}

func (t *transferStatsByCardAmountCache) GetYearlyTransferAmountsByReceiverCard(ctx context.Context, req *model.FindByCardNumberTransferInput) (*model.APIResponseTransferYearAmount, bool) {
	key := fmt.Sprintf(transferYearTransferAmountByReceiverCardKey, req.CardNumber, req.Year)
	result, found := cache.GetFromCache[model.APIResponseTransferYearAmount](ctx, t.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (t *transferStatsByCardAmountCache) SetYearlyTransferAmountsByReceiverCard(ctx context.Context, req *model.FindByCardNumberTransferInput, data *model.APIResponseTransferYearAmount) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(transferYearTransferAmountByReceiverCardKey, req.CardNumber, req.Year)
	cache.SetToCache(ctx, t.store, key, data, ttlDefault)
}
