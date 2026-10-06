package transferstatscache

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type TransferStatsAmountCache interface {
	GetCachedMonthTransferAmounts(ctx context.Context, year int) ([]*salemongo.TransferMonthlyAmountRow, bool)
	SetCachedMonthTransferAmounts(ctx context.Context, year int, data []*salemongo.TransferMonthlyAmountRow)

	GetCachedYearlyTransferAmounts(ctx context.Context, year int) ([]*salemongo.TransferYearlyAmountRow, bool)
	SetCachedYearlyTransferAmounts(ctx context.Context, year int, data []*salemongo.TransferYearlyAmountRow)
}

type TransferStatsStatusCache interface {
	GetCachedMonthTransferStatusSuccess(ctx context.Context, req *requests.MonthStatusTransfer) ([]*salemongo.TransferMonthlyStatusSuccessRow, bool)
	SetCachedMonthTransferStatusSuccess(ctx context.Context, req *requests.MonthStatusTransfer, data []*salemongo.TransferMonthlyStatusSuccessRow)

	GetCachedYearlyTransferStatusSuccess(ctx context.Context, year int) ([]*salemongo.TransferYearlyStatusSuccessRow, bool)
	SetCachedYearlyTransferStatusSuccess(ctx context.Context, year int, data []*salemongo.TransferYearlyStatusSuccessRow)

	GetCachedMonthTransferStatusFailed(ctx context.Context, req *requests.MonthStatusTransfer) ([]*salemongo.TransferMonthlyStatusFailedRow, bool)
	SetCachedMonthTransferStatusFailed(ctx context.Context, req *requests.MonthStatusTransfer, data []*salemongo.TransferMonthlyStatusFailedRow)

	GetCachedYearlyTransferStatusFailed(ctx context.Context, year int) ([]*salemongo.TransferYearlyStatusFailedRow, bool)
	SetCachedYearlyTransferStatusFailed(ctx context.Context, year int, data []*salemongo.TransferYearlyStatusFailedRow)
}
