package mencache

import (
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type mencache struct {
	MerchantQueryCache
	MerchantCommandCache
	MerchantDocumentQueryCache
	MerchantDocumentCommandCache
	MerchantTransactionCache
}

type Mencache interface {
	MerchantQueryCache
	MerchantCommandCache
	MerchantDocumentQueryCache
	MerchantDocumentCommandCache
	MerchantTransactionCache
}

func NewMencache(cacheStore *cache.CacheStore) Mencache {
	return &mencache{
		MerchantQueryCache:           NewMerchantQueryCache(cacheStore),
		MerchantCommandCache:         NewMerchantCommandCache(cacheStore),
		MerchantDocumentQueryCache:   NewMerchantDocumentQueryCache(cacheStore),
		MerchantDocumentCommandCache: NewMerchantDocumentCommandCache(cacheStore),
		MerchantTransactionCache:     NewMerchantTransactionCache(cacheStore),
	}
}
