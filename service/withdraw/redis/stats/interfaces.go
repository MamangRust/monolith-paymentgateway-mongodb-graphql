package withdrawstatscache

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type WithdrawStatsStatusCache interface {
	GetCachedMonthWithdrawStatusSuccessCache(ctx context.Context, req *requests.MonthStatusWithdraw) ([]*salemongo.WithdrawMonthlyStatusSuccessRow, bool)
	SetCachedMonthWithdrawStatusSuccessCache(ctx context.Context, req *requests.MonthStatusWithdraw, data []*salemongo.WithdrawMonthlyStatusSuccessRow)

	GetCachedYearlyWithdrawStatusSuccessCache(ctx context.Context, year int) ([]*salemongo.WithdrawYearlyStatusSuccessRow, bool)
	SetCachedYearlyWithdrawStatusSuccessCache(ctx context.Context, year int, data []*salemongo.WithdrawYearlyStatusSuccessRow)

	GetCachedMonthWithdrawStatusFailedCache(ctx context.Context, req *requests.MonthStatusWithdraw) ([]*salemongo.WithdrawMonthlyStatusFailedRow, bool)
	SetCachedMonthWithdrawStatusFailedCache(ctx context.Context, req *requests.MonthStatusWithdraw, data []*salemongo.WithdrawMonthlyStatusFailedRow)

	GetCachedYearlyWithdrawStatusFailedCache(ctx context.Context, year int) ([]*salemongo.WithdrawYearlyStatusFailedRow, bool)
	SetCachedYearlyWithdrawStatusFailedCache(ctx context.Context, year int, data []*salemongo.WithdrawYearlyStatusFailedRow)
}

type WithdrawStatsAmountCache interface {
	GetCachedMonthlyWithdraws(ctx context.Context, year int) ([]*salemongo.WithdrawMonthlyAmountRow, bool)
	SetCachedMonthlyWithdraws(ctx context.Context, year int, data []*salemongo.WithdrawMonthlyAmountRow)

	GetCachedYearlyWithdraws(ctx context.Context, year int) ([]*salemongo.WithdrawYearlyAmountRow, bool)
	SetCachedYearlyWithdraws(ctx context.Context, year int, data []*salemongo.WithdrawYearlyAmountRow)
}
