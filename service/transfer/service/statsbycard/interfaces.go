package transferstatsbycardservice

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type TransferStatsByCardAmountService interface {
	FindMonthlyTransferAmountsBySenderCardNumber(ctx context.Context, req *requests.MonthYearCardNumber) ([]*salemongo.TransferMonthlyAmountBySenderCardRow, error)
	FindYearlyTransferAmountsBySenderCardNumber(ctx context.Context, req *requests.MonthYearCardNumber) ([]*salemongo.TransferYearlyAmountBySenderCardRow, error)
	FindMonthlyTransferAmountsByReceiverCardNumber(ctx context.Context, req *requests.MonthYearCardNumber) ([]*salemongo.TransferMonthlyAmountByReceiverCardRow, error)
	FindYearlyTransferAmountsByReceiverCardNumber(ctx context.Context, req *requests.MonthYearCardNumber) ([]*salemongo.TransferYearlyAmountByReceiverCardRow, error)
}

type TransferStatsByCardStatusService interface {
	FindMonthTransferStatusSuccessByCardNumber(ctx context.Context, req *requests.MonthStatusTransferCardNumber) ([]*salemongo.TransferMonthlyStatusSuccessByCardRow, error)
	FindYearlyTransferStatusSuccessByCardNumber(ctx context.Context, req *requests.YearStatusTransferCardNumber) ([]*salemongo.TransferYearlyStatusSuccessByCardRow, error)
	FindMonthTransferStatusFailedByCardNumber(ctx context.Context, req *requests.MonthStatusTransferCardNumber) ([]*salemongo.TransferMonthlyStatusFailedByCardRow, error)
	FindYearlyTransferStatusFailedByCardNumber(ctx context.Context, req *requests.YearStatusTransferCardNumber) ([]*salemongo.TransferYearlyStatusFailedByCardRow, error)
}
