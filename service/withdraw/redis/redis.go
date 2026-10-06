package mencache

import (
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

// Mencache represents a cache store for withdraw operations.
type Mencache interface {
	WithdrawQueryCache
	WithdrawCommandCache
}

type mencache struct {
	WithdrawQueryCache
	WithdrawCommandCache
}

func NewMencache(cacheStore *cache.CacheStore) Mencache {
	return &mencache{
		WithdrawQueryCache:   NewWithdrawQueryCache(cacheStore),
		WithdrawCommandCache: NewWithdrawCommandCache(cacheStore),
	}
}
