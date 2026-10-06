package rolegraphqlmapper

import (
	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/role"
)

type RoleGraphqlMapper interface {
	ToGraphqlResponseRole(res *pb.ApiResponseRole) *model.APIResponseRole
	ToGraphqlResponseRoleDeleteAt(res *pb.ApiResponseRoleDeleteAt) *model.APIResponseRoleDeleteAt
	ToGraphqlResponsesRole(res *pb.ApiResponsesRole) *model.APIResponsesRole
	ToGraphqlResponseDelete(res *pb.ApiResponseRoleDelete) *model.APIResponseRoleDelete
	ToGraphqlResponseAll(res *pb.ApiResponseRoleAll) *model.APIResponseRoleAll
	ToGraphqlResponsePaginationRole(res *pb.ApiResponsePaginationRole) *model.APIResponsePaginationRole
	ToGraphqlResponsePaginationRoleDeleteAt(res *pb.ApiResponsePaginationRoleDeleteAt) *model.APIResponsePaginationRoleDeleteAt
}
