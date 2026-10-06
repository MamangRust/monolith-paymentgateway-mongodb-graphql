package transfer_stats_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
)

type TransferStatsAmountCache interface {
	GetCachedMonthTransferAmounts(ctx context.Context, year int) (*model.APIResponseTransferMonthAmount, bool)
	SetCachedMonthTransferAmounts(ctx context.Context, year int, data *model.APIResponseTransferMonthAmount)

	GetCachedYearlyTransferAmounts(ctx context.Context, year int) (*model.APIResponseTransferYearAmount, bool)
	SetCachedYearlyTransferAmounts(ctx context.Context, year int, data *model.APIResponseTransferYearAmount)
}

type TransferStatsStatusCache interface {
	GetCachedMonthTransferStatusSuccess(ctx context.Context, req *model.FindMonthlyTransferStatusInput) (*model.APIResponseTransferMonthStatusSuccess, bool)
	SetCachedMonthTransferStatusSuccess(ctx context.Context, req *model.FindMonthlyTransferStatusInput, data *model.APIResponseTransferMonthStatusSuccess)

	GetCachedYearlyTransferStatusSuccess(ctx context.Context, year int) (*model.APIResponseTransferYearStatusSuccess, bool)
	SetCachedYearlyTransferStatusSuccess(ctx context.Context, year int, data *model.APIResponseTransferYearStatusSuccess)

	GetCachedMonthTransferStatusFailed(ctx context.Context, req *model.FindMonthlyTransferStatusInput) (*model.APIResponseTransferMonthStatusFailed, bool)
	SetCachedMonthTransferStatusFailed(ctx context.Context, req *model.FindMonthlyTransferStatusInput, data *model.APIResponseTransferMonthStatusFailed)

	GetCachedYearlyTransferStatusFailed(ctx context.Context, year int) (*model.APIResponseTransferYearStatusFailed, bool)
	SetCachedYearlyTransferStatusFailed(ctx context.Context, year int, data *model.APIResponseTransferYearStatusFailed)
}
