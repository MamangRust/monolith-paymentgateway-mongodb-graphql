package authgraphqlmapper

import (
	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	pb "github.com/MamangRust/monolith-payment-gateway-pb"
	pbUser "github.com/MamangRust/monolith-payment-gateway-pb/user"
)

type authGraphqlMapper struct {
}

func NewAuthGraphqlMapper() *authGraphqlMapper {
	return &authGraphqlMapper{}
}

func (s *authGraphqlMapper) ToGraphqlVerifyCode(res *pb.ApiResponseVerifyCode) *model.APIResponseVerifyCode {
	return &model.APIResponseVerifyCode{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (s *authGraphqlMapper) ToGraphqlForgotPassword(res *pb.ApiResponseForgotPassword) *model.APIResponseForgotPassword {
	return &model.APIResponseForgotPassword{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (s *authGraphqlMapper) ToGraphqlResetPassword(res *pb.ApiResponseResetPassword) *model.APIResponseResetPassword {
	return &model.APIResponseResetPassword{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (s *authGraphqlMapper) ToGraphqlResponseLogin(res *pb.ApiResponseLogin) *model.APIResponseLogin {
	return &model.APIResponseLogin{
		Status:  res.Status,
		Message: res.Message,
		Data:    s.mapResponseToken(res.Data),
	}
}

func (s *authGraphqlMapper) ToGraphqlResponseRegister(res *pb.ApiResponseRegister) *model.APIResponseRegister {
	return &model.APIResponseRegister{
		Status:  res.Status,
		Message: res.Message,
		Data:    s.mapResponseUser(res.Data),
	}
}

func (s *authGraphqlMapper) ToGraphqlResponseRefreshToken(res *pb.ApiResponseRefreshToken) *model.APIResponseRefreshToken {
	return &model.APIResponseRefreshToken{
		Status:  res.Status,
		Message: res.Message,
		Data:    s.mapResponseToken(res.Data),
	}
}

func (s *authGraphqlMapper) ToGraphqlResponseGetMe(res *pb.ApiResponseGetMe) *model.APIResponseGetMe {
	return &model.APIResponseGetMe{
		Status:  res.Status,
		Message: res.Message,
		Data:    s.mapResponseUser(res.Data),
	}
}

func (s *authGraphqlMapper) mapResponseUser(res *pbUser.UserResponse) *model.UserResponse {
	return &model.UserResponse{
		ID:        res.Id,
		Firstname: res.Firstname,
		Lastname:  res.Lastname,
		Email:     res.Email,
		CreatedAt: res.CreatedAt,
		UpdatedAt: res.UpdatedAt,
	}
}

func (s *authGraphqlMapper) mapResponseToken(res *pb.TokenResponse) *model.TokenResponse {
	return &model.TokenResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	}
}
