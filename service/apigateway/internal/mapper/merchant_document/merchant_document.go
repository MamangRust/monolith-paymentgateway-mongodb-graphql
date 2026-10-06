package merchantdocumentgraphqlmapper

import (
	graphqlmapper "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/mapper"
	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/merchant_document"
)

type merchantDocumentGraphqlMapper struct{}

func NewMerchantDocumentGraphqlMapper() *merchantDocumentGraphqlMapper {
	return &merchantDocumentGraphqlMapper{}
}

func (s *merchantDocumentGraphqlMapper) ToGraphqlResponseMerchantDocument(res *pb.ApiResponseMerchantDocument) *model.APIResponseMerchantDocument {
	return &model.APIResponseMerchantDocument{
		Status:  res.Status,
		Message: res.Message,
		Data:    s.mapResponseMerchantDocument(res.Data),
	}
}

func (s *merchantDocumentGraphqlMapper) ToGraphqlResponseMerchantDocumentDeleteAt(res *pb.ApiResponseMerchantDocumentDeleteAt) *model.APIResponseMerchantDocumentDeleteAt {
	return &model.APIResponseMerchantDocumentDeleteAt{
		Status:  res.Status,
		Message: res.Message,
		Data:    s.mapResponseMerchantDocumentDeleteAt(res.Data),
	}
}

func (s *merchantDocumentGraphqlMapper) ToGraphqlResponseDelete(res *pb.ApiResponseMerchantDocumentDelete) *model.APIResponseMerchantDocumentDelete {
	return &model.APIResponseMerchantDocumentDelete{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (s *merchantDocumentGraphqlMapper) ToGraphqlResponseAll(res *pb.ApiResponseMerchantDocumentAll) *model.APIResponseMerchantDocumentAll {
	return &model.APIResponseMerchantDocumentAll{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (s *merchantDocumentGraphqlMapper) ToGraphqlResponsePaginationMerchantDocument(res *pb.ApiResponsePaginationMerchantDocument) *model.APIResponsePaginationMerchantDocument {
	return &model.APIResponsePaginationMerchantDocument{
		Status:     res.Status,
		Message:    res.Message,
		Data:       s.mapResponsesMerchantDocument(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (s *merchantDocumentGraphqlMapper) ToGraphqlResponsePaginationMerchantDocumentDeleteAt(res *pb.ApiResponsePaginationMerchantDocumentAt) *model.APIResponsePaginationMerchantDocumentAt {
	return &model.APIResponsePaginationMerchantDocumentAt{
		Status:     res.Status,
		Message:    res.Message,
		Data:       s.mapResponsesMerchantDocumentDeleteAt(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (s *merchantDocumentGraphqlMapper) mapResponseMerchantDocument(doc *pb.MerchantDocument) *model.MerchantDocumentResponse {
	if doc == nil {
		return nil
	}

	return &model.MerchantDocumentResponse{
		DocumentID:   int32(doc.DocumentId),
		MerchantID:   int32(doc.MerchantId),
		DocumentType: doc.DocumentType,
		DocumentURL:  doc.DocumentUrl,
		Status:       doc.Status,
		Note:         doc.Note,
		UploadedAt:   doc.UploadedAt,
		UpdatedAt:    doc.UpdatedAt,
	}
}

func (s *merchantDocumentGraphqlMapper) mapResponsesMerchantDocument(docs []*pb.MerchantDocument) []*model.MerchantDocumentResponse {
	var responseDocs []*model.MerchantDocumentResponse
	for _, doc := range docs {
		responseDocs = append(responseDocs, s.mapResponseMerchantDocument(doc))
	}
	return responseDocs
}

func (s *merchantDocumentGraphqlMapper) mapResponseMerchantDocumentDeleteAt(doc *pb.MerchantDocumentDeleteAt) *model.MerchantDocumentResponseDeleteAt {
	if doc == nil {
		return nil
	}

	var deletedAt string
	if doc.DeletedAt != nil {
		deletedAt = doc.DeletedAt.Value
	}

	return &model.MerchantDocumentResponseDeleteAt{
		DocumentID:   int32(doc.DocumentId),
		MerchantID:   int32(doc.MerchantId),
		DocumentType: doc.DocumentType,
		DocumentURL:  doc.DocumentUrl,
		Status:       doc.Status,
		Note:         doc.Note,
		UploadedAt:   doc.UploadedAt,
		UpdatedAt:    doc.UpdatedAt,
		DeletedAt:    &deletedAt,
	}
}

func (s *merchantDocumentGraphqlMapper) mapResponsesMerchantDocumentDeleteAt(docs []*pb.MerchantDocumentDeleteAt) []*model.MerchantDocumentResponseDeleteAt {
	var responseDocs []*model.MerchantDocumentResponseDeleteAt
	for _, doc := range docs {
		responseDocs = append(responseDocs, s.mapResponseMerchantDocumentDeleteAt(doc))
	}
	return responseDocs
}

func (s *merchantDocumentGraphqlMapper) mapResponsesMerchantDocumentFromDeleteAt(docs []*pb.MerchantDocumentDeleteAt) []*model.MerchantDocumentResponse {
	var responseDocs []*model.MerchantDocumentResponse
	for _, doc := range docs {
		if doc == nil {
			continue
		}
		responseDocs = append(responseDocs, &model.MerchantDocumentResponse{
			DocumentID:   int32(doc.DocumentId),
			MerchantID:   int32(doc.MerchantId),
			DocumentType: doc.DocumentType,
			DocumentURL:  doc.DocumentUrl,
			Status:       doc.Status,
			Note:         doc.Note,
			UploadedAt:   doc.UploadedAt,
			UpdatedAt:    doc.UpdatedAt,
		})
	}
	return responseDocs
}
