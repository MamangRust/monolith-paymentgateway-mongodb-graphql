package withdraw_stats_bycard_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
)

type WithdrawStatsByCardStatusCache interface {
	GetCachedMonthWithdrawStatusSuccessByCardNumber(ctx context.Context, req *model.FindMonthlyWithdrawStatusCardNumberInput) (*model.APIResponseWithdrawMonthStatusSuccess, bool)
	SetCachedMonthWithdrawStatusSuccessByCardNumber(ctx context.Context, req *model.FindMonthlyWithdrawStatusCardNumberInput, data *model.APIResponseWithdrawMonthStatusSuccess)

	GetCachedYearlyWithdrawStatusSuccessByCardNumber(ctx context.Context, req *model.FindYearWithdrawStatusCardNumberInput) (*model.APIResponseWithdrawYearStatusSuccess, bool)
	SetCachedYearlyWithdrawStatusSuccessByCardNumber(ctx context.Context, req *model.FindYearWithdrawStatusCardNumberInput, data *model.APIResponseWithdrawYearStatusSuccess)

	GetCachedMonthWithdrawStatusFailedByCardNumber(ctx context.Context, req *model.FindMonthlyWithdrawStatusCardNumberInput) (*model.APIResponseWithdrawMonthStatusFailed, bool)
	SetCachedMonthWithdrawStatusFailedByCardNumber(ctx context.Context, req *model.FindMonthlyWithdrawStatusCardNumberInput, data *model.APIResponseWithdrawMonthStatusFailed)

	GetCachedYearlyWithdrawStatusFailedByCardNumber(ctx context.Context, req *model.FindYearWithdrawStatusCardNumberInput) (*model.APIResponseWithdrawYearStatusFailed, bool)
	SetCachedYearlyWithdrawStatusFailedByCardNumber(ctx context.Context, req *model.FindYearWithdrawStatusCardNumberInput, data *model.APIResponseWithdrawYearStatusFailed)
}

type WithdrawStatsByCardAmountCache interface {
	GetCachedMonthlyWithdrawsByCardNumber(ctx context.Context, req *model.FindYearWithdrawCardNumberInput) (*model.APIResponseWithdrawMonthAmount, bool)
	SetCachedMonthlyWithdrawsByCardNumber(ctx context.Context, req *model.FindYearWithdrawCardNumberInput, data *model.APIResponseWithdrawMonthAmount)

	GetCachedYearlyWithdrawsByCardNumber(ctx context.Context, req *model.FindYearWithdrawCardNumberInput) (*model.APIResponseWithdrawYearAmount, bool)
	SetCachedYearlyWithdrawsByCardNumber(ctx context.Context, req *model.FindYearWithdrawCardNumberInput, data *model.APIResponseWithdrawYearAmount)
}
