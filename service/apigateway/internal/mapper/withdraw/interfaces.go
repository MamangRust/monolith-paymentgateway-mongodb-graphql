package withdrawgraphqlmapper

import (
	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/withdraw"
	pbStats "github.com/MamangRust/monolith-payment-gateway-pb/withdraw/stats"
)

type WithdrawGraphqlMapper interface {
	ToGraphqlWithdrawAll(res *pb.ApiResponseWithdrawAll) *model.APIResponseWithdrawAll
	ToGraphqlWithdrawDelete(res *pb.ApiResponseWithdrawDelete) *model.APIResponseWithdrawDelete
	ToGraphqlResponseWithdraw(res *pb.ApiResponseWithdraw) *model.APIResponseWithdraw
	ToGraphqlResponseWithdraws(res *pb.ApiResponsesWithdraw) *model.APIResponsesWithdraw
	ToGraphqlResponseWithdrawDeleteAt(res *pb.ApIResponseWithdrawDeleteAt) *model.APIResponseWithdrawDeleteAt
	ToGraphqlResponsePaginationWithdraw(res *pb.ApiResponsePaginationWithdraw) *model.APIResponsePaginationWithdraw
	ToGraphqlResponsePaginationWithdrawDeleteAt(res *pb.ApiResponsePaginationWithdrawDeleteAt) *model.APIResponsePaginationWithdrawDeleteAt
	ToGraphqlResponseWithdrawMonthAmount(res *pbStats.ApiResponseWithdrawMonthAmount) *model.APIResponseWithdrawMonthAmount
	ToGraphqlResponseWithdrawYearAmount(res *pbStats.ApiResponseWithdrawYearAmount) *model.APIResponseWithdrawYearAmount
	ToGraphqlResponseWithdrawMonthStatusSuccess(res *pbStats.ApiResponseWithdrawMonthStatusSuccess) *model.APIResponseWithdrawMonthStatusSuccess
	ToGraphqlResponseWithdrawYearStatusSuccess(res *pbStats.ApiResponseWithdrawYearStatusSuccess) *model.APIResponseWithdrawYearStatusSuccess
	ToGraphqlResponseWithdrawMonthStatusFailed(res *pbStats.ApiResponseWithdrawMonthStatusFailed) *model.APIResponseWithdrawMonthStatusFailed
	ToGraphqlResponseWithdrawYearStatusFailed(res *pbStats.ApiResponseWithdrawYearStatusFailed) *model.APIResponseWithdrawYearStatusFailed
}
