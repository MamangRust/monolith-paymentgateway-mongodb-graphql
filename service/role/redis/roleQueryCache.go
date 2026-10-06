package mencache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type roleCachedResponseAll struct {
	Data         []*salemongo.Role `json:"data"`
	TotalRecords *int              `json:"total_records"`
}

type roleQueryCache struct {
	store *cache.CacheStore
}

func NewRoleQueryCache(store *cache.CacheStore) RoleQueryCache {
	return &roleQueryCache{store: store}
}

func (r *roleQueryCache) SetCachedRoles(ctx context.Context, req *requests.FindAllRoles, data []*salemongo.Role, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*salemongo.Role{}
	}

	key := fmt.Sprintf(roleAllCacheKey, req.Page, req.PageSize, req.Search)
	payload := &roleCachedResponseAll{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, r.store, key, payload, ttlDefault)
}

func (r *roleQueryCache) GetCachedRoles(ctx context.Context, req *requests.FindAllRoles) ([]*salemongo.Role, *int, bool) {
	key := fmt.Sprintf(roleAllCacheKey, req.Page, req.PageSize, req.Search)

	result, found := cache.GetFromCache[roleCachedResponseAll](ctx, r.store, key)

	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.TotalRecords, true
}

func (r *roleQueryCache) SetCachedRoleById(ctx context.Context, id int, data *salemongo.Role) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(roleByIdCacheKey, id)
	cache.SetToCache(ctx, r.store, key, data, ttlDefault)
}

func (r *roleQueryCache) GetCachedRoleById(ctx context.Context, id int) (*salemongo.Role, bool) {
	key := fmt.Sprintf(roleByIdCacheKey, id)

	result, found := cache.GetFromCache[*salemongo.Role](ctx, r.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (r *roleQueryCache) SetCachedRoleByUserId(ctx context.Context, userId int, data []*salemongo.Role) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(roleByUserIdCacheKey, userId)
	cache.SetToCache(ctx, r.store, key, &data, ttlDefault)
}

func (r *roleQueryCache) GetCachedRoleByUserId(ctx context.Context, userId int) ([]*salemongo.Role, bool) {
	key := fmt.Sprintf(roleByUserIdCacheKey, userId)

	result, found := cache.GetFromCache[[]*salemongo.Role](ctx, r.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (r *roleQueryCache) SetCachedRoleActive(ctx context.Context, req *requests.FindAllRoles, data []*salemongo.Role, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*salemongo.Role{}
	}

	key := fmt.Sprintf(roleActiveCacheKey, req.Page, req.PageSize, req.Search)
	payload := &roleCachedResponseAll{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, r.store, key, payload, ttlDefault)
}

func (r *roleQueryCache) GetCachedRoleActive(ctx context.Context, req *requests.FindAllRoles) ([]*salemongo.Role, *int, bool) {
	key := fmt.Sprintf(roleActiveCacheKey, req.Page, req.PageSize, req.Search)

	result, found := cache.GetFromCache[roleCachedResponseAll](ctx, r.store, key)

	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.TotalRecords, true
}

func (r *roleQueryCache) SetCachedRoleTrashed(ctx context.Context, req *requests.FindAllRoles, data []*salemongo.Role, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*salemongo.Role{}
	}

	key := fmt.Sprintf(roleTrashedCacheKey, req.Page, req.PageSize, req.Search)
	payload := &roleCachedResponseAll{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, r.store, key, payload, ttlDefault)
}

func (r *roleQueryCache) GetCachedRoleTrashed(ctx context.Context, req *requests.FindAllRoles) ([]*salemongo.Role, *int, bool) {
	key := fmt.Sprintf(roleTrashedCacheKey, req.Page, req.PageSize, req.Search)

	result, found := cache.GetFromCache[roleCachedResponseAll](ctx, r.store, key)

	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.TotalRecords, true
}

func (r *roleQueryCache) SetCachedRoleByName(ctx context.Context, name string, data *salemongo.Role) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(roleByNameCacheKey, name)
	cache.SetToCache(ctx, r.store, key, data, ttlDefault)
}

func (r *roleQueryCache) GetCachedRoleByName(ctx context.Context, name string) (*salemongo.Role, bool) {
	key := fmt.Sprintf(roleByNameCacheKey, name)

	result, found := cache.GetFromCache[*salemongo.Role](ctx, r.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}
