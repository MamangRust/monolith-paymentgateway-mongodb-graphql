package saldostatsrepository

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type SaldoStatsBalanceRepository interface {
	GetMonthlySaldoBalances(ctx context.Context, year int) ([]*salemongo.MonthlySaldoBalanceRow, error)
	GetYearlySaldoBalances(ctx context.Context, year int) ([]*salemongo.YearlySaldoBalanceRow, error)
}

type SaldoStatsTotalSaldoRepository interface {
	GetMonthlyTotalSaldoBalance(ctx context.Context, req *requests.MonthTotalSaldoBalance) ([]*salemongo.MonthlyTotalSaldoBalanceRow, error)
	GetYearTotalSaldoBalance(ctx context.Context, year int) ([]*salemongo.YearlyTotalSaldoBalancesRow, error)
}
