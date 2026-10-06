package merchant_document_cache

import "time"

const (
	merchantDocumentAllCacheKey     = "merchant_document:all:page:%d:pageSize:%d:search:%s"
	merchantDocumentByIdCacheKey    = "merchant_document:id:%d"
	merchantDocumentActiveCacheKey  = "merchant_document:active:page:%d:pageSize:%d:search:%s"
	merchantDocumentTrashedCacheKey = "merchant_document:trashed:page:%d:pageSize:%d:search:%s"

	ttlDefault = 5 * time.Minute
)
