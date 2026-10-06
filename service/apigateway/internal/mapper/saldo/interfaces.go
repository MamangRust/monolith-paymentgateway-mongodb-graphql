package saldographqlmapper

import (
	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/saldo"
	pbStats "github.com/MamangRust/monolith-payment-gateway-pb/saldo/stats"
)

type SaldoGraphqlMapper interface {
	ToGraphqlResponseSaldo(res *pb.ApiResponseSaldo) *model.APIResponseSaldo
	ToGraphqlResponseSaldoDeleteAt(res *pb.ApiResponseSaldoDeleteAt) *model.APIResponseSaldoDeleteAt

	ToGraphqlResponsePaginationSaldo(res *pb.ApiResponsePaginationSaldo) *model.APIResponsePaginationSaldo
	ToGraphqlResponsePaginationSaldoDeleteAt(res *pb.ApiResponsePaginationSaldoDeleteAt) *model.APIResponsePaginationSaldoDeleteAt
	ToGraphqlResponseDelete(res *pb.ApiResponseSaldoDelete) *model.APIResponseSaldoDelete
	ToGraphqlResponseAll(res *pb.ApiResponseSaldoAll) *model.APIResponseSaldoAll
	ToGraphqlResponseMonthTotalSaldo(res *pbStats.ApiResponseMonthTotalSaldo) *model.APIResponseMonthTotalSaldo
	ToGraphqlResponseYearTotalSaldo(res *pbStats.ApiResponseYearTotalSaldo) *model.APIResponseYearTotalSaldo
	ToGraphqlResponseMonthSaldoBalances(res *pbStats.ApiResponseMonthSaldoBalances) *model.APIResponseMonthSaldoBalances
	ToGraphqlResponseYearBalance(res *pbStats.ApiResponseYearSaldoBalances) *model.APIResponseYearSaldoBalances
}
