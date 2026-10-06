package merchantgraphqlmapper

import (
	graphqlmapper "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/mapper"
	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/merchant"
	pbStats "github.com/MamangRust/monolith-payment-gateway-pb/merchant/stats"
)

type merchantResponse struct{}

func NewMerchantResponseMapper() *merchantResponse {
	return &merchantResponse{}
}

func (m *merchantResponse) ToGraphqlResponseMerchant(res *pb.ApiResponseMerchant) *model.APIResponseMerchant {
	return &model.APIResponseMerchant{
		Status:  res.Status,
		Message: res.Message,
		Data:    m.mapMerchantResponse(res.Data),
	}
}

func (m *merchantResponse) ToGraphqlResponsePaginationMerchant(res *pb.ApiResponsePaginationMerchant) *model.APIResponsePaginationMerchant {
	return &model.APIResponsePaginationMerchant{
		Status:     res.Status,
		Message:    res.Message,
		Data:       m.mapMerchantResponses(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (m *merchantResponse) ToGraphqlResponsesMerchant(res *pb.ApiResponsesMerchant) *model.APIResponsesMerchant {
	return &model.APIResponsesMerchant{
		Status:  res.Status,
		Message: res.Message,
		Data:    m.mapMerchantResponses(res.Data),
	}
}

func (m *merchantResponse) ToGraphqlResponseMerchantDeleteAt(res *pb.ApiResponseMerchantDeleteAt) *model.APIResponseMerchantDeleteAt {
	return &model.APIResponseMerchantDeleteAt{
		Status:  res.Status,
		Message: res.Message,
		Data:    m.mapMerchantResponseDeleteAt(res.Data),
	}
}

func (m *merchantResponse) ToGraphqlResponsePaginationMerchantDeleteAt(res *pb.ApiResponsePaginationMerchantDeleteAt) *model.APIResponsePaginationMerchantDeleteAt {
	return &model.APIResponsePaginationMerchantDeleteAt{
		Status:     res.Status,
		Message:    res.Message,
		Data:       m.mapMerchantResponsesDeleteAt(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (m *merchantResponse) ToGraphqlResponsePaginationTransaction(res *pb.ApiResponsePaginationMerchantTransaction) *model.APIResponsePaginationMerchantTransaction {

	return &model.APIResponsePaginationMerchantTransaction{
		Status:     res.Status,
		Message:    res.Message,
		Data:       m.mapMerchantTransactionResponses(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (m *merchantResponse) ToGraphqlMonthlyPaymentMethods(res *pbStats.ApiResponseMerchantMonthlyPaymentMethod) *model.APIResponseMerchantMonthlyPaymentMethod {
	return &model.APIResponseMerchantMonthlyPaymentMethod{
		Status:  res.Status,
		Message: res.Message,
		Data:    m.mapResponsesMonthlyPaymentMethod(res.Data),
	}
}

func (m *merchantResponse) ToGraphqlYearlyPaymentMethods(res *pbStats.ApiResponseMerchantYearlyPaymentMethod) *model.APIResponseMerchantYearlyPaymentMethod {
	return &model.APIResponseMerchantYearlyPaymentMethod{
		Status:  res.Status,
		Message: res.Message,
		Data:    m.mapResponsesYearlyPaymentMethod(res.Data),
	}
}

func (m *merchantResponse) ToGraphqlMonthlyAmounts(res *pbStats.ApiResponseMerchantMonthlyAmount) *model.APIResponseMerchantMonthlyAmount {
	return &model.APIResponseMerchantMonthlyAmount{
		Status:  res.Status,
		Message: res.Message,
		Data:    m.mapResponsesMonthlyAmount(res.Data),
	}
}

func (m *merchantResponse) ToGraphqlYearlyAmounts(res *pbStats.ApiResponseMerchantYearlyAmount) *model.APIResponseMerchantYearlyAmount {
	return &model.APIResponseMerchantYearlyAmount{
		Status:  res.Status,
		Message: res.Message,
		Data:    m.mapResponsesYearlyAmount(res.Data),
	}
}

func (m *merchantResponse) ToGraphqlMonthlyTotalAmounts(res *pbStats.ApiResponseMerchantMonthlyTotalAmount) *model.APIResponseMerchantMonthlyTotalAmount {
	return &model.APIResponseMerchantMonthlyTotalAmount{
		Status:  res.Status,
		Message: res.Message,
		Data:    m.mapResponsesMonthlyTotalAmount(res.Data),
	}
}

func (m *merchantResponse) ToGraphqlYearlyTotalAmounts(res *pbStats.ApiResponseMerchantYearlyTotalAmount) *model.APIResponseMerchantYearlyTotalAmount {
	return &model.APIResponseMerchantYearlyTotalAmount{
		Status:  res.Status,
		Message: res.Message,
		Data:    m.mapResponsesYearlyTotalAmount(res.Data),
	}
}

func (s *merchantResponse) ToGraphqlMerchantDeleteAll(res *pb.ApiResponseMerchantDelete) *model.APIResponseMerchantDelete {
	return &model.APIResponseMerchantDelete{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (s *merchantResponse) ToGraphqlMerchantAll(res *pb.ApiResponseMerchantAll) *model.APIResponseMerchantAll {
	return &model.APIResponseMerchantAll{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (m *merchantResponse) mapMerchantResponse(merchant *pb.MerchantResponse) *model.MerchantResponse {
	return &model.MerchantResponse{
		ID:        int32(merchant.Id),
		Name:      merchant.Name,
		UserID:    int32(merchant.UserId),
		Status:    merchant.Status,
		APIKey:    merchant.ApiKey,
		CreatedAt: merchant.CreatedAt,
		UpdatedAt: merchant.UpdatedAt,
	}
}

func (m *merchantResponse) mapMerchantResponses(merchants []*pb.MerchantResponse) []*model.MerchantResponse {
	var responseMerchants []*model.MerchantResponse

	for _, merchant := range merchants {
		responseMerchants = append(responseMerchants, m.mapMerchantResponse(merchant))
	}

	return responseMerchants
}

func (m *merchantResponse) mapMerchantResponseDeleteAt(merchant *pb.MerchantResponseDeleteAt) *model.MerchantResponseDeletedAt {
	var deletedAt string

	if merchant.DeletedAt != nil {
		deletedAt = merchant.DeletedAt.Value
	}

	return &model.MerchantResponseDeletedAt{
		ID:        int32(merchant.Id),
		Name:      merchant.Name,
		UserID:    int32(merchant.UserId),
		Status:    merchant.Status,
		APIKey:    merchant.ApiKey,
		CreatedAt: merchant.CreatedAt,
		UpdatedAt: merchant.UpdatedAt,
		DeletedAt: &deletedAt,
	}
}

func (m *merchantResponse) mapMerchantResponsesDeleteAt(merchants []*pb.MerchantResponseDeleteAt) []*model.MerchantResponseDeletedAt {
	var responseMerchants []*model.MerchantResponseDeletedAt

	for _, merchant := range merchants {
		responseMerchants = append(responseMerchants, m.mapMerchantResponseDeleteAt(merchant))
	}

	return responseMerchants
}

func (m *merchantResponse) mapMerchantTransactionResponse(merchant *pb.MerchantTransactionResponse) *model.MerchantTransactionResponse {
	return &model.MerchantTransactionResponse{
		ID:              int32(merchant.Id),
		CardNumber:      merchant.CardNumber,
		Amount:          merchant.Amount,
		PaymentMethod:   merchant.PaymentMethod,
		MerchantID:      merchant.MerchantId,
		MerchantName:    merchant.MerchantName,
		TransactionTime: merchant.TransactionTime,
		CreatedAt:       merchant.CreatedAt,
		UpdatedAt:       merchant.UpdatedAt,
	}
}

func (m *merchantResponse) mapMerchantTransactionResponses(r []*pb.MerchantTransactionResponse) []*model.MerchantTransactionResponse {
	var responseMerchants []*model.MerchantTransactionResponse
	for _, merchant := range r {
		responseMerchants = append(responseMerchants, m.mapMerchantTransactionResponse(merchant))
	}

	return responseMerchants
}

func (m *merchantResponse) mapResponseMonthlyPaymentMethod(ms *pbStats.MerchantResponseMonthlyPaymentMethod) *model.MerchantMonthlyPaymentMethodResponse {
	return &model.MerchantMonthlyPaymentMethodResponse{
		Month:         ms.Month,
		PaymentMethod: ms.PaymentMethod,
		TotalAmount:   int32(ms.TotalAmount),
	}
}

func (m *merchantResponse) mapResponsesMonthlyPaymentMethod(r []*pbStats.MerchantResponseMonthlyPaymentMethod) []*model.MerchantMonthlyPaymentMethodResponse {
	var responseMerchants []*model.MerchantMonthlyPaymentMethodResponse
	for _, merchant := range r {
		responseMerchants = append(responseMerchants, m.mapResponseMonthlyPaymentMethod(merchant))
	}

	return responseMerchants
}

func (m *merchantResponse) mapResponseYearlyPaymentMethod(ms *pbStats.MerchantResponseYearlyPaymentMethod) *model.MerchantYearlyPaymentMethodResponse {
	return &model.MerchantYearlyPaymentMethodResponse{
		Year:          ms.Year,
		PaymentMethod: ms.PaymentMethod,
		TotalAmount:   int32(ms.TotalAmount),
	}
}

func (m *merchantResponse) mapResponsesYearlyPaymentMethod(r []*pbStats.MerchantResponseYearlyPaymentMethod) []*model.MerchantYearlyPaymentMethodResponse {
	var responseMerchants []*model.MerchantYearlyPaymentMethodResponse
	for _, merchant := range r {
		responseMerchants = append(responseMerchants, m.mapResponseYearlyPaymentMethod(merchant))
	}

	return responseMerchants
}

func (m *merchantResponse) mapResponseMonthlyAmount(ms *pbStats.MerchantResponseMonthlyAmount) *model.MerchantMonthlyAmountResponse {
	return &model.MerchantMonthlyAmountResponse{
		Month:       ms.Month,
		TotalAmount: int32(ms.TotalAmount),
	}
}

func (m *merchantResponse) mapResponsesMonthlyAmount(r []*pbStats.MerchantResponseMonthlyAmount) []*model.MerchantMonthlyAmountResponse {
	var responseMerchants []*model.MerchantMonthlyAmountResponse
	for _, merchant := range r {
		responseMerchants = append(responseMerchants, m.mapResponseMonthlyAmount(merchant))
	}

	return responseMerchants
}

func (m *merchantResponse) mapResponseYearlyAmount(ms *pbStats.MerchantResponseYearlyAmount) *model.MerchantYearlyAmountResponse {
	return &model.MerchantYearlyAmountResponse{
		Year:        ms.Year,
		TotalAmount: int32(ms.TotalAmount),
	}
}

func (m *merchantResponse) mapResponsesYearlyAmount(r []*pbStats.MerchantResponseYearlyAmount) []*model.MerchantYearlyAmountResponse {
	var responseMerchants []*model.MerchantYearlyAmountResponse
	for _, merchant := range r {
		responseMerchants = append(responseMerchants, m.mapResponseYearlyAmount(merchant))
	}

	return responseMerchants
}

func (m *merchantResponse) mapResponseMonthlyTotalAmount(ms *pbStats.MerchantResponseMonthlyTotalAmount) *model.MerchantMonthlyTotalAmountResponse {
	return &model.MerchantMonthlyTotalAmountResponse{
		Month:       ms.Month,
		Year:        ms.Year,
		TotalAmount: int32(ms.TotalAmount),
	}
}

func (m *merchantResponse) mapResponsesMonthlyTotalAmount(r []*pbStats.MerchantResponseMonthlyTotalAmount) []*model.MerchantMonthlyTotalAmountResponse {
	var responseMerchants []*model.MerchantMonthlyTotalAmountResponse
	for _, merchant := range r {
		responseMerchants = append(responseMerchants, m.mapResponseMonthlyTotalAmount(merchant))
	}

	return responseMerchants
}

func (m *merchantResponse) mapResponseYearlyTotalAmount(ms *pbStats.MerchantResponseYearlyTotalAmount) *model.MerchantYearlyTotalAmountResponse {
	return &model.MerchantYearlyTotalAmountResponse{
		Year:        ms.Year,
		TotalAmount: int32(ms.TotalAmount),
	}
}

func (m *merchantResponse) mapResponsesYearlyTotalAmount(r []*pbStats.MerchantResponseYearlyTotalAmount) []*model.MerchantYearlyTotalAmountResponse {
	var responseMerchants []*model.MerchantYearlyTotalAmountResponse
	for _, merchant := range r {
		responseMerchants = append(responseMerchants, m.mapResponseYearlyTotalAmount(merchant))
	}

	return responseMerchants
}
