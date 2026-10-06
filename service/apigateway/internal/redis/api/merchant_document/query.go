package merchant_document_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
)

type merchantDocumentQueryCache struct {
	store *cache.CacheStore
}

func NewMerchantDocumentQueryCache(store *cache.CacheStore) MerchantDocumentQueryCache {
	return &merchantDocumentQueryCache{store: store}
}

func (r *merchantDocumentQueryCache) SetCachedMerchantDocuments(ctx context.Context, req model.FindAllMerchantDocumentsInput, data *model.APIResponsePaginationMerchantDocument) {
	if data == nil {
		return
	}
	// Check pointers before dereferencing
	var page, pageSize int32
	var search string
	if req.Page != nil {
		page = *req.Page
	}
	if req.PageSize != nil {
		pageSize = *req.PageSize
	}
	if req.Search != nil {
		search = *req.Search
	}
	key := fmt.Sprintf(merchantDocumentAllCacheKey, page, pageSize, search)
	cache.SetToCache(ctx, r.store, key, data, ttlDefault)
}

func (r *merchantDocumentQueryCache) GetCachedMerchantDocuments(ctx context.Context, req model.FindAllMerchantDocumentsInput) (*model.APIResponsePaginationMerchantDocument, bool) {
	var page, pageSize int32
	var search string
	if req.Page != nil {
		page = *req.Page
	}
	if req.PageSize != nil {
		pageSize = *req.PageSize
	}
	if req.Search != nil {
		search = *req.Search
	}
	key := fmt.Sprintf(merchantDocumentAllCacheKey, page, pageSize, search)
	result, found := cache.GetFromCache[model.APIResponsePaginationMerchantDocument](ctx, r.store, key)
	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (r *merchantDocumentQueryCache) SetCachedMerchantDocumentActive(ctx context.Context, req model.FindAllMerchantDocumentsInput, data *model.APIResponsePaginationMerchantDocumentAt) {
	if data == nil {
		return
	}
	var page, pageSize int32
	var search string
	if req.Page != nil {
		page = *req.Page
	}
	if req.PageSize != nil {
		pageSize = *req.PageSize
	}
	if req.Search != nil {
		search = *req.Search
	}
	key := fmt.Sprintf(merchantDocumentActiveCacheKey, page, pageSize, search)
	cache.SetToCache(ctx, r.store, key, data, ttlDefault)
}

func (r *merchantDocumentQueryCache) GetCachedMerchantDocumentActive(ctx context.Context, req model.FindAllMerchantDocumentsInput) (*model.APIResponsePaginationMerchantDocumentAt, bool) {
	var page, pageSize int32
	var search string
	if req.Page != nil {
		page = *req.Page
	}
	if req.PageSize != nil {
		pageSize = *req.PageSize
	}
	if req.Search != nil {
		search = *req.Search
	}
	key := fmt.Sprintf(merchantDocumentActiveCacheKey, page, pageSize, search)
	result, found := cache.GetFromCache[model.APIResponsePaginationMerchantDocumentAt](ctx, r.store, key)
	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (r *merchantDocumentQueryCache) SetCachedMerchantDocumentTrashed(ctx context.Context, req model.FindAllMerchantDocumentsInput, data *model.APIResponsePaginationMerchantDocumentAt) {
	if data == nil {
		return
	}
	var page, pageSize int32
	var search string
	if req.Page != nil {
		page = *req.Page
	}
	if req.PageSize != nil {
		pageSize = *req.PageSize
	}
	if req.Search != nil {
		search = *req.Search
	}
	key := fmt.Sprintf(merchantDocumentTrashedCacheKey, page, pageSize, search)
	cache.SetToCache(ctx, r.store, key, data, ttlDefault)
}

func (r *merchantDocumentQueryCache) GetCachedMerchantDocumentTrashed(ctx context.Context, req model.FindAllMerchantDocumentsInput) (*model.APIResponsePaginationMerchantDocumentAt, bool) {
	var page, pageSize int32
	var search string
	if req.Page != nil {
		page = *req.Page
	}
	if req.PageSize != nil {
		pageSize = *req.PageSize
	}
	if req.Search != nil {
		search = *req.Search
	}
	key := fmt.Sprintf(merchantDocumentTrashedCacheKey, page, pageSize, search)
	result, found := cache.GetFromCache[model.APIResponsePaginationMerchantDocumentAt](ctx, r.store, key)
	if !found || result == nil {
		return nil, false
	}
	return result, true
}

func (r *merchantDocumentQueryCache) SetCachedMerchantDocumentById(ctx context.Context, id int, data *model.APIResponseMerchantDocument) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantDocumentByIdCacheKey, id)
	cache.SetToCache(ctx, r.store, key, data, ttlDefault)
}

func (r *merchantDocumentQueryCache) GetCachedMerchantDocumentById(ctx context.Context, id int) (*model.APIResponseMerchantDocument, bool) {
	key := fmt.Sprintf(merchantDocumentByIdCacheKey, id)
	result, found := cache.GetFromCache[model.APIResponseMerchantDocument](ctx, r.store, key)
	if !found || result == nil {
		return nil, false
	}
	return result, true
}
