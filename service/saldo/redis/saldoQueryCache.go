package mencache

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type saldoCachedResponse struct {
	Data         []*repository.SaldoResult `json:"data"`
	TotalRecords *int                      `json:"total_records"`
}

type saldoQueryCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewSaldoQueryCache(store *sharedcachehelpers.CacheStore) SaldoQueryCache {
	return &saldoQueryCache{store: store}
}

func (s *saldoQueryCache) GetCachedSaldos(ctx context.Context, req *requests.FindAllSaldos) ([]*repository.SaldoResult, *int, bool) {
	key := fmt.Sprintf(saldoAllCacheKey, req.Page, req.PageSize, req.Search)
	result, found := sharedcachehelpers.GetFromCache[saldoCachedResponse](ctx, s.store, key)
	if !found || result == nil {
		return nil, nil, false
	}
	return result.Data, result.TotalRecords, true
}

func (s *saldoQueryCache) SetCachedSaldos(ctx context.Context, req *requests.FindAllSaldos, data []*repository.SaldoResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*repository.SaldoResult{}
	}
	key := fmt.Sprintf(saldoAllCacheKey, req.Page, req.PageSize, req.Search)
	payload := &saldoCachedResponse{Data: data, TotalRecords: total}
	sharedcachehelpers.SetToCache(ctx, s.store, key, payload, ttlDefault)
}

func (s *saldoQueryCache) GetCachedSaldoByActive(ctx context.Context, req *requests.FindAllSaldos) ([]*repository.SaldoResult, *int, bool) {
	key := fmt.Sprintf(saldoActiveCacheKey, req.Page, req.PageSize, req.Search)
	result, found := sharedcachehelpers.GetFromCache[saldoCachedResponse](ctx, s.store, key)
	if !found || result == nil {
		return nil, nil, false
	}
	return result.Data, result.TotalRecords, true
}

func (s *saldoQueryCache) SetCachedSaldoByActive(ctx context.Context, req *requests.FindAllSaldos, data []*repository.SaldoResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*repository.SaldoResult{}
	}
	key := fmt.Sprintf(saldoActiveCacheKey, req.Page, req.PageSize, req.Search)
	payload := &saldoCachedResponse{Data: data, TotalRecords: total}
	sharedcachehelpers.SetToCache(ctx, s.store, key, payload, ttlDefault)
}

func (s *saldoQueryCache) GetCachedSaldoByTrashed(ctx context.Context, req *requests.FindAllSaldos) ([]*repository.SaldoResult, *int, bool) {
	key := fmt.Sprintf(saldoTrashedCacheKey, req.Page, req.PageSize, req.Search)
	result, found := sharedcachehelpers.GetFromCache[saldoCachedResponse](ctx, s.store, key)
	if !found || result == nil {
		return nil, nil, false
	}
	return result.Data, result.TotalRecords, true
}

func (s *saldoQueryCache) SetCachedSaldoByTrashed(ctx context.Context, req *requests.FindAllSaldos, data []*repository.SaldoResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*repository.SaldoResult{}
	}
	key := fmt.Sprintf(saldoTrashedCacheKey, req.Page, req.PageSize, req.Search)
	payload := &saldoCachedResponse{Data: data, TotalRecords: total}
	sharedcachehelpers.SetToCache(ctx, s.store, key, payload, ttlDefault)
}

func (s *saldoQueryCache) GetCachedSaldoById(ctx context.Context, saldo_id int) (*repository.SaldoResult, bool) {
	key := fmt.Sprintf(saldoByIdCacheKey, saldo_id)
	result, found := sharedcachehelpers.GetFromCache[*repository.SaldoResult](ctx, s.store, key)
	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (s *saldoQueryCache) SetCachedSaldoById(ctx context.Context, saldo_id int, data *repository.SaldoResult) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(saldoByIdCacheKey, saldo_id)
	sharedcachehelpers.SetToCache(ctx, s.store, key, data, ttlDefault)
}

func (s *saldoQueryCache) GetCachedSaldoByCardNumber(ctx context.Context, card_number string) (*salemongo.Saldo, bool) {
	key := fmt.Sprintf(saldoByCardNumberKey, card_number)
	result, found := sharedcachehelpers.GetFromCache[*salemongo.Saldo](ctx, s.store, key)
	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (s *saldoQueryCache) SetCachedSaldoByCardNumber(ctx context.Context, card_number string, data *salemongo.Saldo) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(saldoByCardNumberKey, card_number)
	sharedcachehelpers.SetToCache(ctx, s.store, key, data, ttlDefault)
}
