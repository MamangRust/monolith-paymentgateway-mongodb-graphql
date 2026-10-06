package cardgraphqlmapper

import (
	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/card"
	pbStats "github.com/MamangRust/monolith-payment-gateway-pb/card/stats"
)

type CardGraphqlMapper interface {
	ToGraphqlResponsePaginationCard(res *pb.ApiResponsePaginationCard) *model.APIResponsePaginationCard
	ToGraphqlResponseCard(res *pb.ApiResponseCard) *model.APIResponseCard
	ToGraphqlResponseAll(res *pb.ApiResponseCardAll) *model.APIResponseCardAll
	ToGraphqlResponseDelete(res *pb.ApiResponseCardDelete) *model.APIResponseCardDelete
	ToGraphqlResponseCardDeleteAt(res *pb.ApiResponseCardDeleteAt) *model.APIResponseCardDeleteAt
	ToGraphqlResponsePaginationCardDeleteAt(res *pb.ApiResponsePaginationCardDeleteAt) *model.APIResponsePaginationCardDeleteAt
	ToGraphqlDashboardCard(res *pb.ApiResponseDashboardCard) *model.APIResponseDashboardCard
	ToGraphqlDashboardCardCardNumber(res *pb.ApiResponseDashboardCardNumber) *model.APIResponseDashboardCardNumber
	ToGraphqlMonthlyBalances(res *pbStats.ApiResponseMonthlyBalance) *model.APIResponseMonthlyBalance
	ToGraphqlYearlyBalances(res *pbStats.ApiResponseYearlyBalance) *model.APIResponseYearlyBalance
	ToGraphqlMonthlyAmounts(res *pb.ApiResponseMonthlyAmount) *model.APIResponseMonthlyAmount
	ToGraphqlYearlyAmounts(res *pb.ApiResponseYearlyAmount) *model.APIResponseYearlyAmount
}
