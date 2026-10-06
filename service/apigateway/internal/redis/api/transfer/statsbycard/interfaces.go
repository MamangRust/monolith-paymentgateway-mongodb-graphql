package transfer_stats_bycard_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
)

type TransferStatsByCardAmountCache interface {
	GetMonthlyTransferAmountsBySenderCard(ctx context.Context, req *model.FindByCardNumberTransferInput) (*model.APIResponseTransferMonthAmount, bool)
	SetMonthlyTransferAmountsBySenderCard(ctx context.Context, req *model.FindByCardNumberTransferInput, data *model.APIResponseTransferMonthAmount)

	GetMonthlyTransferAmountsByReceiverCard(ctx context.Context, req *model.FindByCardNumberTransferInput) (*model.APIResponseTransferMonthAmount, bool)
	SetMonthlyTransferAmountsByReceiverCard(ctx context.Context, req *model.FindByCardNumberTransferInput, data *model.APIResponseTransferMonthAmount)

	GetYearlyTransferAmountsBySenderCard(ctx context.Context, req *model.FindByCardNumberTransferInput) (*model.APIResponseTransferYearAmount, bool)
	SetYearlyTransferAmountsBySenderCard(ctx context.Context, req *model.FindByCardNumberTransferInput, data *model.APIResponseTransferYearAmount)

	GetYearlyTransferAmountsByReceiverCard(ctx context.Context, req *model.FindByCardNumberTransferInput) (*model.APIResponseTransferYearAmount, bool)
	SetYearlyTransferAmountsByReceiverCard(ctx context.Context, req *model.FindByCardNumberTransferInput, data *model.APIResponseTransferYearAmount)
}

type TransferStatsByCardStatusCache interface {
	GetMonthTransferStatusSuccessByCard(ctx context.Context, req *model.FindMonthlyTransferStatusCardNumberInput) (*model.APIResponseTransferMonthStatusSuccess, bool)
	SetMonthTransferStatusSuccessByCard(ctx context.Context, req *model.FindMonthlyTransferStatusCardNumberInput, data *model.APIResponseTransferMonthStatusSuccess)

	GetYearlyTransferStatusSuccessByCard(ctx context.Context, req *model.FindYearTransferStatusCardNumberInput) (*model.APIResponseTransferYearStatusSuccess, bool)
	SetYearlyTransferStatusSuccessByCard(ctx context.Context, req *model.FindYearTransferStatusCardNumberInput, data *model.APIResponseTransferYearStatusSuccess)

	GetMonthTransferStatusFailedByCard(ctx context.Context, req *model.FindMonthlyTransferStatusCardNumberInput) (*model.APIResponseTransferMonthStatusFailed, bool)
	SetMonthTransferStatusFailedByCard(ctx context.Context, req *model.FindMonthlyTransferStatusCardNumberInput, data *model.APIResponseTransferMonthStatusFailed)

	GetYearlyTransferStatusFailedByCard(ctx context.Context, req *model.FindYearTransferStatusCardNumberInput) (*model.APIResponseTransferYearStatusFailed, bool)
	SetYearlyTransferStatusFailedByCard(ctx context.Context, req *model.FindYearTransferStatusCardNumberInput, data *model.APIResponseTransferYearStatusFailed)
}
