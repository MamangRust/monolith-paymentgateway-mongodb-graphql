package withdraw_stats_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
)

type WithdrawStatsStatusCache interface {
	GetCachedMonthWithdrawStatusSuccessCache(ctx context.Context, req *model.FindMonthlyWithdrawStatusInput) (*model.APIResponseWithdrawMonthStatusSuccess, bool)
	SetCachedMonthWithdrawStatusSuccessCache(ctx context.Context, req *model.FindMonthlyWithdrawStatusInput, data *model.APIResponseWithdrawMonthStatusSuccess)

	GetCachedYearlyWithdrawStatusSuccessCache(ctx context.Context, year int) (*model.APIResponseWithdrawYearStatusSuccess, bool)
	SetCachedYearlyWithdrawStatusSuccessCache(ctx context.Context, year int, data *model.APIResponseWithdrawYearStatusSuccess)

	GetCachedMonthWithdrawStatusFailedCache(ctx context.Context, req *model.FindMonthlyWithdrawStatusInput) (*model.APIResponseWithdrawMonthStatusFailed, bool)
	SetCachedMonthWithdrawStatusFailedCache(ctx context.Context, req *model.FindMonthlyWithdrawStatusInput, data *model.APIResponseWithdrawMonthStatusFailed)

	GetCachedYearlyWithdrawStatusFailedCache(ctx context.Context, year int) (*model.APIResponseWithdrawYearStatusFailed, bool)
	SetCachedYearlyWithdrawStatusFailedCache(ctx context.Context, year int, data *model.APIResponseWithdrawYearStatusFailed)
}

type WithdrawStatsAmountCache interface {
	GetCachedMonthlyWithdraws(ctx context.Context, year int) (*model.APIResponseWithdrawMonthAmount, bool)
	SetCachedMonthlyWithdraws(ctx context.Context, year int, data *model.APIResponseWithdrawMonthAmount)

	GetCachedYearlyWithdraws(ctx context.Context, year int) (*model.APIResponseWithdrawYearAmount, bool)
	SetCachedYearlyWithdraws(ctx context.Context, year int, data *model.APIResponseWithdrawYearAmount)
}
