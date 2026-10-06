package role_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type roleQueryCache struct {
	store *cache.CacheStore
}

func NewRoleQueryCache(store *cache.CacheStore) RoleQueryCache {
	return &roleQueryCache{store: store}
}

func (r *roleQueryCache) SetCachedRoles(ctx context.Context, req *model.FindAllRoleInput, data *model.APIResponsePaginationRole) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(roleAllCacheKey, *req.Page, *req.PageSize, safeString(req.Search))

	cache.SetToCache(ctx, r.store, key, data, ttlDefault)
}

func (r *roleQueryCache) GetCachedRoles(ctx context.Context, req *model.FindAllRoleInput) (*model.APIResponsePaginationRole, bool) {
	key := fmt.Sprintf(roleAllCacheKey, *req.Page, *req.PageSize, safeString(req.Search))

	result, found := cache.GetFromCache[model.APIResponsePaginationRole](ctx, r.store, key)

	if !found || result == nil {
		return nil, false
	}

	return result, true
}

func (r *roleQueryCache) SetCachedRoleById(ctx context.Context, id int, data *model.APIResponseRole) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(roleByIdCacheKey, id)

	cache.SetToCache(ctx, r.store, key, data, ttlDefault)
}

func (r *roleQueryCache) GetCachedRoleById(ctx context.Context, id int) (*model.APIResponseRole, bool) {
	key := fmt.Sprintf(roleByIdCacheKey, id)

	result, found := cache.GetFromCache[model.APIResponseRole](ctx, r.store, key)

	if !found || result == nil {
		return nil, false
	}

	return result, true
}

func (r *roleQueryCache) SetCachedRoleByUserId(ctx context.Context, userId int, data *model.APIResponsesRole) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(roleByUserIdCacheKey, userId)

	cache.SetToCache(ctx, r.store, key, data, ttlDefault)
}

func (r *roleQueryCache) GetCachedRoleByUserId(ctx context.Context, userId int) (*model.APIResponsesRole, bool) {
	key := fmt.Sprintf(roleByUserIdCacheKey, userId)

	result, found := cache.GetFromCache[model.APIResponsesRole](ctx, r.store, key)

	if !found || result == nil {
		return nil, false
	}

	return result, true
}

func (r *roleQueryCache) SetCachedRoleActive(ctx context.Context, req *model.FindAllRoleInput, data *model.APIResponsePaginationRoleDeleteAt) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(roleActiveCacheKey, *req.Page, *req.PageSize, safeString(req.Search))

	cache.SetToCache(ctx, r.store, key, data, ttlDefault)
}

func (r *roleQueryCache) GetCachedRoleActive(ctx context.Context, req *model.FindAllRoleInput) (*model.APIResponsePaginationRoleDeleteAt, bool) {
	key := fmt.Sprintf(roleActiveCacheKey, *req.Page, *req.PageSize, safeString(req.Search))

	result, found := cache.GetFromCache[model.APIResponsePaginationRoleDeleteAt](ctx, r.store, key)

	if !found || result == nil {
		return nil, false
	}

	return result, true
}

func (r *roleQueryCache) SetCachedRoleTrashed(ctx context.Context, req *model.FindAllRoleInput, data *model.APIResponsePaginationRoleDeleteAt) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(roleTrashedCacheKey, *req.Page, *req.PageSize, safeString(req.Search))

	cache.SetToCache(ctx, r.store, key, data, ttlDefault)
}

func (r *roleQueryCache) GetCachedRoleTrashed(ctx context.Context, req *model.FindAllRoleInput) (*model.APIResponsePaginationRoleDeleteAt, bool) {
	key := fmt.Sprintf(roleTrashedCacheKey, *req.Page, *req.PageSize, safeString(req.Search))

	result, found := cache.GetFromCache[model.APIResponsePaginationRoleDeleteAt](ctx, r.store, key)

	if !found || result == nil {
		return nil, false
	}

	return result, true
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
