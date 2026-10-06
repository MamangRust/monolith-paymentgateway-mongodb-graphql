package authgraphqlmapper

import (
	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	pb "github.com/MamangRust/monolith-payment-gateway-pb"
)

type AuthGraphqlMapper interface {
	ToGraphqlVerifyCode(res *pb.ApiResponseVerifyCode) *model.APIResponseVerifyCode
	ToGraphqlForgotPassword(res *pb.ApiResponseForgotPassword) *model.APIResponseForgotPassword
	ToGraphqlResetPassword(res *pb.ApiResponseResetPassword) *model.APIResponseResetPassword
	ToGraphqlResponseLogin(res *pb.ApiResponseLogin) *model.APIResponseLogin
	ToGraphqlResponseRegister(res *pb.ApiResponseRegister) *model.APIResponseRegister
	ToGraphqlResponseRefreshToken(res *pb.ApiResponseRefreshToken) *model.APIResponseRefreshToken
	ToGraphqlResponseGetMe(res *pb.ApiResponseGetMe) *model.APIResponseGetMe
}
