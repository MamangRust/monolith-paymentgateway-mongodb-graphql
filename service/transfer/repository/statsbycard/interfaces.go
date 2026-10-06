package transferstatsbycardrepository

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type TransferStatsByCardAmountSenderRepository interface {
	GetMonthlyTransferAmountsBySenderCardNumber(ctx context.Context, req *requests.MonthYearCardNumber) ([]*salemongo.TransferMonthlyAmountBySenderCardRow, error)
	GetYearlyTransferAmountsBySenderCardNumber(ctx context.Context, req *requests.MonthYearCardNumber) ([]*salemongo.TransferYearlyAmountBySenderCardRow, error)
}

type TransferStatsByCardAmountReceiverRepository interface {
	GetMonthlyTransferAmountsByReceiverCardNumber(ctx context.Context, req *requests.MonthYearCardNumber) ([]*salemongo.TransferMonthlyAmountByReceiverCardRow, error)
	GetYearlyTransferAmountsByReceiverCardNumber(ctx context.Context, req *requests.MonthYearCardNumber) ([]*salemongo.TransferYearlyAmountByReceiverCardRow, error)
}

type TransferStatsByCardStatusRepository interface {
	GetMonthTransferStatusSuccessByCardNumber(ctx context.Context, req *requests.MonthStatusTransferCardNumber) ([]*salemongo.TransferMonthlyStatusSuccessByCardRow, error)
	GetYearlyTransferStatusSuccessByCardNumber(ctx context.Context, req *requests.YearStatusTransferCardNumber) ([]*salemongo.TransferYearlyStatusSuccessByCardRow, error)
	GetMonthTransferStatusFailedByCardNumber(ctx context.Context, req *requests.MonthStatusTransferCardNumber) ([]*salemongo.TransferMonthlyStatusFailedByCardRow, error)
	GetYearlyTransferStatusFailedByCardNumber(ctx context.Context, req *requests.YearStatusTransferCardNumber) ([]*salemongo.TransferYearlyStatusFailedByCardRow, error)
}
