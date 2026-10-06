package user_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type userQueryCache struct {
	store *cache.CacheStore
}

func NewUserQueryCache(store *cache.CacheStore) UserQueryCache {
	return &userQueryCache{store: store}
}

func (s *userQueryCache) GetCachedUsersCache(ctx context.Context, req *model.FindAllUserInput) (*model.APIResponsePaginationUser, bool) {
	key := fmt.Sprintf(userAllCacheKey, *req.Page, *req.PageSize, safeString(req.Search))

	result, found := cache.GetFromCache[model.APIResponsePaginationUser](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}

	return result, true
}

func (s *userQueryCache) SetCachedUsersCache(ctx context.Context, req *model.FindAllUserInput, data *model.APIResponsePaginationUser) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(userAllCacheKey, *req.Page, *req.PageSize, safeString(req.Search))

	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}

func (s *userQueryCache) GetCachedUserActiveCache(ctx context.Context, req *model.FindAllUserInput) (*model.APIResponsePaginationUserDeleteAt, bool) {
	key := fmt.Sprintf(userActiveCacheKey, *req.Page, *req.PageSize, safeString(req.Search))

	result, found := cache.GetFromCache[model.APIResponsePaginationUserDeleteAt](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}

	return result, true
}

func (s *userQueryCache) SetCachedUserActiveCache(ctx context.Context, req *model.FindAllUserInput, data *model.APIResponsePaginationUserDeleteAt) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(userActiveCacheKey, *req.Page, *req.PageSize, safeString(req.Search))

	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}

func (s *userQueryCache) GetCachedUserTrashedCache(ctx context.Context, req *model.FindAllUserInput) (*model.APIResponsePaginationUserDeleteAt, bool) {
	key := fmt.Sprintf(userTrashedCacheKey, *req.Page, *req.PageSize, safeString(req.Search))

	result, found := cache.GetFromCache[model.APIResponsePaginationUserDeleteAt](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}

	return result, true
}

func (s *userQueryCache) SetCachedUserTrashedCache(ctx context.Context, req *model.FindAllUserInput, data *model.APIResponsePaginationUserDeleteAt) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(userTrashedCacheKey, *req.Page, *req.PageSize, safeString(req.Search))

	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}

func (s *userQueryCache) GetCachedUserCache(ctx context.Context, id int) (*model.APIResponseUser, bool) {
	key := fmt.Sprintf(userByIdCacheKey, id)

	result, found := cache.GetFromCache[model.APIResponseUser](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}

	return result, true
}

func (s *userQueryCache) SetCachedUserCache(ctx context.Context, data *model.APIResponseUser) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(userByIdCacheKey, data.Data.ID)
	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
