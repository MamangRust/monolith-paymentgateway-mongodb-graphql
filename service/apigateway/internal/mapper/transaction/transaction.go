package transactiongraphqlmapper

import (
	graphqlmapper "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/mapper"
	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/transaction"
	pbStats "github.com/MamangRust/monolith-payment-gateway-pb/transaction/stats"
)

type transactionGraphqlMapper struct {
}

func NewTransactionGraphqlMapper() *transactionGraphqlMapper {
	return &transactionGraphqlMapper{}
}

func (t *transactionGraphqlMapper) ToGraphqlResponseTransactionMonthStatusSuccess(res *pbStats.ApiResponseTransactionMonthStatusSuccess) *model.APIResponseTransactionMonthStatusSuccess {
	return &model.APIResponseTransactionMonthStatusSuccess{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesTransactionMonthStatusSuccess(res.Data),
	}
}

func (t *transactionGraphqlMapper) ToGraphqlResponseTransactionYearStatusSuccess(res *pbStats.ApiResponseTransactionYearStatusSuccess) *model.APIResponseTransactionYearStatusSuccess {
	return &model.APIResponseTransactionYearStatusSuccess{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesTransactionYearStatusSuccess(res.Data),
	}
}

func (t *transactionGraphqlMapper) ToGraphqlResponseTransactionMonthStatusFailed(res *pbStats.ApiResponseTransactionMonthStatusFailed) *model.APIResponseTransactionMonthStatusFailed {
	return &model.APIResponseTransactionMonthStatusFailed{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesTransactionMonthStatusFailed(res.Data),
	}
}

func (t *transactionGraphqlMapper) ToGraphqlResponseTransactionYearStatusFailed(res *pbStats.ApiResponseTransactionYearStatusFailed) *model.APIResponseTransactionYearStatusFailed {
	return &model.APIResponseTransactionYearStatusFailed{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesTransactionYearStatusFailed(res.Data),
	}
}

func (t *transactionGraphqlMapper) ToGraphqlResponseTransactionMonthMethod(res *pbStats.ApiResponseTransactionMonthMethod) *model.APIResponseTransactionMonthMethod {
	return &model.APIResponseTransactionMonthMethod{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesTransactionMonthlyMethod(res.Data),
	}
}

func (t *transactionGraphqlMapper) ToGraphqlResponseTransactionYearMethod(res *pbStats.ApiResponseTransactionYearMethod) *model.APIResponseTransactionYearMethod {
	return &model.APIResponseTransactionYearMethod{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesTransactionYearlyMethod(res.Data),
	}
}

func (t *transactionGraphqlMapper) ToGraphqlResponseTransactionMonthAmount(res *pbStats.ApiResponseTransactionMonthAmount) *model.APIResponseTransactionMonthAmount {
	return &model.APIResponseTransactionMonthAmount{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesTransactionMonthlyAmount(res.Data),
	}
}

func (t *transactionGraphqlMapper) ToGraphqlResponseTransactionYearAmount(res *pbStats.ApiResponseTransactionYearAmount) *model.APIResponseTransactionYearAmount {
	return &model.APIResponseTransactionYearAmount{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesTransactionYearlyAmount(res.Data),
	}
}

func (t *transactionGraphqlMapper) ToGraphqlTransactionAll(res *pb.ApiResponseTransactionAll) *model.APIResponseTransactionAll {
	return &model.APIResponseTransactionAll{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (t *transactionGraphqlMapper) ToGraphqlTransactionDelete(res *pb.ApiResponseTransactionDelete) *model.APIResponseTransactionDelete {
	return &model.APIResponseTransactionDelete{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (t *transactionGraphqlMapper) ToGraphqlResponseTransaction(res *pb.ApiResponseTransaction) *model.APIResponseTransaction {
	return &model.APIResponseTransaction{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponseTransaction(res.Data),
	}
}

func (t *transactionGraphqlMapper) ToGraphqlResponseTransactions(res *pb.ApiResponseTransactions) *model.APIResponseTransactions {
	return &model.APIResponseTransactions{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesTransaction(res.Data),
	}
}

func (t *transactionGraphqlMapper) ToGraphqlResponseTransactionDeleteAt(res *pb.ApiResponseTransactionDeleteAt) *model.APIResponseTransactionDeleteAt {
	return &model.APIResponseTransactionDeleteAt{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponseTransactionDeleteAt(res.Data),
	}
}

func (t *transactionGraphqlMapper) ToGraphqlPaginationTransaction(res *pb.ApiResponsePaginationTransaction) *model.APIResponsePaginationTransaction {
	return &model.APIResponsePaginationTransaction{
		Status:     res.Status,
		Message:    res.Message,
		Data:       t.mapResponsesTransaction(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (t *transactionGraphqlMapper) ToGraphqlPaginationTransactionDeleteAt(res *pb.ApiResponsePaginationTransactionDeleteAt) *model.APIResponsePaginationTransactionDeleteAt {
	return &model.APIResponsePaginationTransactionDeleteAt{
		Status:     res.Status,
		Message:    res.Message,
		Data:       t.mapResponsesTransactionDeleteAt(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (t *transactionGraphqlMapper) mapResponseTransaction(transaction *pb.TransactionResponse) *model.TransactionResponse {
	return &model.TransactionResponse{
		ID:              int32(transaction.Id),
		TransactionNo:   transaction.TransactionNo,
		CardNumber:      transaction.CardNumber,
		Amount:          int32(transaction.Amount),
		PaymentMethod:   transaction.PaymentMethod,
		TransactionTime: transaction.TransactionTime,
		MerchantID:      int32(transaction.MerchantId),
		CreatedAt:       transaction.CreatedAt,
		UpdatedAt:       transaction.UpdatedAt,
	}
}

func (t *transactionGraphqlMapper) mapResponsesTransaction(transactions []*pb.TransactionResponse) []*model.TransactionResponse {
	var result []*model.TransactionResponse

	for _, transaction := range transactions {
		result = append(result, t.mapResponseTransaction(transaction))
	}

	return result
}

func (t *transactionGraphqlMapper) mapResponseTransactionDeleteAt(transaction *pb.TransactionResponseDeleteAt) *model.TransactionResponseDeletedAt {
	var deletedAt string

	if transaction.DeletedAt != nil {
		deletedAt = transaction.DeletedAt.Value
	}

	return &model.TransactionResponseDeletedAt{
		ID:              int32(transaction.Id),
		TransactionNo:   transaction.TransactionNo,
		CardNumber:      transaction.CardNumber,
		Amount:          int32(transaction.Amount),
		PaymentMethod:   transaction.PaymentMethod,
		TransactionTime: transaction.TransactionTime,
		MerchantID:      int32(transaction.MerchantId),
		CreatedAt:       transaction.CreatedAt,
		UpdatedAt:       transaction.UpdatedAt,
		DeletedAt:       &deletedAt,
	}
}

func (t *transactionGraphqlMapper) mapResponsesTransactionDeleteAt(transactions []*pb.TransactionResponseDeleteAt) []*model.TransactionResponseDeletedAt {
	var result []*model.TransactionResponseDeletedAt

	for _, transaction := range transactions {
		result = append(result, t.mapResponseTransactionDeleteAt(transaction))
	}

	return result
}

func (t *transactionGraphqlMapper) mapResponseTransactionMonthStatusSuccess(s *pbStats.TransactionMonthStatusSuccessResponse) *model.TransactionMonthStatusSuccessResponse {
	return &model.TransactionMonthStatusSuccessResponse{
		Year:         s.Year,
		Month:        s.Month,
		TotalSuccess: int32(s.TotalSuccess),
		TotalAmount:  int32(s.TotalAmount),
	}
}

func (t *transactionGraphqlMapper) mapResponsesTransactionMonthStatusSuccess(Transactions []*pbStats.TransactionMonthStatusSuccessResponse) []*model.TransactionMonthStatusSuccessResponse {
	var responses []*model.TransactionMonthStatusSuccessResponse

	for _, Transaction := range Transactions {
		responses = append(responses, t.mapResponseTransactionMonthStatusSuccess(Transaction))
	}

	return responses
}

func (t *transactionGraphqlMapper) mapResponseTransactionMonthStatusFailed(s *pbStats.TransactionMonthStatusFailedResponse) *model.TransactionMonthStatusFailedResponse {
	return &model.TransactionMonthStatusFailedResponse{
		Year:        s.Year,
		Month:       s.Month,
		TotalFailed: int32(s.TotalFailed),
		TotalAmount: int32(s.TotalAmount),
	}
}

func (t *transactionGraphqlMapper) mapResponsesTransactionMonthStatusFailed(Transactions []*pbStats.TransactionMonthStatusFailedResponse) []*model.TransactionMonthStatusFailedResponse {
	var responses []*model.TransactionMonthStatusFailedResponse

	for _, Transaction := range Transactions {
		responses = append(responses, t.mapResponseTransactionMonthStatusFailed(Transaction))
	}

	return responses
}

func (t *transactionGraphqlMapper) mapResponseTransactionYearStatusSuccess(s *pbStats.TransactionYearStatusSuccessResponse) *model.TransactionYearStatusSuccessResponse {
	return &model.TransactionYearStatusSuccessResponse{
		Year:         s.Year,
		TotalSuccess: int32(s.TotalSuccess),
		TotalAmount:  int32(s.TotalAmount),
	}
}

func (t *transactionGraphqlMapper) mapResponsesTransactionYearStatusSuccess(Transactions []*pbStats.TransactionYearStatusSuccessResponse) []*model.TransactionYearStatusSuccessResponse {
	var responses []*model.TransactionYearStatusSuccessResponse

	for _, Transaction := range Transactions {
		responses = append(responses, t.mapResponseTransactionYearStatusSuccess(Transaction))
	}

	return responses
}

func (t *transactionGraphqlMapper) mapResponseTransactionYearStatusFailed(s *pbStats.TransactionYearStatusFailedResponse) *model.TransactionYearStatusFailedResponse {
	return &model.TransactionYearStatusFailedResponse{
		Year:        s.Year,
		TotalFailed: int32(s.TotalFailed),
		TotalAmount: int32(s.TotalAmount),
	}
}

func (t *transactionGraphqlMapper) mapResponsesTransactionYearStatusFailed(Transactions []*pbStats.TransactionYearStatusFailedResponse) []*model.TransactionYearStatusFailedResponse {
	var responses []*model.TransactionYearStatusFailedResponse

	for _, Transaction := range Transactions {
		responses = append(responses, t.mapResponseTransactionYearStatusFailed(Transaction))
	}

	return responses
}

func (t *transactionGraphqlMapper) mapResponseTransactionMonthlyMethod(s *pbStats.TransactionMonthMethodResponse) *model.TransactionMonthMethodResponse {
	return &model.TransactionMonthMethodResponse{
		Month:             s.Month,
		PaymentMethod:     s.PaymentMethod,
		TotalTransactions: int32(s.TotalTransactions),
		TotalAmount:       int32(s.TotalAmount),
	}
}

func (s *transactionGraphqlMapper) mapResponsesTransactionMonthlyMethod(Transactions []*pbStats.TransactionMonthMethodResponse) []*model.TransactionMonthMethodResponse {
	var responses []*model.TransactionMonthMethodResponse

	for _, Transaction := range Transactions {
		responses = append(responses, s.mapResponseTransactionMonthlyMethod(Transaction))
	}

	return responses
}

func (t *transactionGraphqlMapper) mapResponseTransactionYearlyMethod(s *pbStats.TransactionYearMethodResponse) *model.TransactionYearMethodResponse {
	return &model.TransactionYearMethodResponse{
		Year:              s.Year,
		PaymentMethod:     s.PaymentMethod,
		TotalTransactions: int32(s.TotalTransactions),
		TotalAmount:       int32(s.TotalAmount),
	}
}

func (s *transactionGraphqlMapper) mapResponsesTransactionYearlyMethod(Transactions []*pbStats.TransactionYearMethodResponse) []*model.TransactionYearMethodResponse {
	var responses []*model.TransactionYearMethodResponse

	for _, Transaction := range Transactions {
		responses = append(responses, s.mapResponseTransactionYearlyMethod(Transaction))
	}

	return responses
}

func (t *transactionGraphqlMapper) mapResponseTransactionMonthlyAmount(s *pbStats.TransactionMonthAmountResponse) *model.TransactionMonthAmountResponse {
	return &model.TransactionMonthAmountResponse{
		Month:       s.Month,
		TotalAmount: int32(s.TotalAmount),
	}
}

func (s *transactionGraphqlMapper) mapResponsesTransactionMonthlyAmount(Transactions []*pbStats.TransactionMonthAmountResponse) []*model.TransactionMonthAmountResponse {
	var responses []*model.TransactionMonthAmountResponse

	for _, Transaction := range Transactions {
		responses = append(responses, s.mapResponseTransactionMonthlyAmount(Transaction))
	}

	return responses
}

func (t *transactionGraphqlMapper) mapResponseTransactionYearlyAmount(s *pbStats.TransactionYearlyAmountResponse) *model.TransactionYearAmountResponse {
	return &model.TransactionYearAmountResponse{
		Year:        s.Year,
		TotalAmount: int32(s.TotalAmount),
	}
}

func (s *transactionGraphqlMapper) mapResponsesTransactionYearlyAmount(Transactions []*pbStats.TransactionYearlyAmountResponse) []*model.TransactionYearAmountResponse {
	var responses []*model.TransactionYearAmountResponse

	for _, Transaction := range Transactions {
		responses = append(responses, s.mapResponseTransactionYearlyAmount(Transaction))
	}

	return responses
}
