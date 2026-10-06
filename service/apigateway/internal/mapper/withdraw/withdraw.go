package withdrawgraphqlmapper

import (
	graphqlmapper "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/mapper"
	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/withdraw"
	pbStats "github.com/MamangRust/monolith-payment-gateway-pb/withdraw/stats"
)

type withdrawGraphqlMapper struct {
}

func NewWithdrawGraphqlMapper() *withdrawGraphqlMapper {
	return &withdrawGraphqlMapper{}
}

func (t *withdrawGraphqlMapper) ToGraphqlWithdrawAll(res *pb.ApiResponseWithdrawAll) *model.APIResponseWithdrawAll {
	return &model.APIResponseWithdrawAll{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (t *withdrawGraphqlMapper) ToGraphqlWithdrawDelete(res *pb.ApiResponseWithdrawDelete) *model.APIResponseWithdrawDelete {
	return &model.APIResponseWithdrawDelete{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (t *withdrawGraphqlMapper) ToGraphqlResponseWithdraw(res *pb.ApiResponseWithdraw) *model.APIResponseWithdraw {
	return &model.APIResponseWithdraw{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponseWithdraw(res.Data),
	}
}

func (t *withdrawGraphqlMapper) ToGraphqlResponseWithdraws(res *pb.ApiResponsesWithdraw) *model.APIResponsesWithdraw {
	return &model.APIResponsesWithdraw{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesWithdraw(res.Data),
	}
}

func (t *withdrawGraphqlMapper) ToGraphqlResponseWithdrawDeleteAt(res *pb.ApIResponseWithdrawDeleteAt) *model.APIResponseWithdrawDeleteAt {
	return &model.APIResponseWithdrawDeleteAt{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponseWithdrawDeleteAt(res.Data),
	}
}

func (t *withdrawGraphqlMapper) ToGraphqlResponsePaginationWithdraw(res *pb.ApiResponsePaginationWithdraw) *model.APIResponsePaginationWithdraw {
	return &model.APIResponsePaginationWithdraw{
		Status:     res.Status,
		Message:    res.Message,
		Data:       t.mapResponsesWithdraw(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (t *withdrawGraphqlMapper) ToGraphqlResponsePaginationWithdrawDeleteAt(res *pb.ApiResponsePaginationWithdrawDeleteAt) *model.APIResponsePaginationWithdrawDeleteAt {
	return &model.APIResponsePaginationWithdrawDeleteAt{
		Status:     res.Status,
		Message:    res.Message,
		Data:       t.mapResponsesWithdrawDeleteAt(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (t *withdrawGraphqlMapper) ToGraphqlResponseWithdrawMonthAmount(res *pbStats.ApiResponseWithdrawMonthAmount) *model.APIResponseWithdrawMonthAmount {
	return &model.APIResponseWithdrawMonthAmount{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesWithdrawMonthAmount(res.Data),
	}
}

func (t *withdrawGraphqlMapper) ToGraphqlResponseWithdrawYearAmount(res *pbStats.ApiResponseWithdrawYearAmount) *model.APIResponseWithdrawYearAmount {
	return &model.APIResponseWithdrawYearAmount{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesWithdrawYearAmount(res.Data),
	}
}

func (t *withdrawGraphqlMapper) ToGraphqlResponseWithdrawMonthStatusSuccess(res *pbStats.ApiResponseWithdrawMonthStatusSuccess) *model.APIResponseWithdrawMonthStatusSuccess {
	return &model.APIResponseWithdrawMonthStatusSuccess{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesMonthStatusSuccess(res.Data),
	}
}

func (t *withdrawGraphqlMapper) ToGraphqlResponseWithdrawYearStatusSuccess(res *pbStats.ApiResponseWithdrawYearStatusSuccess) *model.APIResponseWithdrawYearStatusSuccess {
	return &model.APIResponseWithdrawYearStatusSuccess{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesYearStatusSuccess(res.Data),
	}
}

func (t *withdrawGraphqlMapper) ToGraphqlResponseWithdrawMonthStatusFailed(res *pbStats.ApiResponseWithdrawMonthStatusFailed) *model.APIResponseWithdrawMonthStatusFailed {
	return &model.APIResponseWithdrawMonthStatusFailed{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesMonthStatusFailed(res.Data),
	}
}

func (t *withdrawGraphqlMapper) ToGraphqlResponseWithdrawYearStatusFailed(res *pbStats.ApiResponseWithdrawYearStatusFailed) *model.APIResponseWithdrawYearStatusFailed {
	return &model.APIResponseWithdrawYearStatusFailed{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesYearStatusFailed(res.Data),
	}
}

func (t *withdrawGraphqlMapper) mapResponseWithdraw(Withdraw *pb.WithdrawResponse) *model.WithdrawResponse {
	return &model.WithdrawResponse{
		WithdrawID:     int32(Withdraw.WithdrawId),
		WithdrawNo:     Withdraw.WithdrawNo,
		CardNumber:     Withdraw.CardNumber,
		WithdrawAmount: int32(Withdraw.WithdrawAmount),
		WithdrawTime:   Withdraw.WithdrawTime,
		CreatedAt:      Withdraw.CreatedAt,
		UpdatedAt:      Withdraw.UpdatedAt,
	}
}

func (t *withdrawGraphqlMapper) mapResponsesWithdraw(Withdraws []*pb.WithdrawResponse) []*model.WithdrawResponse {
	var responses []*model.WithdrawResponse

	for _, Withdraw := range Withdraws {
		responses = append(responses, t.mapResponseWithdraw(Withdraw))
	}

	return responses
}

func (t *withdrawGraphqlMapper) mapResponseWithdrawDeleteAt(Withdraw *pb.WithdrawResponseDeleteAt) *model.WithdrawResponseDeletedAt {
	var deletedAt string

	if Withdraw.DeletedAt != nil {
		deletedAt = Withdraw.DeletedAt.Value
	}

	return &model.WithdrawResponseDeletedAt{
		WithdrawID:     int32(Withdraw.WithdrawId),
		WithdrawNo:     Withdraw.WithdrawNo,
		CardNumber:     Withdraw.CardNumber,
		WithdrawAmount: int32(Withdraw.WithdrawAmount),
		WithdrawTime:   Withdraw.WithdrawTime,
		CreatedAt:      Withdraw.CreatedAt,
		UpdatedAt:      Withdraw.UpdatedAt,
		DeletedAt:      &deletedAt,
	}
}

func (t *withdrawGraphqlMapper) mapResponsesWithdrawDeleteAt(Withdraws []*pb.WithdrawResponseDeleteAt) []*model.WithdrawResponseDeletedAt {
	var responses []*model.WithdrawResponseDeletedAt

	for _, Withdraw := range Withdraws {
		responses = append(responses, t.mapResponseWithdrawDeleteAt(Withdraw))
	}

	return responses
}

func (t *withdrawGraphqlMapper) mapResponseMonthStatusSuccess(data *pbStats.WithdrawMonthStatusSuccessResponse) *model.WithdrawMonthStatusSuccessResponse {
	return &model.WithdrawMonthStatusSuccessResponse{
		Year:         data.Year,
		Month:        data.Month,
		TotalSuccess: int32(data.TotalSuccess),
		TotalAmount:  int32(data.TotalAmount),
	}
}

func (t *withdrawGraphqlMapper) mapResponsesMonthStatusSuccess(Withdraws []*pbStats.WithdrawMonthStatusSuccessResponse) []*model.WithdrawMonthStatusSuccessResponse {
	var responses []*model.WithdrawMonthStatusSuccessResponse

	for _, Withdraw := range Withdraws {
		responses = append(responses, t.mapResponseMonthStatusSuccess(Withdraw))
	}

	return responses
}

func (t *withdrawGraphqlMapper) mapResponseYearStatusSuccess(data *pbStats.WithdrawYearStatusSuccessResponse) *model.WithdrawYearStatusSuccessResponse {
	return &model.WithdrawYearStatusSuccessResponse{
		Year:         data.Year,
		TotalSuccess: int32(data.TotalSuccess),
		TotalAmount:  int32(data.TotalAmount),
	}
}

func (t *withdrawGraphqlMapper) mapResponsesYearStatusSuccess(Withdraws []*pbStats.WithdrawYearStatusSuccessResponse) []*model.WithdrawYearStatusSuccessResponse {
	var responses []*model.WithdrawYearStatusSuccessResponse

	for _, Withdraw := range Withdraws {
		responses = append(responses, t.mapResponseYearStatusSuccess(Withdraw))
	}

	return responses
}

func (t *withdrawGraphqlMapper) mapResponseMonthStatusFailed(data *pbStats.WithdrawMonthStatusFailedResponse) *model.WithdrawMonthStatusFailedResponse {
	return &model.WithdrawMonthStatusFailedResponse{
		Year:        data.Year,
		Month:       data.Month,
		TotalFailed: int32(data.TotalFailed),
		TotalAmount: int32(data.TotalAmount),
	}
}

func (t *withdrawGraphqlMapper) mapResponsesMonthStatusFailed(Withdraws []*pbStats.WithdrawMonthStatusFailedResponse) []*model.WithdrawMonthStatusFailedResponse {
	var responses []*model.WithdrawMonthStatusFailedResponse

	for _, Withdraw := range Withdraws {
		responses = append(responses, t.mapResponseMonthStatusFailed(Withdraw))
	}

	return responses
}

func (t *withdrawGraphqlMapper) mapResponseYearStatusFailed(data *pbStats.WithdrawYearStatusFailedResponse) *model.WithdrawYearStatusFailedResponse {
	return &model.WithdrawYearStatusFailedResponse{
		Year:        data.Year,
		TotalFailed: int32(data.TotalFailed),
		TotalAmount: int32(data.TotalAmount),
	}
}

func (t *withdrawGraphqlMapper) mapResponsesYearStatusFailed(Withdraws []*pbStats.WithdrawYearStatusFailedResponse) []*model.WithdrawYearStatusFailedResponse {
	var responses []*model.WithdrawYearStatusFailedResponse

	for _, Withdraw := range Withdraws {
		responses = append(responses, t.mapResponseYearStatusFailed(Withdraw))
	}

	return responses
}

func (t *withdrawGraphqlMapper) mapResponseWithdrawMonthAmount(Withdraw *pbStats.WithdrawMonthlyAmountResponse) *model.WithdrawMonthAmountResponse {
	return &model.WithdrawMonthAmountResponse{
		Month:       Withdraw.Month,
		TotalAmount: int32(Withdraw.TotalAmount),
	}
}

func (t *withdrawGraphqlMapper) mapResponsesWithdrawMonthAmount(Withdraws []*pbStats.WithdrawMonthlyAmountResponse) []*model.WithdrawMonthAmountResponse {
	var responses []*model.WithdrawMonthAmountResponse

	for _, Withdraw := range Withdraws {
		responses = append(responses, t.mapResponseWithdrawMonthAmount(Withdraw))
	}

	return responses
}

func (t *withdrawGraphqlMapper) mapResponseWithdrawYearAmount(Withdraw *pbStats.WithdrawYearlyAmountResponse) *model.WithdrawYearAmountResponse {
	return &model.WithdrawYearAmountResponse{
		Year:        Withdraw.Year,
		TotalAmount: int32(Withdraw.TotalAmount),
	}
}

func (t *withdrawGraphqlMapper) mapResponsesWithdrawYearAmount(Withdraws []*pbStats.WithdrawYearlyAmountResponse) []*model.WithdrawYearAmountResponse {
	var responses []*model.WithdrawYearAmountResponse

	for _, Withdraw := range Withdraws {
		responses = append(responses, t.mapResponseWithdrawYearAmount(Withdraw))
	}

	return responses
}
