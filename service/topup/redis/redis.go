package mencache

import (
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type Mencache interface {
	TopupQueryCache
	TopupCommandCache
}

type mencache struct {
	TopupQueryCache
	TopupCommandCache
}

// NewMencache creates a new Mencache instance using the given dependencies.
// It creates a new cache store using the given context, Redis client, and logger,
// and returns a Mencache struct with initialized caches for topup query and topup command.
func NewMencache(cacheStore *cache.CacheStore) Mencache {

	return &mencache{
		TopupQueryCache:   NewTopupQueryCache(cacheStore),
		TopupCommandCache: NewTopupCommandCache(cacheStore),
	}
}
