package saldo_stats_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
)

type SaldoStatsTotalCache interface {
	GetMonthlyTotalSaldoBalanceCache(ctx context.Context, req *model.FindMonthlySaldoTotalBalanceInput) (*model.APIResponseMonthTotalSaldo, bool)
	SetMonthlyTotalSaldoCache(ctx context.Context, req *model.FindMonthlySaldoTotalBalanceInput, data *model.APIResponseMonthTotalSaldo)

	GetYearTotalSaldoBalanceCache(ctx context.Context, year int) (*model.APIResponseYearTotalSaldo, bool)
	SetYearTotalSaldoBalanceCache(ctx context.Context, year int, data *model.APIResponseYearTotalSaldo)
}

type SaldoStatsBalanceCache interface {
	GetMonthlySaldoBalanceCache(ctx context.Context, year int) (*model.APIResponseMonthSaldoBalances, bool)
	SetMonthlySaldoBalanceCache(ctx context.Context, year int, data *model.APIResponseMonthSaldoBalances)

	GetYearlySaldoBalanceCache(ctx context.Context, year int) (*model.APIResponseYearSaldoBalances, bool)
	SetYearlySaldoBalanceCache(ctx context.Context, year int, data *model.APIResponseYearSaldoBalances)
}
