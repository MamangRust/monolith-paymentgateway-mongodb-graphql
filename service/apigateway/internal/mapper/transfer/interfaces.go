package transfergraphqlmapper

import (
	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/transfer"
	pbStats "github.com/MamangRust/monolith-payment-gateway-pb/transfer/stats"
)

type TransferGraphqlMapper interface {
	ToGraphqlTransferAll(res *pb.ApiResponseTransferAll) *model.APIResponseTransferAll
	ToGraphqlTransferDelete(res *pb.ApiResponseTransferDelete) *model.APIResponseTransferDelete
	ToGraphqlResponseTransfer(res *pb.ApiResponseTransfer) *model.APIResponseTransfer
	ToGraphqlResponseTransfers(res *pb.ApiResponseTransfers) *model.APIResponseTransfers
	ToGraphqlResponseTransferDeleteAt(res *pb.ApIResponseTransferDeleteAt) *model.APIResponseTransferDeleteAt
	ToGraphqlResponsePaginationTransfer(res *pb.ApiResponsePaginationTransfer) *model.APIResponsePaginationTransfer
	ToGraphqlResponsePaginationTransferDeleteAt(res *pb.ApiResponsePaginationTransferDeleteAt) *model.APIResponsePaginationTransferDeleteAt
	ToGraphqlResponseTransferMonthAmount(res *pbStats.ApiResponseTransferMonthAmount) *model.APIResponseTransferMonthAmount
	ToGraphqlResponseTransferYearAmount(res *pbStats.ApiResponseTransferYearAmount) *model.APIResponseTransferYearAmount
	ToGraphqlResponseTransferMonthStatusSuccess(res *pbStats.ApiResponseTransferMonthStatusSuccess) *model.APIResponseTransferMonthStatusSuccess
	ToGraphqlResponseTransferYearStatusSuccess(res *pbStats.ApiResponseTransferYearStatusSuccess) *model.APIResponseTransferYearStatusSuccess
	ToGraphqlResponseTransferMonthStatusFailed(res *pbStats.ApiResponseTransferMonthStatusFailed) *model.APIResponseTransferMonthStatusFailed
	ToGraphqlResponseTransferYearStatusFailed(res *pbStats.ApiResponseTransferYearStatusFailed) *model.APIResponseTransferYearStatusFailed
}
