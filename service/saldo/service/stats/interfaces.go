package saldostatsservice

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type SaldoStatsTotalBalanceService interface {
	FindMonthlyTotalSaldoBalance(ctx context.Context, req *requests.MonthTotalSaldoBalance) ([]*salemongo.MonthlyTotalSaldoBalanceRow, error)
	FindYearTotalSaldoBalance(ctx context.Context, year int) ([]*salemongo.YearlyTotalSaldoBalancesRow, error)
}

type SaldoStatsBalanceService interface {
	FindMonthlySaldoBalances(ctx context.Context, year int) ([]*salemongo.MonthlySaldoBalanceRow, error)
	FindYearlySaldoBalances(ctx context.Context, year int) ([]*salemongo.YearlySaldoBalanceRow, error)
}
