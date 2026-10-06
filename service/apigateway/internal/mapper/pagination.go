package graphqlmapper

import (
	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/common"
)

func MapPaginationMeta(s *pb.PaginationMeta) *model.PaginationMeta {
	return &model.PaginationMeta{
		CurrentPage:  int32(s.CurrentPage),
		PageSize:     int32(s.PageSize),
		TotalRecords: int32(s.TotalRecords),
		TotalPages:   int32(s.TotalPages),
	}
}
