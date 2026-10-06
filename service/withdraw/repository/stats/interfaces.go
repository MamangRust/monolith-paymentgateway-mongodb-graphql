package withdrawstatsrepository

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type WithdrawStatsStatusRepository interface {
	GetMonthWithdrawStatusSuccess(ctx context.Context, req *requests.MonthStatusWithdraw) ([]*salemongo.WithdrawMonthlyStatusSuccessRow, error)
	GetYearlyWithdrawStatusSuccess(ctx context.Context, year int) ([]*salemongo.WithdrawYearlyStatusSuccessRow, error)
	GetMonthWithdrawStatusFailed(ctx context.Context, req *requests.MonthStatusWithdraw) ([]*salemongo.WithdrawMonthlyStatusFailedRow, error)
	GetYearlyWithdrawStatusFailed(ctx context.Context, year int) ([]*salemongo.WithdrawYearlyStatusFailedRow, error)
}

type WithdrawStatsAmountRepository interface {
	GetMonthlyWithdraws(ctx context.Context, year int) ([]*salemongo.WithdrawMonthlyAmountRow, error)
	GetYearlyWithdraws(ctx context.Context, year int) ([]*salemongo.WithdrawYearlyAmountRow, error)
}
