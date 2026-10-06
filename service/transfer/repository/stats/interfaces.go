package transferstatsrepository

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type TransferStatsAmountRepository interface {
	GetMonthlyTransferAmounts(ctx context.Context, year int) ([]*salemongo.TransferMonthlyAmountRow, error)
	GetYearlyTransferAmounts(ctx context.Context, year int) ([]*salemongo.TransferYearlyAmountRow, error)
}

type TransferStatsStatusRepository interface {
	GetMonthTransferStatusSuccess(ctx context.Context, req *requests.MonthStatusTransfer) ([]*salemongo.TransferMonthlyStatusSuccessRow, error)
	GetYearlyTransferStatusSuccess(ctx context.Context, year int) ([]*salemongo.TransferYearlyStatusSuccessRow, error)
	GetMonthTransferStatusFailed(ctx context.Context, req *requests.MonthStatusTransfer) ([]*salemongo.TransferMonthlyStatusFailedRow, error)
	GetYearlyTransferStatusFailed(ctx context.Context, year int) ([]*salemongo.TransferYearlyStatusFailedRow, error)
}
