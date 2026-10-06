package mencache

import (
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type Mencache interface {
	TransferQueryCache
	TransferCommandCache
}

type mencache struct {
	TransferQueryCache
	TransferCommandCache
}

func NewMencache(cacheStore *cache.CacheStore) Mencache {
	return &mencache{
		TransferQueryCache:   NewTransferQueryCache(cacheStore),
		TransferCommandCache: NewTransferCommandCache(cacheStore),
	}
}
