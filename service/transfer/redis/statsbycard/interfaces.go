package transferstatsbycardcache

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type TransferStatsByCardAmountCache interface {
	GetMonthlyTransferAmountsBySenderCard(ctx context.Context, req *requests.MonthYearCardNumber) ([]*salemongo.TransferMonthlyAmountBySenderCardRow, bool)
	SetMonthlyTransferAmountsBySenderCard(ctx context.Context, req *requests.MonthYearCardNumber, data []*salemongo.TransferMonthlyAmountBySenderCardRow)

	GetMonthlyTransferAmountsByReceiverCard(ctx context.Context, req *requests.MonthYearCardNumber) ([]*salemongo.TransferMonthlyAmountByReceiverCardRow, bool)
	SetMonthlyTransferAmountsByReceiverCard(ctx context.Context, req *requests.MonthYearCardNumber, data []*salemongo.TransferMonthlyAmountByReceiverCardRow)

	GetYearlyTransferAmountsBySenderCard(ctx context.Context, req *requests.MonthYearCardNumber) ([]*salemongo.TransferYearlyAmountBySenderCardRow, bool)
	SetYearlyTransferAmountsBySenderCard(ctx context.Context, req *requests.MonthYearCardNumber, data []*salemongo.TransferYearlyAmountBySenderCardRow)

	GetYearlyTransferAmountsByReceiverCard(ctx context.Context, req *requests.MonthYearCardNumber) ([]*salemongo.TransferYearlyAmountByReceiverCardRow, bool)
	SetYearlyTransferAmountsByReceiverCard(ctx context.Context, req *requests.MonthYearCardNumber, data []*salemongo.TransferYearlyAmountByReceiverCardRow)
}

type TransferStatsByCardStatusCache interface {
	GetMonthTransferStatusSuccessByCard(ctx context.Context, req *requests.MonthStatusTransferCardNumber) ([]*salemongo.TransferMonthlyStatusSuccessByCardRow, bool)
	SetMonthTransferStatusSuccessByCard(ctx context.Context, req *requests.MonthStatusTransferCardNumber, data []*salemongo.TransferMonthlyStatusSuccessByCardRow)

	GetYearlyTransferStatusSuccessByCard(ctx context.Context, req *requests.YearStatusTransferCardNumber) ([]*salemongo.TransferYearlyStatusSuccessByCardRow, bool)
	SetYearlyTransferStatusSuccessByCard(ctx context.Context, req *requests.YearStatusTransferCardNumber, data []*salemongo.TransferYearlyStatusSuccessByCardRow)

	GetMonthTransferStatusFailedByCard(ctx context.Context, req *requests.MonthStatusTransferCardNumber) ([]*salemongo.TransferMonthlyStatusFailedByCardRow, bool)
	SetMonthTransferStatusFailedByCard(ctx context.Context, req *requests.MonthStatusTransferCardNumber, data []*salemongo.TransferMonthlyStatusFailedByCardRow)

	GetYearlyTransferStatusFailedByCard(ctx context.Context, req *requests.YearStatusTransferCardNumber) ([]*salemongo.TransferYearlyStatusFailedByCardRow, bool)
	SetYearlyTransferStatusFailedByCard(ctx context.Context, req *requests.YearStatusTransferCardNumber, data []*salemongo.TransferYearlyStatusFailedByCardRow)
}
