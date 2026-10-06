package auth_cache

import (
	"context"
	"fmt"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

var ttlDefault = 5 * time.Minute

type loginCache struct {
	store *cache.CacheStore
}

func NewLoginCache(store *cache.CacheStore) *loginCache {
	return &loginCache{store: store}
}

func (s *loginCache) GetCachedLogin(ctx context.Context, email string) (*model.APIResponseLogin, bool) {
	key := fmt.Sprintf(keylogin, email)

	result, found := cache.GetFromCache[model.APIResponseLogin](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}

	return result, true
}

func (s *loginCache) SetCachedLogin(ctx context.Context, email string, data *model.APIResponseLogin) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(keylogin, email)

	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}
