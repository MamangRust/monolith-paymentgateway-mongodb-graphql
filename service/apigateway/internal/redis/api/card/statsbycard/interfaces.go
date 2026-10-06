package card_stats_bycard_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
)

type CardStatsBalanceByCardCache interface {
	GetMonthlyBalanceByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput) (*model.APIResponseMonthlyBalance, bool)
	GetYearlyBalanceByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput) (*model.APIResponseYearlyBalance, bool)

	SetMonthlyBalanceByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput, data *model.APIResponseMonthlyBalance)
	SetYearlyBalanceByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput, data *model.APIResponseYearlyBalance)
}

type CardStatsTopupByCardCache interface {
	GetMonthlyTopupByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput) (*model.APIResponseMonthlyAmount, bool)
	GetYearlyTopupByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput) (*model.APIResponseYearlyAmount, bool)

	SetMonthlyTopupByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput, data *model.APIResponseMonthlyAmount)
	SetYearlyTopupByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput, data *model.APIResponseYearlyAmount)
}

type CardStatsWithdrawByCardCache interface {
	GetMonthlyWithdrawByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput) (*model.APIResponseMonthlyAmount, bool)
	GetYearlyWithdrawByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput) (*model.APIResponseYearlyAmount, bool)

	SetMonthlyWithdrawByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput, data *model.APIResponseMonthlyAmount)
	SetYearlyWithdrawByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput, data *model.APIResponseYearlyAmount)
}

type CardStatsTransactionByCardCache interface {
	GetMonthlyTransactionByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput) (*model.APIResponseMonthlyAmount, bool)
	GetYearlyTransactionByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput) (*model.APIResponseYearlyAmount, bool)

	SetMonthlyTransactionByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput, data *model.APIResponseMonthlyAmount)
	SetYearlyTransactionByNumberCache(ctx context.Context, req *model.FindYearCardNumberInput, data *model.APIResponseYearlyAmount)
}

type CardStatsTransferByCardCache interface {
	GetMonthlyTransferBySenderCache(ctx context.Context, req *model.FindYearCardNumberInput) (*model.APIResponseMonthlyAmount, bool)
	GetYearlyTransferBySenderCache(ctx context.Context, req *model.FindYearCardNumberInput) (*model.APIResponseYearlyAmount, bool)

	SetMonthlyTransferBySenderCache(ctx context.Context, req *model.FindYearCardNumberInput, data *model.APIResponseMonthlyAmount)
	SetYearlyTransferBySenderCache(ctx context.Context, req *model.FindYearCardNumberInput, data *model.APIResponseYearlyAmount)

	GetMonthlyTransferByReceiverCache(ctx context.Context, req *model.FindYearCardNumberInput) (*model.APIResponseMonthlyAmount, bool)
	GetYearlyTransferByReceiverCache(ctx context.Context, req *model.FindYearCardNumberInput) (*model.APIResponseYearlyAmount, bool)

	SetMonthlyTransferByReceiverCache(ctx context.Context, req *model.FindYearCardNumberInput, data *model.APIResponseMonthlyAmount)
	SetYearlyTransferByReceiverCache(ctx context.Context, req *model.FindYearCardNumberInput, data *model.APIResponseYearlyAmount)
}
