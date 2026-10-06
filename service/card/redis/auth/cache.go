package cardauthmencache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type cardAuthCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewCardAuthCache(store *sharedcachehelpers.CacheStore) CardAuthCache {
	return &cardAuthCache{store: store}
}

func (c *cardAuthCache) GetByTxnID(ctx context.Context, txnID string) (*salemongo.CardAuthTransaction, bool) {
	key := fmt.Sprintf(authTxnByTxnIDCacheKey, txnID)
	result, found := sharedcachehelpers.GetFromCache[*salemongo.CardAuthTransaction](ctx, c.store, key)
	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (c *cardAuthCache) SetByTxnID(ctx context.Context, txnID string, data *salemongo.CardAuthTransaction) {
	key := fmt.Sprintf(authTxnByTxnIDCacheKey, txnID)
	sharedcachehelpers.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *cardAuthCache) DeleteByTxnID(ctx context.Context, txnID string) {
	key := fmt.Sprintf(authTxnByTxnIDCacheKey, txnID)
	sharedcachehelpers.DeleteFromCache(ctx, c.store, key)
}
