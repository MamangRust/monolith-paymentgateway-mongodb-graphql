package mencache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-topup/repository"
)

type topupCachedResponseAll struct {
	Data  []*repository.TopupQueryResult `json:"data"`
	Total *int                           `json:"total_records"`
}

type topupCachedResponseByCard struct {
	Data  []*repository.TopupByCardResult `json:"data"`
	Total *int                            `json:"total_records"`
}

type topupCachedResponseActive struct {
	Data  []*repository.TopupQueryResult `json:"data"`
	Total *int                           `json:"total_records"`
}

type topupCachedResponseTrashed struct {
	Data  []*repository.TopupQueryResult `json:"data"`
	Total *int                           `json:"total_records"`
}

type topupQueryCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewTopupQueryCache(store *sharedcachehelpers.CacheStore) TopupQueryCache {
	return &topupQueryCache{store: store}
}

func (c *topupQueryCache) GetCachedTopupsCache(ctx context.Context, req *requests.FindAllTopups) ([]*repository.TopupQueryResult, *int, bool) {
	key := fmt.Sprintf(topupAllCacheKey, req.Page, req.PageSize, req.Search)

	result, found := sharedcachehelpers.GetFromCache[topupCachedResponseAll](ctx, c.store, key)

	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.Total, true
}

func (c *topupQueryCache) SetCachedTopupsCache(ctx context.Context, req *requests.FindAllTopups, data []*repository.TopupQueryResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*repository.TopupQueryResult{}
	}

	key := fmt.Sprintf(topupAllCacheKey, req.Page, req.PageSize, req.Search)
	payload := &topupCachedResponseAll{Data: data, Total: total}
	sharedcachehelpers.SetToCache(ctx, c.store, key, payload, ttlDefault)
}

func (c *topupQueryCache) GetCacheTopupByCardCache(ctx context.Context, req *requests.FindAllTopupsByCardNumber) ([]*repository.TopupByCardResult, *int, bool) {
	key := fmt.Sprintf(topupByCardCacheKey, req.CardNumber, req.Page, req.PageSize, req.Search)

	result, found := sharedcachehelpers.GetFromCache[topupCachedResponseByCard](ctx, c.store, key)

	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.Total, true
}

func (c *topupQueryCache) SetCacheTopupByCardCache(ctx context.Context, req *requests.FindAllTopupsByCardNumber, data []*repository.TopupByCardResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*repository.TopupByCardResult{}
	}

	key := fmt.Sprintf(topupByCardCacheKey, req.CardNumber, req.Page, req.PageSize, req.Search)
	payload := &topupCachedResponseByCard{Data: data, Total: total}
	sharedcachehelpers.SetToCache(ctx, c.store, key, payload, ttlDefault)
}

func (c *topupQueryCache) GetCachedTopupActiveCache(ctx context.Context, req *requests.FindAllTopups) ([]*repository.TopupQueryResult, *int, bool) {
	key := fmt.Sprintf(topupActiveCacheKey, req.Page, req.PageSize, req.Search)

	result, found := sharedcachehelpers.GetFromCache[topupCachedResponseActive](ctx, c.store, key)

	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.Total, true
}

func (c *topupQueryCache) SetCachedTopupActiveCache(ctx context.Context, req *requests.FindAllTopups, data []*repository.TopupQueryResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*repository.TopupQueryResult{}
	}

	key := fmt.Sprintf(topupActiveCacheKey, req.Page, req.PageSize, req.Search)
	payload := &topupCachedResponseActive{Data: data, Total: total}
	sharedcachehelpers.SetToCache(ctx, c.store, key, payload, ttlDefault)
}

func (c *topupQueryCache) GetCachedTopupTrashedCache(ctx context.Context, req *requests.FindAllTopups) ([]*repository.TopupQueryResult, *int, bool) {
	key := fmt.Sprintf(topupTrashedCacheKey, req.Page, req.PageSize, req.Search)

	result, found := sharedcachehelpers.GetFromCache[topupCachedResponseTrashed](ctx, c.store, key)

	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.Total, true
}

func (c *topupQueryCache) SetCachedTopupTrashedCache(ctx context.Context, req *requests.FindAllTopups, data []*repository.TopupQueryResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*repository.TopupQueryResult{}
	}

	key := fmt.Sprintf(topupTrashedCacheKey, req.Page, req.PageSize, req.Search)
	payload := &topupCachedResponseTrashed{Data: data, Total: total}
	sharedcachehelpers.SetToCache(ctx, c.store, key, payload, ttlDefault)
}

func (c *topupQueryCache) GetCachedTopupCache(ctx context.Context, id int) (*salemongo.Topup, bool) {
	key := fmt.Sprintf(topupByIdCacheKey, id)

	result, found := sharedcachehelpers.GetFromCache[*salemongo.Topup](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *topupQueryCache) SetCachedTopupCache(ctx context.Context, data *salemongo.Topup) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(topupByIdCacheKey, data.TopupID)
	sharedcachehelpers.SetToCache(ctx, c.store, key, data, ttlDefault)
}
