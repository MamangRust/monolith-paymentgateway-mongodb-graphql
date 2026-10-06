package merchantdocumentgraphqlmapper

import (
	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/merchant_document"
)

type MerchantDocumentGraphqlMapper interface {
	ToGraphqlResponseMerchantDocument(res *pb.ApiResponseMerchantDocument) *model.APIResponseMerchantDocument
	ToGraphqlResponseMerchantDocumentDeleteAt(res *pb.ApiResponseMerchantDocumentDeleteAt) *model.APIResponseMerchantDocumentDeleteAt

	ToGraphqlResponseDelete(res *pb.ApiResponseMerchantDocumentDelete) *model.APIResponseMerchantDocumentDelete
	ToGraphqlResponseAll(res *pb.ApiResponseMerchantDocumentAll) *model.APIResponseMerchantDocumentAll
	ToGraphqlResponsePaginationMerchantDocument(res *pb.ApiResponsePaginationMerchantDocument) *model.APIResponsePaginationMerchantDocument
	ToGraphqlResponsePaginationMerchantDocumentDeleteAt(res *pb.ApiResponsePaginationMerchantDocumentAt) *model.APIResponsePaginationMerchantDocumentAt
}
