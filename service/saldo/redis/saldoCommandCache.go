package mencache

import (
	"context"
	"fmt"

	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

// saldoCommandCache is a struct that represents the cache store
type saldoCommandCache struct {
	store *sharedcachehelpers.CacheStore
}

// NewSaldoCommandCache creates a new instance of saldoCommandCache with the provided CacheStore.
// It returns a pointer to the newly created saldoCommandCache.
func NewSaldoCommandCache(store *sharedcachehelpers.CacheStore) SaldoCommandCache {
	return &saldoCommandCache{store: store}
}

// DeleteSaldoCache removes the cache entry associated with the specified saldo ID.
// It formats the cache key using the saldo ID and deletes the entry from the cache store.
func (s *saldoCommandCache) DeleteSaldoCache(ctx context.Context, saldo_id int) {
	key := fmt.Sprintf(saldoByIdCacheKey, saldo_id)
	sharedcachehelpers.DeleteFromCache(ctx, s.store, key)
}

// DeleteSaldoCacheByCardNumber removes the card-number cache entry. Deleting a
// missing key is a no-op, so this is safe to call repeatedly (replay-safe).
func (s *saldoCommandCache) DeleteSaldoCacheByCardNumber(ctx context.Context, card_number string) {
	key := fmt.Sprintf(saldoByCardNumberKey, card_number)
	sharedcachehelpers.DeleteFromCache(ctx, s.store, key)
}
