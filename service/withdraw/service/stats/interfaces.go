package withdrawstatsservice

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type WithdrawStatsStatusService interface {
	FindMonthWithdrawStatusSuccess(ctx context.Context, req *requests.MonthStatusWithdraw) ([]*salemongo.WithdrawMonthlyStatusSuccessRow, error)
	FindYearlyWithdrawStatusSuccess(ctx context.Context, year int) ([]*salemongo.WithdrawYearlyStatusSuccessRow, error)
	FindMonthWithdrawStatusFailed(ctx context.Context, req *requests.MonthStatusWithdraw) ([]*salemongo.WithdrawMonthlyStatusFailedRow, error)
	FindYearlyWithdrawStatusFailed(ctx context.Context, year int) ([]*salemongo.WithdrawYearlyStatusFailedRow, error)
}

type WithdrawStatsAmountService interface {
	FindMonthlyWithdraws(ctx context.Context, year int) ([]*salemongo.WithdrawMonthlyAmountRow, error)
	FindYearlyWithdraws(ctx context.Context, year int) ([]*salemongo.WithdrawYearlyAmountRow, error)
}
