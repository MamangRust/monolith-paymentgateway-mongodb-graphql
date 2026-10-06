package topupgraphqlmapper

import (
	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/topup"
	pbStats "github.com/MamangRust/monolith-payment-gateway-pb/topup/stats"
)

type TopupGraphqlMapper interface {
	ToGraphqlResponseTopup(res *pb.ApiResponseTopup) *model.APIResponseTopup
	ToGraphqlResponseTopupDeleteAt(res *pb.ApiResponseTopupDeleteAt) *model.APIResponseTopupDeleteAt
	ToGraphqlTopupAll(res *pb.ApiResponseTopupAll) *model.APIResponseTopupAll
	ToGraphqlTopupDelete(res *pb.ApiResponseTopupDelete) *model.APIResponseTopupDelete
	ToGraphqlResponsePaginationTopup(res *pb.ApiResponsePaginationTopup) *model.APIResponsePaginationTopup
	ToGraphqlResponsePaginationTopupDeleteAt(res *pb.ApiResponsePaginationTopupDeleteAt) *model.APIResponsePaginationTopupDeleteAt
	ToGraphqlResponseTopupMonthStatusSuccess(res *pbStats.ApiResponseTopupMonthStatusSuccess) *model.APIResponseTopupMonthStatusSuccess
	ToGraphqlResponseTopupYearStatusSuccess(res *pbStats.ApiResponseTopupYearStatusSuccess) *model.APIResponseTopupYearStatusSuccess
	ToGraphqlResponseTopupMonthStatusFailed(res *pbStats.ApiResponseTopupMonthStatusFailed) *model.APIResponseTopupMonthStatusFailed
	ToGraphqlResponseTopupYearStatusFailed(res *pbStats.ApiResponseTopupYearStatusFailed) *model.APIResponseTopupYearStatusFailed
	ToGraphqlResponseTopupMonthMethod(res *pbStats.ApiResponseTopupMonthMethod) *model.APIResponseTopupMonthMethod
	ToGraphqlResponseTopupYearMethod(res *pbStats.ApiResponseTopupYearMethod) *model.APIResponseTopupYearMethod
	ToGraphqlResponseTopupMonthAmount(res *pbStats.ApiResponseTopupMonthAmount) *model.APIResponseTopupMonthAmount
	ToGraphqlResponseTopupYearAmount(res *pbStats.ApiResponseTopupYearAmount) *model.APIResponseTopupYearAmount
}
