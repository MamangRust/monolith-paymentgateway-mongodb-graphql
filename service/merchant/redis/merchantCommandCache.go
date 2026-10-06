package mencache

import (
	"context"
	"fmt"

	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

// merchantCommandCache is a struct that represents the cache store
type merchantCommandCache struct {
	store *sharedcachehelpers.CacheStore
}

// NewMerchantCommandCache returns a new instance of merchantCommandCache
func NewMerchantCommandCache(store *sharedcachehelpers.CacheStore) MerchantCommandCache {
	return &merchantCommandCache{store: store}
}

// DeleteCachedMerchant removes the cache entry associated with the specified merchant ID.
// It formats the cache key using the merchant ID and deletes the entry from the cache store.
func (s *merchantCommandCache) DeleteCachedMerchant(ctx context.Context, id int) {
	key := fmt.Sprintf(merchantByIdCacheKey, id)
	sharedcachehelpers.DeleteFromCache(ctx, s.store, key)
}

func (s *merchantCommandCache) InvalidateMerchantListCaches(ctx context.Context) {
	for _, pattern := range []string{"merchant:all:page:*", "merchant:active:page:*", "merchant:trashed:page:*", "merchant:user_id:*", "merchant:api_key:*"} {
		sharedcachehelpers.DeleteFromCachePattern(ctx, s.store, pattern)
	}
}

func (s *merchantCommandCache) DeleteCachedMerchantByUserID(ctx context.Context, userId int) {
	key := fmt.Sprintf(merchantByUserIdCacheKey, userId)
	sharedcachehelpers.DeleteFromCache(ctx, s.store, key)
}

func (s *merchantCommandCache) DeleteCachedMerchantByAPIKey(ctx context.Context, apiKey string) {
	key := fmt.Sprintf(merchantByApiKeyCacheKey, apiKey)
	sharedcachehelpers.DeleteFromCache(ctx, s.store, key)
}
