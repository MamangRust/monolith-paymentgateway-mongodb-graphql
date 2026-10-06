package saldo_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type saldoQueryCache struct {
	store *cache.CacheStore
}

func NewSaldoQueryCache(store *cache.CacheStore) SaldoQueryCache {
	return &saldoQueryCache{store: store}
}

func (s *saldoQueryCache) GetCachedSaldos(ctx context.Context, req *model.FindAllSaldoInput) (*model.APIResponsePaginationSaldo, bool) {
	key := fmt.Sprintf(saldoAllCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	result, found := cache.GetFromCache[model.APIResponsePaginationSaldo](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (s *saldoQueryCache) GetCachedSaldoByActive(ctx context.Context, req *model.FindAllSaldoInput) (*model.APIResponsePaginationSaldoDeleteAt, bool) {
	key := fmt.Sprintf(saldoActiveCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	result, found := cache.GetFromCache[model.APIResponsePaginationSaldoDeleteAt](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (s *saldoQueryCache) GetCachedSaldoByTrashed(ctx context.Context, req *model.FindAllSaldoInput) (*model.APIResponsePaginationSaldoDeleteAt, bool) {
	key := fmt.Sprintf(saldoTrashedCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	result, found := cache.GetFromCache[model.APIResponsePaginationSaldoDeleteAt](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (s *saldoQueryCache) GetCachedSaldoById(ctx context.Context, saldo_id int) (*model.APIResponseSaldo, bool) {
	key := fmt.Sprintf(saldoByIdCacheKey, saldo_id)
	result, found := cache.GetFromCache[model.APIResponseSaldo](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (s *saldoQueryCache) GetCachedSaldoByCardNumber(ctx context.Context, card_number string) (*model.APIResponseSaldo, bool) {
	key := fmt.Sprintf(saldoByCardNumberKey, card_number)
	result, found := cache.GetFromCache[model.APIResponseSaldo](ctx, s.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (s *saldoQueryCache) SetCachedSaldos(ctx context.Context, req *model.FindAllSaldoInput, data *model.APIResponsePaginationSaldo) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(saldoAllCacheKey, *req.Page, *req.PageSize, safeString(req.Search))

	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}

func (s *saldoQueryCache) SetCachedSaldoByActive(ctx context.Context, req *model.FindAllSaldoInput, data *model.APIResponsePaginationSaldoDeleteAt) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(saldoActiveCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}

func (s *saldoQueryCache) SetCachedSaldoByTrashed(ctx context.Context, req *model.FindAllSaldoInput, data *model.APIResponsePaginationSaldoDeleteAt) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(saldoTrashedCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}

func (s *saldoQueryCache) SetCachedSaldoById(ctx context.Context, saldo_id int, data *model.APIResponseSaldo) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(saldoByIdCacheKey, saldo_id)
	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}

func (s *saldoQueryCache) SetCachedSaldoByCardNumber(ctx context.Context, card_number string, data *model.APIResponseSaldo) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(saldoByCardNumberKey, card_number)
	cache.SetToCache(ctx, s.store, key, data, ttlDefault)
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
