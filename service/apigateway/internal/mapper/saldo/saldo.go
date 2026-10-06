package saldographqlmapper

import (
	graphqlmapper "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/mapper"
	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/saldo"
	pbStats "github.com/MamangRust/monolith-payment-gateway-pb/saldo/stats"
)

type saldoGraphqlMapper struct {
}

func NewSaldoGraphqlMapper() *saldoGraphqlMapper {
	return &saldoGraphqlMapper{}
}

func (s *saldoGraphqlMapper) ToGraphqlResponseSaldo(res *pb.ApiResponseSaldo) *model.APIResponseSaldo {
	return &model.APIResponseSaldo{
		Status:  res.Status,
		Message: res.Message,
		Data:    s.mapResponseSaldo(res.Data),
	}
}

func (s *saldoGraphqlMapper) ToGraphqlResponseSaldoDeleteAt(res *pb.ApiResponseSaldoDeleteAt) *model.APIResponseSaldoDeleteAt {
	return &model.APIResponseSaldoDeleteAt{
		Status:  res.Status,
		Message: res.Message,
		Data:    s.mapResponseSaldoDeleteAt(res.Data),
	}
}

func (s *saldoGraphqlMapper) ToGraphqlResponsePaginationSaldo(res *pb.ApiResponsePaginationSaldo) *model.APIResponsePaginationSaldo {
	return &model.APIResponsePaginationSaldo{
		Status:     res.Status,
		Message:    res.Message,
		Data:       s.mapResponsesSaldo(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (s *saldoGraphqlMapper) ToGraphqlResponsePaginationSaldoDeleteAt(res *pb.ApiResponsePaginationSaldoDeleteAt) *model.APIResponsePaginationSaldoDeleteAt {
	return &model.APIResponsePaginationSaldoDeleteAt{
		Status:     res.Status,
		Message:    res.Message,
		Data:       s.mapResponsesSaldoDeleteAt(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (s *saldoGraphqlMapper) ToGraphqlResponseDelete(res *pb.ApiResponseSaldoDelete) *model.APIResponseSaldoDelete {
	return &model.APIResponseSaldoDelete{
		Status:  res.Status,
		Message: res.Message,
	}
}
func (s *saldoGraphqlMapper) ToGraphqlResponseAll(res *pb.ApiResponseSaldoAll) *model.APIResponseSaldoAll {
	return &model.APIResponseSaldoAll{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (s *saldoGraphqlMapper) ToGraphqlResponseMonthTotalSaldo(res *pbStats.ApiResponseMonthTotalSaldo) *model.APIResponseMonthTotalSaldo {
	return &model.APIResponseMonthTotalSaldo{
		Status:  res.Status,
		Message: res.Message,
		Data:    s.mapSaldoMonthTotalBalanceResponses(res.Data),
	}
}

func (s *saldoGraphqlMapper) ToGraphqlResponseYearTotalSaldo(res *pbStats.ApiResponseYearTotalSaldo) *model.APIResponseYearTotalSaldo {
	return &model.APIResponseYearTotalSaldo{
		Status:  res.Status,
		Message: res.Message,
		Data:    s.mapSaldoYearTotalBalanceResponses(res.Data),
	}
}

func (s *saldoGraphqlMapper) ToGraphqlResponseMonthSaldoBalances(res *pbStats.ApiResponseMonthSaldoBalances) *model.APIResponseMonthSaldoBalances {
	return &model.APIResponseMonthSaldoBalances{
		Status:  res.Status,
		Message: res.Message,
		Data:    s.mapSaldoMonthBalanceResponses(res.Data),
	}
}

func (s *saldoGraphqlMapper) ToGraphqlResponseYearBalance(res *pbStats.ApiResponseYearSaldoBalances) *model.APIResponseYearSaldoBalances {
	return &model.APIResponseYearSaldoBalances{
		Status:  res.Status,
		Message: res.Message,
		Data:    s.mapSaldoYearBalanceResponses(res.Data),
	}
}

func (s *saldoGraphqlMapper) mapResponseSaldo(saldo *pb.SaldoResponse) *model.SaldoResponse {
	withdrawAmount := int32(saldo.WithdrawAmount)

	return &model.SaldoResponse{
		SaldoID:        int32(saldo.SaldoId),
		CardNumber:     saldo.CardNumber,
		TotalBalance:   int32(saldo.TotalBalance),
		WithdrawTime:   &saldo.WithdrawTime,
		WithdrawAmount: &withdrawAmount,
		CreatedAt:      saldo.CreatedAt,
		UpdatedAt:      saldo.UpdatedAt,
	}
}

func (s *saldoGraphqlMapper) mapResponsesSaldo(saldos []*pb.SaldoResponse) []*model.SaldoResponse {
	var responseSaldos []*model.SaldoResponse

	for _, saldo := range saldos {
		responseSaldos = append(responseSaldos, s.mapResponseSaldo(saldo))
	}

	return responseSaldos
}

func (s *saldoGraphqlMapper) mapResponseSaldoDeleteAt(saldo *pb.SaldoResponseDeleteAt) *model.SaldoResponseDeletedAt {
	withdrawAmount := int32(saldo.WithdrawAmount)

	return &model.SaldoResponseDeletedAt{
		SaldoID:        int32(saldo.SaldoId),
		CardNumber:     saldo.CardNumber,
		TotalBalance:   int32(saldo.TotalBalance),
		WithdrawTime:   &saldo.WithdrawTime,
		WithdrawAmount: &withdrawAmount,
		CreatedAt:      saldo.CreatedAt,
		UpdatedAt:      saldo.UpdatedAt,
	}
}

func (s *saldoGraphqlMapper) mapResponsesSaldoDeleteAt(saldos []*pb.SaldoResponseDeleteAt) []*model.SaldoResponseDeletedAt {
	var responseSaldos []*model.SaldoResponseDeletedAt

	for _, saldo := range saldos {
		responseSaldos = append(responseSaldos, s.mapResponseSaldoDeleteAt(saldo))
	}

	return responseSaldos
}

func (s *saldoGraphqlMapper) mapSaldoMonthTotalBalanceResponse(saldo *pbStats.SaldoMonthTotalBalanceResponse) *model.SaldoMonthTotalBalanceResponse {
	totalBalance := 0

	if saldo.TotalBalance != 0 {
		totalBalance = int(saldo.TotalBalance)
	}

	return &model.SaldoMonthTotalBalanceResponse{
		Month:        saldo.Month,
		Year:         saldo.Year,
		TotalBalance: int32(totalBalance),
	}
}

func (s *saldoGraphqlMapper) mapSaldoMonthTotalBalanceResponses(saldos []*pbStats.SaldoMonthTotalBalanceResponse) []*model.SaldoMonthTotalBalanceResponse {
	var responsesSaldo []*model.SaldoMonthTotalBalanceResponse

	for _, saldo := range saldos {
		responsesSaldo = append(responsesSaldo, s.mapSaldoMonthTotalBalanceResponse(saldo))
	}

	return responsesSaldo
}

func (s *saldoGraphqlMapper) mapSaldoYearTotalBalanceResponse(saldo *pbStats.SaldoYearTotalBalanceResponse) *model.SaldoYearTotalBalanceResponse {
	totalBalance := 0

	if saldo.TotalBalance != 0 {
		totalBalance = int(saldo.TotalBalance)
	}

	return &model.SaldoYearTotalBalanceResponse{
		Year:         saldo.Year,
		TotalBalance: int32(totalBalance),
	}
}

func (s *saldoGraphqlMapper) mapSaldoYearTotalBalanceResponses(saldos []*pbStats.SaldoYearTotalBalanceResponse) []*model.SaldoYearTotalBalanceResponse {
	var responsesSaldo []*model.SaldoYearTotalBalanceResponse

	for _, saldo := range saldos {
		responsesSaldo = append(responsesSaldo, s.mapSaldoYearTotalBalanceResponse(saldo))
	}

	return responsesSaldo
}

func (s *saldoGraphqlMapper) mapSaldoMonthBalanceResponse(saldo *pbStats.SaldoMonthBalanceResponse) *model.SaldoMonthBalanceResponse {
	totalBalance := 0

	if saldo.TotalBalance != 0 {
		totalBalance = int(saldo.TotalBalance)
	}

	return &model.SaldoMonthBalanceResponse{
		Month:        saldo.Month,
		TotalBalance: int32(totalBalance),
	}
}

func (s *saldoGraphqlMapper) mapSaldoMonthBalanceResponses(saldos []*pbStats.SaldoMonthBalanceResponse) []*model.SaldoMonthBalanceResponse {
	var responsesSaldo []*model.SaldoMonthBalanceResponse

	for _, saldo := range saldos {
		responsesSaldo = append(responsesSaldo, s.mapSaldoMonthBalanceResponse(saldo))
	}

	return responsesSaldo
}

func (s *saldoGraphqlMapper) mapSaldoYearBalanceResponse(saldo *pbStats.SaldoYearBalanceResponse) *model.SaldoYearBalanceResponse {
	totalBalance := 0

	if saldo.TotalBalance != 0 {
		totalBalance = int(saldo.TotalBalance)
	}

	return &model.SaldoYearBalanceResponse{
		Year:         saldo.Year,
		TotalBalance: int32(totalBalance),
	}
}

func (s *saldoGraphqlMapper) mapSaldoYearBalanceResponses(saldos []*pbStats.SaldoYearBalanceResponse) []*model.SaldoYearBalanceResponse {
	var responsesSaldo []*model.SaldoYearBalanceResponse

	for _, saldo := range saldos {
		responsesSaldo = append(responsesSaldo, s.mapSaldoYearBalanceResponse(saldo))
	}

	return responsesSaldo
}
