package mencache

import (
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type Mencache interface {
	SaldoQueryCache
	SaldoCommandCache
}

type mencache struct {
	SaldoQueryCache
	SaldoCommandCache
}

func NewMencache(cacheStore *cache.CacheStore) Mencache {
	return &mencache{
		SaldoQueryCache:   NewSaldoQueryCache(cacheStore),
		SaldoCommandCache: NewSaldoCommandCache(cacheStore),
	}
}
