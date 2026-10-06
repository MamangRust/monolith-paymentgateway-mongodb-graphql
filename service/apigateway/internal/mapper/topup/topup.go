package topupgraphqlmapper

import (
	graphqlmapper "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/mapper"
	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/topup"
	pbStats "github.com/MamangRust/monolith-payment-gateway-pb/topup/stats"
)

type topupGraphqlMapper struct {
}

func NewTopupGraphqlMapper() *topupGraphqlMapper {
	return &topupGraphqlMapper{}
}

func (t *topupGraphqlMapper) ToGraphqlResponseTopup(res *pb.ApiResponseTopup) *model.APIResponseTopup {
	return &model.APIResponseTopup{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponseTopup(res.Data),
	}
}

func (t *topupGraphqlMapper) ToGraphqlResponseTopupDeleteAt(res *pb.ApiResponseTopupDeleteAt) *model.APIResponseTopupDeleteAt {
	return &model.APIResponseTopupDeleteAt{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponseTopupDeleteAt(res.Data),
	}
}

func (t *topupGraphqlMapper) ToGraphqlTopupAll(res *pb.ApiResponseTopupAll) *model.APIResponseTopupAll {
	return &model.APIResponseTopupAll{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (t *topupGraphqlMapper) ToGraphqlTopupDelete(res *pb.ApiResponseTopupDelete) *model.APIResponseTopupDelete {
	return &model.APIResponseTopupDelete{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (t *topupGraphqlMapper) ToGraphqlResponsePaginationTopup(res *pb.ApiResponsePaginationTopup) *model.APIResponsePaginationTopup {
	return &model.APIResponsePaginationTopup{
		Status:     res.Status,
		Message:    res.Message,
		Data:       t.mapResponsesTopup(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (t *topupGraphqlMapper) ToGraphqlResponsePaginationTopupDeleteAt(res *pb.ApiResponsePaginationTopupDeleteAt) *model.APIResponsePaginationTopupDeleteAt {
	return &model.APIResponsePaginationTopupDeleteAt{
		Status:     res.Status,
		Message:    res.Message,
		Data:       t.mapResponsesTopupDeleteAt(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (t *topupGraphqlMapper) ToGraphqlResponseTopupMonthStatusSuccess(res *pbStats.ApiResponseTopupMonthStatusSuccess) *model.APIResponseTopupMonthStatusSuccess {
	return &model.APIResponseTopupMonthStatusSuccess{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesTopupMonthStatusSuccess(res.Data),
	}
}

func (t *topupGraphqlMapper) ToGraphqlResponseTopupYearStatusSuccess(res *pbStats.ApiResponseTopupYearStatusSuccess) *model.APIResponseTopupYearStatusSuccess {
	return &model.APIResponseTopupYearStatusSuccess{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesTopupYearStatusSuccess(res.Data),
	}
}

func (t *topupGraphqlMapper) ToGraphqlResponseTopupMonthStatusFailed(res *pbStats.ApiResponseTopupMonthStatusFailed) *model.APIResponseTopupMonthStatusFailed {
	return &model.APIResponseTopupMonthStatusFailed{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesTopupMonthStatusFailed(res.Data),
	}
}

func (t *topupGraphqlMapper) ToGraphqlResponseTopupYearStatusFailed(res *pbStats.ApiResponseTopupYearStatusFailed) *model.APIResponseTopupYearStatusFailed {
	return &model.APIResponseTopupYearStatusFailed{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesTopupYearStatusFailed(res.Data),
	}
}

func (t *topupGraphqlMapper) ToGraphqlResponseTopupMonthMethod(res *pbStats.ApiResponseTopupMonthMethod) *model.APIResponseTopupMonthMethod {
	return &model.APIResponseTopupMonthMethod{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesTopupMonthlyMethod(res.Data),
	}
}

func (t *topupGraphqlMapper) ToGraphqlResponseTopupYearMethod(res *pbStats.ApiResponseTopupYearMethod) *model.APIResponseTopupYearMethod {
	return &model.APIResponseTopupYearMethod{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesTopupYearlyMethod(res.Data),
	}
}

func (t *topupGraphqlMapper) ToGraphqlResponseTopupMonthAmount(res *pbStats.ApiResponseTopupMonthAmount) *model.APIResponseTopupMonthAmount {
	return &model.APIResponseTopupMonthAmount{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesTopupMonthlyAmount(res.Data),
	}
}

func (t *topupGraphqlMapper) ToGraphqlResponseTopupYearAmount(res *pbStats.ApiResponseTopupYearAmount) *model.APIResponseTopupYearAmount {
	return &model.APIResponseTopupYearAmount{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesTopupYearlyAmount(res.Data),
	}
}

func (t *topupGraphqlMapper) mapResponseTopup(topup *pb.TopupResponse) *model.TopupResponse {
	return &model.TopupResponse{
		ID:          int32(topup.Id),
		CardNumber:  topup.CardNumber,
		TopupNo:     topup.TopupNo,
		TopupAmount: int32(topup.TopupAmount),
		TopupMethod: topup.TopupMethod,
		TopupTime:   &topup.TopupTime,
		CreatedAt:   topup.CreatedAt,
		UpdatedAt:   topup.UpdatedAt,
	}
}

func (t *topupGraphqlMapper) mapResponsesTopup(topups []*pb.TopupResponse) []*model.TopupResponse {
	var responses []*model.TopupResponse

	for _, topup := range topups {
		responses = append(responses, t.mapResponseTopup(topup))
	}

	return responses
}

func (t *topupGraphqlMapper) mapResponseTopupDeleteAt(topup *pb.TopupResponseDeleteAt) *model.TopupResponseDeletedAt {
	var deletedAt string

	if topup.DeletedAt != nil {
		deletedAt = topup.DeletedAt.Value
	}

	return &model.TopupResponseDeletedAt{
		ID:          int32(topup.Id),
		CardNumber:  topup.CardNumber,
		TopupNo:     topup.TopupNo,
		TopupAmount: int32(topup.TopupAmount),
		TopupMethod: topup.TopupMethod,
		TopupTime:   &topup.TopupTime,
		CreatedAt:   topup.CreatedAt,
		UpdatedAt:   topup.UpdatedAt,
		DeletedAt:   &deletedAt,
	}
}

func (t *topupGraphqlMapper) mapResponsesTopupDeleteAt(topups []*pb.TopupResponseDeleteAt) []*model.TopupResponseDeletedAt {
	var responses []*model.TopupResponseDeletedAt

	for _, topup := range topups {
		responses = append(responses, t.mapResponseTopupDeleteAt(topup))
	}

	return responses
}

func (t *topupGraphqlMapper) mapResponseTopupMonthStatusSuccess(s *pbStats.TopupMonthStatusSuccessResponse) *model.TopupMonthStatusSuccessResponse {
	return &model.TopupMonthStatusSuccessResponse{
		Year:         s.Year,
		Month:        s.Month,
		TotalSuccess: int32(s.TotalSuccess),
		TotalAmount:  int32(s.TotalAmount),
	}
}

func (t *topupGraphqlMapper) mapResponsesTopupMonthStatusSuccess(topups []*pbStats.TopupMonthStatusSuccessResponse) []*model.TopupMonthStatusSuccessResponse {
	var responses []*model.TopupMonthStatusSuccessResponse

	for _, topup := range topups {
		responses = append(responses, t.mapResponseTopupMonthStatusSuccess(topup))
	}

	return responses
}

func (t *topupGraphqlMapper) mapResponseTopupMonthStatusFailed(s *pbStats.TopupMonthStatusFailedResponse) *model.TopupMonthStatusFailedResponse {
	return &model.TopupMonthStatusFailedResponse{
		Year:        s.Year,
		Month:       s.Month,
		TotalFailed: int32(s.TotalFailed),
		TotalAmount: int32(s.TotalAmount),
	}
}

func (t *topupGraphqlMapper) mapResponsesTopupMonthStatusFailed(topups []*pbStats.TopupMonthStatusFailedResponse) []*model.TopupMonthStatusFailedResponse {
	var responses []*model.TopupMonthStatusFailedResponse

	for _, topup := range topups {
		responses = append(responses, t.mapResponseTopupMonthStatusFailed(topup))
	}

	return responses
}

func (t *topupGraphqlMapper) mapResponseTopupYearStatusSuccess(s *pbStats.TopupYearStatusSuccessResponse) *model.TopupYearStatusSuccessResponse {
	return &model.TopupYearStatusSuccessResponse{
		Year:         s.Year,
		TotalSuccess: int32(s.TotalSuccess),
		TotalAmount:  int32(s.TotalAmount),
	}
}

func (t *topupGraphqlMapper) mapResponsesTopupYearStatusSuccess(topups []*pbStats.TopupYearStatusSuccessResponse) []*model.TopupYearStatusSuccessResponse {
	var responses []*model.TopupYearStatusSuccessResponse

	for _, topup := range topups {
		responses = append(responses, t.mapResponseTopupYearStatusSuccess(topup))
	}

	return responses
}

func (t *topupGraphqlMapper) mapResponseTopupYearStatusFailed(s *pbStats.TopupYearStatusFailedResponse) *model.TopupYearStatusFailedResponse {
	return &model.TopupYearStatusFailedResponse{
		Year:        s.Year,
		TotalFailed: int32(s.TotalFailed),
		TotalAmount: int32(s.TotalAmount),
	}
}

func (t *topupGraphqlMapper) mapResponsesTopupYearStatusFailed(topups []*pbStats.TopupYearStatusFailedResponse) []*model.TopupYearStatusFailedResponse {
	var responses []*model.TopupYearStatusFailedResponse

	for _, topup := range topups {
		responses = append(responses, t.mapResponseTopupYearStatusFailed(topup))
	}

	return responses
}

func (t *topupGraphqlMapper) mapResponseTopupMonthlyMethod(s *pbStats.TopupMonthMethodResponse) *model.TopupMonthMethodResponse {
	return &model.TopupMonthMethodResponse{
		Month:       s.Month,
		TopupMethod: s.TopupMethod,
		TotalTopups: int32(s.TotalTopups),
		TotalAmount: int32(s.TotalAmount),
	}
}

func (s *topupGraphqlMapper) mapResponsesTopupMonthlyMethod(topups []*pbStats.TopupMonthMethodResponse) []*model.TopupMonthMethodResponse {
	var responses []*model.TopupMonthMethodResponse

	for _, topup := range topups {
		responses = append(responses, s.mapResponseTopupMonthlyMethod(topup))
	}

	return responses
}

func (t *topupGraphqlMapper) mapResponseTopupYearlyMethod(s *pbStats.TopupYearlyMethodResponse) *model.TopupYearMethodResponse {
	return &model.TopupYearMethodResponse{
		Year:        s.Year,
		TopupMethod: s.TopupMethod,
		TotalTopups: int32(s.TotalTopups),
		TotalAmount: int32(s.TotalAmount),
	}
}

func (s *topupGraphqlMapper) mapResponsesTopupYearlyMethod(topups []*pbStats.TopupYearlyMethodResponse) []*model.TopupYearMethodResponse {
	var responses []*model.TopupYearMethodResponse

	for _, topup := range topups {
		responses = append(responses, s.mapResponseTopupYearlyMethod(topup))
	}

	return responses
}

func (t *topupGraphqlMapper) mapResponseTopupMonthlyAmount(s *pbStats.TopupMonthAmountResponse) *model.TopupMonthAmountResponse {
	return &model.TopupMonthAmountResponse{
		Month:       s.Month,
		TotalAmount: int32(s.TotalAmount),
	}
}

func (s *topupGraphqlMapper) mapResponsesTopupMonthlyAmount(topups []*pbStats.TopupMonthAmountResponse) []*model.TopupMonthAmountResponse {
	var responses []*model.TopupMonthAmountResponse

	for _, topup := range topups {
		responses = append(responses, s.mapResponseTopupMonthlyAmount(topup))
	}

	return responses
}

func (t *topupGraphqlMapper) mapResponseTopupYearlyAmount(s *pbStats.TopupYearlyAmountResponse) *model.TopupYearAmountResponse {
	return &model.TopupYearAmountResponse{
		Year:        s.Year,
		TotalAmount: int32(s.TotalAmount),
	}
}

func (s *topupGraphqlMapper) mapResponsesTopupYearlyAmount(topups []*pbStats.TopupYearlyAmountResponse) []*model.TopupYearAmountResponse {
	var responses []*model.TopupYearAmountResponse

	for _, topup := range topups {
		responses = append(responses, s.mapResponseTopupYearlyAmount(topup))
	}

	return responses
}
