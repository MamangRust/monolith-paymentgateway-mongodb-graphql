package mencache

import (
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type Mencache interface {
	TransactionQueryCache
	TransactionCommandCache
}

// Mencache represents a cache store for transaction queries and commands.
type mencache struct {
	TransactionQueryCache
	TransactionCommandCache
}

func NewMencache(cacheStore *cache.CacheStore) Mencache {
	return &mencache{
		TransactionQueryCache:   NewTransactionQueryCache(cacheStore),
		TransactionCommandCache: NewTransactionCommandCache(cacheStore),
	}
}
