package card_stats_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
)

type CardStatsBalanceCache interface {
	GetMonthlyBalanceCache(ctx context.Context, year int) (*model.APIResponseMonthlyBalance, bool)
	SetMonthlyBalanceCache(ctx context.Context, year int, data *model.APIResponseMonthlyBalance)

	GetYearlyBalanceCache(ctx context.Context, year int) (*model.APIResponseYearlyBalance, bool)
	SetYearlyBalanceCache(ctx context.Context, year int, data *model.APIResponseYearlyBalance)
}

type CardStatsTopupCache interface {
	GetMonthlyTopupCache(ctx context.Context, year int) (*model.APIResponseMonthlyAmount, bool)
	SetMonthlyTopupCache(ctx context.Context, year int, data *model.APIResponseMonthlyAmount)

	GetYearlyTopupCache(ctx context.Context, year int) (*model.APIResponseYearlyAmount, bool)
	SetYearlyTopupCache(ctx context.Context, year int, data *model.APIResponseYearlyAmount)
}

type CardStatsWithdrawCache interface {
	GetMonthlyWithdrawCache(ctx context.Context, year int) (*model.APIResponseMonthlyAmount, bool)
	SetMonthlyWithdrawCache(ctx context.Context, year int, data *model.APIResponseMonthlyAmount)

	GetYearlyWithdrawCache(ctx context.Context, year int) (*model.APIResponseYearlyAmount, bool)
	SetYearlyWithdrawCache(ctx context.Context, year int, data *model.APIResponseYearlyAmount)
}

type CardStatsTransactionCache interface {
	GetMonthlyTransactionCache(ctx context.Context, year int) (*model.APIResponseMonthlyAmount, bool)
	SetMonthlyTransactionCache(ctx context.Context, year int, data *model.APIResponseMonthlyAmount)

	GetYearlyTransactionCache(ctx context.Context, year int) (*model.APIResponseYearlyAmount, bool)
	SetYearlyTransactionCache(ctx context.Context, year int, data *model.APIResponseYearlyAmount)
}

type CardStatsTransferCache interface {
	GetMonthlyTransferSenderCache(ctx context.Context, year int) (*model.APIResponseMonthlyAmount, bool)
	SetMonthlyTransferSenderCache(ctx context.Context, year int, data *model.APIResponseMonthlyAmount)

	GetYearlyTransferSenderCache(ctx context.Context, year int) (*model.APIResponseYearlyAmount, bool)
	SetYearlyTransferSenderCache(ctx context.Context, year int, data *model.APIResponseYearlyAmount)

	GetMonthlyTransferReceiverCache(ctx context.Context, year int) (*model.APIResponseMonthlyAmount, bool)
	SetMonthlyTransferReceiverCache(ctx context.Context, year int, data *model.APIResponseMonthlyAmount)

	GetYearlyTransferReceiverCache(ctx context.Context, year int) (*model.APIResponseYearlyAmount, bool)
	SetYearlyTransferReceiverCache(ctx context.Context, year int, data *model.APIResponseYearlyAmount)
}
