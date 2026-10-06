package rolegraphqlmapper

import (
	graphqlmapper "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/mapper"
	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/role"
)

type roleGraphqlMapper struct {
}

func NewRoleGraphqlMapper() *roleGraphqlMapper {
	return &roleGraphqlMapper{}
}

func (s *roleGraphqlMapper) ToGraphqlResponseAll(res *pb.ApiResponseRoleAll) *model.APIResponseRoleAll {
	return &model.APIResponseRoleAll{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (s *roleGraphqlMapper) ToGraphqlResponseDelete(res *pb.ApiResponseRoleDelete) *model.APIResponseRoleDelete {
	return &model.APIResponseRoleDelete{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (s *roleGraphqlMapper) ToGraphqlResponseRole(res *pb.ApiResponseRole) *model.APIResponseRole {
	return &model.APIResponseRole{
		Status:  res.Status,
		Message: res.Message,
		Data:    s.mapResponseRole(res.Data),
	}
}

func (s *roleGraphqlMapper) ToGraphqlResponseRoleDeleteAt(res *pb.ApiResponseRoleDeleteAt) *model.APIResponseRoleDeleteAt {
	return &model.APIResponseRoleDeleteAt{
		Status:  res.Status,
		Message: res.Message,
		Data:    s.mapResponseRoleDeleteAt(res.Data),
	}
}

func (s *roleGraphqlMapper) ToGraphqlResponsesRole(res *pb.ApiResponsesRole) *model.APIResponsesRole {
	return &model.APIResponsesRole{
		Status:  res.Status,
		Message: res.Message,
		Data:    s.mapResponsesRole(res.Data),
	}
}

func (s *roleGraphqlMapper) ToGraphqlResponsePaginationRole(res *pb.ApiResponsePaginationRole) *model.APIResponsePaginationRole {
	return &model.APIResponsePaginationRole{
		Status:     res.Status,
		Message:    res.Message,
		Data:       s.mapResponsesRole(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (s *roleGraphqlMapper) ToGraphqlResponsePaginationRoleDeleteAt(res *pb.ApiResponsePaginationRoleDeleteAt) *model.APIResponsePaginationRoleDeleteAt {
	return &model.APIResponsePaginationRoleDeleteAt{
		Status:     res.Status,
		Message:    res.Message,
		Data:       s.mapResponsesRoleDeleteAt(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (s *roleGraphqlMapper) mapResponseRole(role *pb.RoleResponse) *model.RoleResponse {
	return &model.RoleResponse{
		ID:        int32(role.Id),
		Name:      role.Name,
		CreatedAt: role.CreatedAt,
		UpdatedAt: role.UpdatedAt,
	}
}

func (s *roleGraphqlMapper) mapResponsesRole(roles []*pb.RoleResponse) []*model.RoleResponse {
	var responseRoles []*model.RoleResponse

	for _, role := range roles {
		responseRoles = append(responseRoles, s.mapResponseRole(role))
	}

	return responseRoles
}

func (s *roleGraphqlMapper) mapResponseRoleDeleteAt(role *pb.RoleResponseDeleteAt) *model.RoleResponseDeletedAt {
	var deletedAt string

	if role.DeletedAt != nil {
		deletedAt = role.DeletedAt.Value
	}

	return &model.RoleResponseDeletedAt{
		ID:        int32(role.Id),
		Name:      role.Name,
		CreatedAt: role.CreatedAt,
		UpdatedAt: role.UpdatedAt,
		DeletedAt: &deletedAt,
	}
}

func (s *roleGraphqlMapper) mapResponsesRoleDeleteAt(roles []*pb.RoleResponseDeleteAt) []*model.RoleResponseDeletedAt {
	var responseRoles []*model.RoleResponseDeletedAt

	for _, role := range roles {
		responseRoles = append(responseRoles, s.mapResponseRoleDeleteAt(role))
	}

	return responseRoles
}
