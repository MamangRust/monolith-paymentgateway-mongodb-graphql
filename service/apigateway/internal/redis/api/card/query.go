package card_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type cardQueryCache struct {
	store *cache.CacheStore
}

func NewCardQueryCache(store *cache.CacheStore) CardQueryCache {
	return &cardQueryCache{store: store}
}

func (c *cardQueryCache) GetByIdCache(ctx context.Context, cardID int) (*model.APIResponseCard, bool) {
	key := fmt.Sprintf(cardByIdCacheKey, cardID)
	result, found := cache.GetFromCache[*model.APIResponseCard](ctx, c.store, key)
	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (c *cardQueryCache) GetFindAllCache(ctx context.Context, req *model.FindAllCardInput) (*model.APIResponsePaginationCard, bool) {
	key := fmt.Sprintf(cardAllCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	result, found := cache.GetFromCache[model.APIResponsePaginationCard](ctx, c.store, key)
	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *cardQueryCache) GetByActiveCache(ctx context.Context, req *model.FindAllCardInput) (*model.APIResponsePaginationCardDeleteAt, bool) {
	key := fmt.Sprintf(cardActiveCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	result, found := cache.GetFromCache[model.APIResponsePaginationCardDeleteAt](ctx, c.store, key)
	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *cardQueryCache) GetByTrashedCache(ctx context.Context, req *model.FindAllCardInput) (*model.APIResponsePaginationCardDeleteAt, bool) {
	key := fmt.Sprintf(cardTrashedCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	result, found := cache.GetFromCache[model.APIResponsePaginationCardDeleteAt](ctx, c.store, key)
	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (c *cardQueryCache) GetByUserIDCache(ctx context.Context, userID int) (*model.APIResponseCard, bool) {
	key := fmt.Sprintf(cardByUserIdCacheKey, userID)
	result, found := cache.GetFromCache[*model.APIResponseCard](ctx, c.store, key)
	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (c *cardQueryCache) GetByCardNumberCache(
	ctx context.Context,
	cardNumber string,
) (*model.APIResponseCard, bool) {

	key := fmt.Sprintf(cardByCardNumCacheKey, cardNumber)

	result, found := cache.GetFromCache[model.APIResponseCard](ctx, c.store, key)
	if !found || result == nil {
		return nil, false
	}

	return result, true
}

func (c *cardQueryCache) SetByIdCache(ctx context.Context, cardID int, data *model.APIResponseCard) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cardByIdCacheKey, cardID)
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *cardQueryCache) SetByCardNumberCache(
	ctx context.Context,
	cardNumber string,
	data *model.APIResponseCard,
) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(cardByCardNumCacheKey, cardNumber)
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *cardQueryCache) SetFindAllCache(ctx context.Context, req *model.FindAllCardInput, data *model.APIResponsePaginationCard) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cardAllCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *cardQueryCache) SetByActiveCache(ctx context.Context, req *model.FindAllCardInput, data *model.APIResponsePaginationCardDeleteAt) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cardActiveCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *cardQueryCache) SetByTrashedCache(ctx context.Context, req *model.FindAllCardInput, data *model.APIResponsePaginationCardDeleteAt) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cardTrashedCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *cardQueryCache) SetByUserIDCache(ctx context.Context, userID int, data *model.APIResponseCard) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cardByUserIdCacheKey, userID)
	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *cardQueryCache) DeleteByIdCache(ctx context.Context, cardID int) {
	key := fmt.Sprintf(cardByIdCacheKey, cardID)
	cache.DeleteFromCache(ctx, c.store, key)
}

func (c *cardQueryCache) DeleteByUserIDCache(ctx context.Context, userID int) {
	key := fmt.Sprintf(cardByUserIdCacheKey, userID)
	cache.DeleteFromCache(ctx, c.store, key)
}

func (c *cardQueryCache) DeleteByCardNumberCache(ctx context.Context, cardNumber string) {
	key := fmt.Sprintf(cardByCardNumCacheKey, cardNumber)
	cache.DeleteFromCache(ctx, c.store, key)
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
