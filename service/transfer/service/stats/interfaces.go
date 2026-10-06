package transferstatsservice

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type TransferStatsStatusService interface {
	FindMonthTransferStatusSuccess(ctx context.Context, req *requests.MonthStatusTransfer) ([]*salemongo.TransferMonthlyStatusSuccessRow, error)
	FindYearlyTransferStatusSuccess(ctx context.Context, year int) ([]*salemongo.TransferYearlyStatusSuccessRow, error)
	FindMonthTransferStatusFailed(ctx context.Context, req *requests.MonthStatusTransfer) ([]*salemongo.TransferMonthlyStatusFailedRow, error)
	FindYearlyTransferStatusFailed(ctx context.Context, year int) ([]*salemongo.TransferYearlyStatusFailedRow, error)
}

type TransferStatsAmountService interface {
	FindMonthlyTransferAmounts(ctx context.Context, year int) ([]*salemongo.TransferMonthlyAmountRow, error)
	FindYearlyTransferAmounts(ctx context.Context, year int) ([]*salemongo.TransferYearlyAmountRow, error)
}
