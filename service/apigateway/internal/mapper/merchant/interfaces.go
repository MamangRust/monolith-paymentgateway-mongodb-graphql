package merchantgraphqlmapper

import (
	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/merchant"
	pbStats "github.com/MamangRust/monolith-payment-gateway-pb/merchant/stats"
)

type MerchantGraphqlMapper interface {
	ToGraphqlResponseMerchant(res *pb.ApiResponseMerchant) *model.APIResponseMerchant
	ToGraphqlResponsesMerchant(res *pb.ApiResponsesMerchant) *model.APIResponsesMerchant
	ToGraphqlResponsePaginationMerchant(res *pb.ApiResponsePaginationMerchant) *model.APIResponsePaginationMerchant
	ToGraphqlResponseMerchantDeleteAt(res *pb.ApiResponseMerchantDeleteAt) *model.APIResponseMerchantDeleteAt
	ToGraphqlResponsePaginationMerchantDeleteAt(res *pb.ApiResponsePaginationMerchantDeleteAt) *model.APIResponsePaginationMerchantDeleteAt
	ToGraphqlResponsePaginationTransaction(res *pb.ApiResponsePaginationMerchantTransaction) *model.APIResponsePaginationMerchantTransaction
	ToGraphqlMonthlyPaymentMethods(res *pbStats.ApiResponseMerchantMonthlyPaymentMethod) *model.APIResponseMerchantMonthlyPaymentMethod
	ToGraphqlYearlyPaymentMethods(res *pbStats.ApiResponseMerchantYearlyPaymentMethod) *model.APIResponseMerchantYearlyPaymentMethod
	ToGraphqlMonthlyAmounts(res *pbStats.ApiResponseMerchantMonthlyAmount) *model.APIResponseMerchantMonthlyAmount
	ToGraphqlYearlyAmounts(res *pbStats.ApiResponseMerchantYearlyAmount) *model.APIResponseMerchantYearlyAmount
	ToGraphqlMonthlyTotalAmounts(res *pbStats.ApiResponseMerchantMonthlyTotalAmount) *model.APIResponseMerchantMonthlyTotalAmount
	ToGraphqlYearlyTotalAmounts(res *pbStats.ApiResponseMerchantYearlyTotalAmount) *model.APIResponseMerchantYearlyTotalAmount
	ToGraphqlMerchantDeleteAll(res *pb.ApiResponseMerchantDelete) *model.APIResponseMerchantDelete
	ToGraphqlMerchantAll(res *pb.ApiResponseMerchantAll) *model.APIResponseMerchantAll
}
