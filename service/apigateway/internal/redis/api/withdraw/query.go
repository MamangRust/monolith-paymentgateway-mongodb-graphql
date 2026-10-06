package withdraw_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type withdrawQueryCache struct {
	store *cache.CacheStore
}

func NewWithdrawQueryCache(store *cache.CacheStore) WithdrawQueryCache {
	return &withdrawQueryCache{store: store}
}

func (w *withdrawQueryCache) GetCachedWithdrawsCache(ctx context.Context, req *model.FindAllWithdrawInput) (*model.APIResponsePaginationWithdraw, bool) {
	key := fmt.Sprintf(withdrawAllCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	result, found := cache.GetFromCache[model.APIResponsePaginationWithdraw](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (w *withdrawQueryCache) GetCachedWithdrawByCardCache(ctx context.Context, req *model.FindAllWithdrawByCardNumberInput) (*model.APIResponsePaginationWithdraw, bool) {
	key := fmt.Sprintf(withdrawByCardCacheKey, req.CardNumber, *req.Page, *req.PageSize, safeString(req.Search))
	result, found := cache.GetFromCache[model.APIResponsePaginationWithdraw](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (w *withdrawQueryCache) GetCachedWithdrawActiveCache(ctx context.Context, req *model.FindAllWithdrawInput) (*model.APIResponsePaginationWithdrawDeleteAt, bool) {
	key := fmt.Sprintf(withdrawActiveCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	result, found := cache.GetFromCache[model.APIResponsePaginationWithdrawDeleteAt](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (w *withdrawQueryCache) GetCachedWithdrawTrashedCache(ctx context.Context, req *model.FindAllWithdrawInput) (*model.APIResponsePaginationWithdrawDeleteAt, bool) {
	key := fmt.Sprintf(withdrawTrashedCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	result, found := cache.GetFromCache[model.APIResponsePaginationWithdrawDeleteAt](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (w *withdrawQueryCache) GetCachedWithdrawCache(ctx context.Context, id int) (*model.APIResponseWithdraw, bool) {
	key := fmt.Sprintf(withdrawByIdCacheKey, id)
	result, found := cache.GetFromCache[model.APIResponseWithdraw](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (w *withdrawQueryCache) SetCachedWithdrawsCache(ctx context.Context, req *model.FindAllWithdrawInput, data *model.APIResponsePaginationWithdraw) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(withdrawAllCacheKey, *req.Page, *req.PageSize, safeString(req.Search))

	cache.SetToCache(ctx, w.store, key, data, ttlDefault)
}

func (w *withdrawQueryCache) SetCachedWithdrawByCardCache(ctx context.Context, req *model.FindAllWithdrawByCardNumberInput, data *model.APIResponsePaginationWithdraw) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(withdrawByCardCacheKey, req.CardNumber, *req.Page, *req.PageSize, safeString(req.Search))
	cache.SetToCache(ctx, w.store, key, data, ttlDefault)
}

func (w *withdrawQueryCache) SetCachedWithdrawActiveCache(ctx context.Context, req *model.FindAllWithdrawInput, data *model.APIResponsePaginationWithdrawDeleteAt) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(withdrawActiveCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	cache.SetToCache(ctx, w.store, key, data, ttlDefault)
}

func (w *withdrawQueryCache) SetCachedWithdrawTrashedCache(ctx context.Context, req *model.FindAllWithdrawInput, data *model.APIResponsePaginationWithdrawDeleteAt) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(withdrawTrashedCacheKey, *req.Page, *req.PageSize, safeString(req.Search))
	cache.SetToCache(ctx, w.store, key, data, ttlDefault)
}

func (w *withdrawQueryCache) SetCachedWithdrawCache(ctx context.Context, data *model.APIResponseWithdraw) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(withdrawByIdCacheKey, data.Data.WithdrawID)
	cache.SetToCache(ctx, w.store, key, data, ttlDefault)
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
