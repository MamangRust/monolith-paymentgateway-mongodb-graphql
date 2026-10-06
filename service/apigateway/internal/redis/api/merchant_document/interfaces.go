package merchant_document_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
)

type MerchantDocumentQueryCache interface {
	GetCachedMerchantDocuments(ctx context.Context, req model.FindAllMerchantDocumentsInput) (*model.APIResponsePaginationMerchantDocument, bool)
	SetCachedMerchantDocuments(ctx context.Context, req model.FindAllMerchantDocumentsInput, data *model.APIResponsePaginationMerchantDocument)

	GetCachedMerchantDocumentActive(ctx context.Context, req model.FindAllMerchantDocumentsInput) (*model.APIResponsePaginationMerchantDocumentAt, bool)
	SetCachedMerchantDocumentActive(ctx context.Context, req model.FindAllMerchantDocumentsInput, data *model.APIResponsePaginationMerchantDocumentAt)

	GetCachedMerchantDocumentTrashed(ctx context.Context, req model.FindAllMerchantDocumentsInput) (*model.APIResponsePaginationMerchantDocumentAt, bool)
	SetCachedMerchantDocumentTrashed(ctx context.Context, req model.FindAllMerchantDocumentsInput, data *model.APIResponsePaginationMerchantDocumentAt)

	GetCachedMerchantDocumentById(ctx context.Context, id int) (*model.APIResponseMerchantDocument, bool)
	SetCachedMerchantDocumentById(ctx context.Context, id int, data *model.APIResponseMerchantDocument)
}

type MerchantDocumentCommandCache interface {
	DeleteCachedMerchantDocument(ctx context.Context, id int)
}
