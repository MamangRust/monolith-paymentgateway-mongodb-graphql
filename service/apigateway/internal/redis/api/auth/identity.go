package auth_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type identityCache struct {
	store *cache.CacheStore
}

func NewIdentityCache(store *cache.CacheStore) *identityCache {
	return &identityCache{store: store}
}

func (c *identityCache) SetRefreshToken(
	ctx context.Context,
	token string,
	response *model.APIResponseRefreshToken,
) {
	if response == nil {
		return
	}

	key := fmt.Sprintf(keyIdentityRefreshToken, token)
	cache.SetToCache(ctx, c.store, key, response, ttlDefault)
}

func (c *identityCache) GetRefreshToken(ctx context.Context, token string) (*model.APIResponseRefreshToken, bool) {
	key := fmt.Sprintf(keyIdentityRefreshToken, token)

	result, found := cache.GetFromCache[model.APIResponseRefreshToken](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return result, true
}

func (c *identityCache) DeleteRefreshToken(
	ctx context.Context,
	token string,
) {
	key := fmt.Sprintf(keyIdentityRefreshToken, token)
	cache.DeleteFromCache(ctx, c.store, key)
}

func (c *identityCache) SetCachedUserInfo(ctx context.Context, userId int, data *model.APIResponseGetMe) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(keyIdentityUserInfo, userId)

	cache.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *identityCache) GetCachedUserInfo(ctx context.Context, userId int) (*model.APIResponseGetMe, bool) {
	key := fmt.Sprintf(keyIdentityUserInfo, userId)

	result, found := cache.GetFromCache[model.APIResponseGetMe](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return result, true
}

func (c *identityCache) DeleteCachedUserInfo(ctx context.Context, userId int) {
	key := fmt.Sprintf(keyIdentityUserInfo, userId)
	cache.DeleteFromCache(ctx, c.store, key)
}
