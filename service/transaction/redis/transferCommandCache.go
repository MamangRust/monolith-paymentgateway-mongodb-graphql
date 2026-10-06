package mencache

import (
	"context"
	"fmt"

	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

// transactionCommandCache represents a cache for transaction commands.
type transactionCommandCache struct {
	store *sharedcachehelpers.CacheStore
}

// NewTransactionCommandCache creates a new instance of transactionCommandCache with the provided CacheStore.
// It returns a pointer to the newly created transactionCommandCache.
func NewTransactionCommandCache(store *sharedcachehelpers.CacheStore) TransactionCommandCache {
	return &transactionCommandCache{store: store}
}

// DeleteTransactionCache removes a cached transaction entry by its ID.
//
// Parameters:
//   - ctx: The context for timeout and cancellation.
//   - id: The transaction ID whose cache entry should be deleted.
func (t *transactionCommandCache) DeleteTransactionCache(ctx context.Context, id int) {
	sharedcachehelpers.DeleteFromCache(ctx, t.store, fmt.Sprintf(transactionByIdCacheKey, id))
}
