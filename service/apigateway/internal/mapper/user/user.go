package usergraphqlmapper

import (
	graphqlmapper "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/mapper"
	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/user"
)

type userGraphqlMapper struct {
}

func NewUserGraphqlMapper() *userGraphqlMapper {
	return &userGraphqlMapper{}
}

func (u *userGraphqlMapper) ToGraphqlResponseUserDelete(res *pb.ApiResponseUserDelete) *model.APIResponseUserDelete {
	return &model.APIResponseUserDelete{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (u *userGraphqlMapper) ToGraphqlResponseUserAll(res *pb.ApiResponseUserAll) *model.APIResponseUserAll {
	return &model.APIResponseUserAll{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (u *userGraphqlMapper) ToGraphqlResponseUser(res *pb.ApiResponseUser) *model.APIResponseUser {
	return &model.APIResponseUser{
		Status:  res.Status,
		Message: res.Message,
		Data:    u.mapUserResponse(res.Data),
	}
}

func (u *userGraphqlMapper) ToGraphqlResponseUserDeleteAt(res *pb.ApiResponseUserDeleteAt) *model.APIResponseUserDeleteAt {
	return &model.APIResponseUserDeleteAt{
		Status:  res.Status,
		Message: res.Message,
		Data:    u.mapUserResponseDeleteAt(res.Data),
	}
}

func (u *userGraphqlMapper) ToGraphqlResponseUsers(res *pb.ApiResponsesUser) *model.APIResponsesUser {
	return &model.APIResponsesUser{
		Status:  res.Status,
		Message: res.Message,
		Data:    u.mapUserResponses(res.Data),
	}
}

func (u *userGraphqlMapper) ToGraphqlResponsePaginationUser(res *pb.ApiResponsePaginationUser) *model.APIResponsePaginationUser {
	return &model.APIResponsePaginationUser{
		Status:     res.Status,
		Message:    res.Message,
		Data:       u.mapUserResponses(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (u *userGraphqlMapper) ToGraphqlResponsePaginationUserDeleteAt(res *pb.ApiResponsePaginationUserDeleteAt) *model.APIResponsePaginationUserDeleteAt {
	return &model.APIResponsePaginationUserDeleteAt{
		Status:     res.Status,
		Message:    res.Message,
		Data:       u.mapUserResponsesDeleteAt(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (u *userGraphqlMapper) mapUserResponse(user *pb.UserResponse) *model.UserResponse {
	return &model.UserResponse{
		ID:        int32(user.Id),
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func (u *userGraphqlMapper) mapUserResponses(users []*pb.UserResponse) []*model.UserResponse {
	var responses []*model.UserResponse

	for _, user := range users {
		responses = append(responses, u.mapUserResponse(user))
	}

	return responses
}

func (u *userGraphqlMapper) mapUserResponseDeleteAt(user *pb.UserResponseDeleteAt) *model.UserResponseDeletedAt {
	var deletedAt string

	if user.DeletedAt != nil {
		deletedAt = user.DeletedAt.Value
	}

	return &model.UserResponseDeletedAt{
		ID:        int32(user.Id),
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		DeletedAt: &deletedAt,
	}
}

func (u *userGraphqlMapper) mapUserResponsesDeleteAt(users []*pb.UserResponseDeleteAt) []*model.UserResponseDeletedAt {
	var responses []*model.UserResponseDeletedAt

	for _, user := range users {
		responses = append(responses, u.mapUserResponseDeleteAt(user))
	}

	return responses
}
