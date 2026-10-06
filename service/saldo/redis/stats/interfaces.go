package saldostatscache

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type SaldoStatsTotalCache interface {
	GetMonthlyTotalSaldoBalanceCache(ctx context.Context, req *requests.MonthTotalSaldoBalance) ([]*salemongo.MonthlyTotalSaldoBalanceRow, bool)
	SetMonthlyTotalSaldoCache(ctx context.Context, req *requests.MonthTotalSaldoBalance, data []*salemongo.MonthlyTotalSaldoBalanceRow)

	GetYearTotalSaldoBalanceCache(ctx context.Context, year int) ([]*salemongo.YearlyTotalSaldoBalancesRow, bool)
	SetYearTotalSaldoBalanceCache(ctx context.Context, year int, data []*salemongo.YearlyTotalSaldoBalancesRow)
}

type SaldoStatsBalanceCache interface {
	GetMonthlySaldoBalanceCache(ctx context.Context, year int) ([]*salemongo.MonthlySaldoBalanceRow, bool)
	SetMonthlySaldoBalanceCache(ctx context.Context, year int, data []*salemongo.MonthlySaldoBalanceRow)

	GetYearlySaldoBalanceCache(ctx context.Context, year int) ([]*salemongo.YearlySaldoBalanceRow, bool)
	SetYearlySaldoBalanceCache(ctx context.Context, year int, data []*salemongo.YearlySaldoBalanceRow)
}
