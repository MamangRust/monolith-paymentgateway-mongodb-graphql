package transfergraphqlmapper

import (
	graphqlmapper "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/mapper"
	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/transfer"
	pbStats "github.com/MamangRust/monolith-payment-gateway-pb/transfer/stats"
)

type transferGraphqlMapper struct {
}

func NewTransferGraphqlMapper() *transferGraphqlMapper {
	return &transferGraphqlMapper{}
}

func (t *transferGraphqlMapper) ToGraphqlTransferAll(res *pb.ApiResponseTransferAll) *model.APIResponseTransferAll {
	return &model.APIResponseTransferAll{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (t *transferGraphqlMapper) ToGraphqlTransferDelete(res *pb.ApiResponseTransferDelete) *model.APIResponseTransferDelete {
	return &model.APIResponseTransferDelete{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (t *transferGraphqlMapper) ToGraphqlResponseTransfer(res *pb.ApiResponseTransfer) *model.APIResponseTransfer {
	return &model.APIResponseTransfer{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponseTransfer(res.Data),
	}
}

func (t *transferGraphqlMapper) ToGraphqlResponseTransfers(res *pb.ApiResponseTransfers) *model.APIResponseTransfers {
	return &model.APIResponseTransfers{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesTransfer(res.Data),
	}
}

func (t *transferGraphqlMapper) ToGraphqlResponseTransferDeleteAt(res *pb.ApIResponseTransferDeleteAt) *model.APIResponseTransferDeleteAt {
	return &model.APIResponseTransferDeleteAt{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponseTransferDeleteAt(res.Data),
	}
}

func (t *transferGraphqlMapper) ToGraphqlResponsePaginationTransfer(res *pb.ApiResponsePaginationTransfer) *model.APIResponsePaginationTransfer {
	return &model.APIResponsePaginationTransfer{
		Status:     res.Status,
		Message:    res.Message,
		Data:       t.mapResponsesTransfer(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (t *transferGraphqlMapper) ToGraphqlResponsePaginationTransferDeleteAt(res *pb.ApiResponsePaginationTransferDeleteAt) *model.APIResponsePaginationTransferDeleteAt {
	return &model.APIResponsePaginationTransferDeleteAt{
		Status:     res.Status,
		Message:    res.Message,
		Data:       t.mapResponsesTransferDeleteAt(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (t *transferGraphqlMapper) ToGraphqlResponseTransferMonthAmount(res *pbStats.ApiResponseTransferMonthAmount) *model.APIResponseTransferMonthAmount {
	return &model.APIResponseTransferMonthAmount{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesTransferMonthAmount(res.Data),
	}
}

func (t *transferGraphqlMapper) ToGraphqlResponseTransferYearAmount(res *pbStats.ApiResponseTransferYearAmount) *model.APIResponseTransferYearAmount {
	return &model.APIResponseTransferYearAmount{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesTransferYearAmount(res.Data),
	}
}

func (t *transferGraphqlMapper) ToGraphqlResponseTransferMonthStatusSuccess(res *pbStats.ApiResponseTransferMonthStatusSuccess) *model.APIResponseTransferMonthStatusSuccess {
	return &model.APIResponseTransferMonthStatusSuccess{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesMonthStatusSuccess(res.Data),
	}
}

func (t *transferGraphqlMapper) ToGraphqlResponseTransferYearStatusSuccess(res *pbStats.ApiResponseTransferYearStatusSuccess) *model.APIResponseTransferYearStatusSuccess {
	return &model.APIResponseTransferYearStatusSuccess{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesYearStatusSuccess(res.Data),
	}
}

func (t *transferGraphqlMapper) ToGraphqlResponseTransferMonthStatusFailed(res *pbStats.ApiResponseTransferMonthStatusFailed) *model.APIResponseTransferMonthStatusFailed {
	return &model.APIResponseTransferMonthStatusFailed{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesMonthStatusFailed(res.Data),
	}
}

func (t *transferGraphqlMapper) ToGraphqlResponseTransferYearStatusFailed(res *pbStats.ApiResponseTransferYearStatusFailed) *model.APIResponseTransferYearStatusFailed {
	return &model.APIResponseTransferYearStatusFailed{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesYearStatusFailed(res.Data),
	}
}

func (t *transferGraphqlMapper) mapResponseTransfer(transfer *pb.TransferResponse) *model.TransferResponse {
	return &model.TransferResponse{
		ID:             int32(transfer.Id),
		TransferNo:     transfer.TransferNo,
		TransferFrom:   transfer.TransferFrom,
		TransferTo:     transfer.TransferTo,
		TransferAmount: int32(transfer.TransferAmount),
		TransferTime:   transfer.TransferTime,
		CreatedAt:      transfer.CreatedAt,
		UpdatedAt:      transfer.UpdatedAt,
	}
}

func (t *transferGraphqlMapper) mapResponsesTransfer(transfers []*pb.TransferResponse) []*model.TransferResponse {
	var responses []*model.TransferResponse

	for _, transfer := range transfers {
		responses = append(responses, t.mapResponseTransfer(transfer))
	}

	return responses
}

func (t *transferGraphqlMapper) mapResponseTransferDeleteAt(transfer *pb.TransferResponseDeleteAt) *model.TransferResponseDeletedAt {
	var deletedAt string

	if transfer.DeletedAt != nil {
		deletedAt = transfer.DeletedAt.Value
	}

	return &model.TransferResponseDeletedAt{
		ID:             int32(transfer.Id),
		TransferNo:     transfer.TransferNo,
		TransferFrom:   transfer.TransferFrom,
		TransferTo:     transfer.TransferTo,
		TransferAmount: int32(transfer.TransferAmount),
		TransferTime:   transfer.TransferTime,
		CreatedAt:      transfer.CreatedAt,
		UpdatedAt:      transfer.UpdatedAt,
		DeletedAt:      &deletedAt,
	}
}

func (t *transferGraphqlMapper) mapResponsesTransferDeleteAt(transfers []*pb.TransferResponseDeleteAt) []*model.TransferResponseDeletedAt {
	var responses []*model.TransferResponseDeletedAt

	for _, transfer := range transfers {
		responses = append(responses, t.mapResponseTransferDeleteAt(transfer))
	}

	return responses
}

func (t *transferGraphqlMapper) mapResponseMonthStatusSuccess(data *pbStats.TransferMonthStatusSuccessResponse) *model.TransferMonthStatusSuccessResponse {
	return &model.TransferMonthStatusSuccessResponse{
		Year:         data.Year,
		Month:        data.Month,
		TotalSuccess: int32(data.TotalSuccess),
		TotalAmount:  int32(data.TotalAmount),
	}
}

func (t *transferGraphqlMapper) mapResponsesMonthStatusSuccess(transfers []*pbStats.TransferMonthStatusSuccessResponse) []*model.TransferMonthStatusSuccessResponse {
	var responses []*model.TransferMonthStatusSuccessResponse

	for _, transfer := range transfers {
		responses = append(responses, t.mapResponseMonthStatusSuccess(transfer))
	}

	return responses
}

func (t *transferGraphqlMapper) mapResponseYearStatusSuccess(data *pbStats.TransferYearStatusSuccessResponse) *model.TransferYearStatusSuccessResponse {
	return &model.TransferYearStatusSuccessResponse{
		Year:         data.Year,
		TotalSuccess: int32(data.TotalSuccess),
		TotalAmount:  int32(data.TotalAmount),
	}
}

func (t *transferGraphqlMapper) mapResponsesYearStatusSuccess(transfers []*pbStats.TransferYearStatusSuccessResponse) []*model.TransferYearStatusSuccessResponse {
	var responses []*model.TransferYearStatusSuccessResponse

	for _, transfer := range transfers {
		responses = append(responses, t.mapResponseYearStatusSuccess(transfer))
	}

	return responses
}

func (t *transferGraphqlMapper) mapResponseMonthStatusFailed(data *pbStats.TransferMonthStatusFailedResponse) *model.TransferMonthStatusFailedResponse {
	return &model.TransferMonthStatusFailedResponse{
		Year:        data.Year,
		Month:       data.Month,
		TotalFailed: int32(data.TotalFailed),
		TotalAmount: int32(data.TotalAmount),
	}
}

func (t *transferGraphqlMapper) mapResponsesMonthStatusFailed(transfers []*pbStats.TransferMonthStatusFailedResponse) []*model.TransferMonthStatusFailedResponse {
	var responses []*model.TransferMonthStatusFailedResponse

	for _, transfer := range transfers {
		responses = append(responses, t.mapResponseMonthStatusFailed(transfer))
	}

	return responses
}

func (t *transferGraphqlMapper) mapResponseYearStatusFailed(data *pbStats.TransferYearStatusFailedResponse) *model.TransferYearStatusFailedResponse {
	return &model.TransferYearStatusFailedResponse{
		Year:        data.Year,
		TotalFailed: int32(data.TotalFailed),
		TotalAmount: int32(data.TotalAmount),
	}
}

func (t *transferGraphqlMapper) mapResponsesYearStatusFailed(transfers []*pbStats.TransferYearStatusFailedResponse) []*model.TransferYearStatusFailedResponse {
	var responses []*model.TransferYearStatusFailedResponse

	for _, transfer := range transfers {
		responses = append(responses, t.mapResponseYearStatusFailed(transfer))
	}

	return responses
}

func (t *transferGraphqlMapper) mapResponseTransferMonthAmount(transfer *pbStats.TransferMonthAmountResponse) *model.TransferMonthAmountResponse {
	return &model.TransferMonthAmountResponse{
		Month:       transfer.Month,
		TotalAmount: int32(transfer.TotalAmount),
	}
}

func (t *transferGraphqlMapper) mapResponsesTransferMonthAmount(transfers []*pbStats.TransferMonthAmountResponse) []*model.TransferMonthAmountResponse {
	var responses []*model.TransferMonthAmountResponse

	for _, transfer := range transfers {
		responses = append(responses, t.mapResponseTransferMonthAmount(transfer))
	}

	return responses
}

func (t *transferGraphqlMapper) mapResponseTransferYearAmount(transfer *pbStats.TransferYearAmountResponse) *model.TransferYearAmountResponse {
	return &model.TransferYearAmountResponse{
		Year:        transfer.Year,
		TotalAmount: int32(transfer.TotalAmount),
	}
}

func (t *transferGraphqlMapper) mapResponsesTransferYearAmount(transfers []*pbStats.TransferYearAmountResponse) []*model.TransferYearAmountResponse {
	var responses []*model.TransferYearAmountResponse

	for _, transfer := range transfers {
		responses = append(responses, t.mapResponseTransferYearAmount(transfer))
	}

	return responses
}
