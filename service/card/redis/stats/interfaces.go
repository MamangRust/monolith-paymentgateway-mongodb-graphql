package cardstatsmencache

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
)

type CardStatsBalanceCache interface {
	GetMonthlyBalanceCache(ctx context.Context, year int) ([]*salemongo.MonthlyBalanceRow, bool)
	SetMonthlyBalanceCache(ctx context.Context, year int, data []*salemongo.MonthlyBalanceRow)

	GetYearlyBalanceCache(ctx context.Context, year int) ([]*salemongo.YearlyBalanceRow, bool)
	SetYearlyBalanceCache(ctx context.Context, year int, data []*salemongo.YearlyBalanceRow)
}

type CardStatsTopupCache interface {
	GetMonthlyTopupCache(ctx context.Context, year int) ([]*salemongo.MonthlyTopupRow, bool)
	SetMonthlyTopupCache(ctx context.Context, year int, data []*salemongo.MonthlyTopupRow)

	GetYearlyTopupCache(ctx context.Context, year int) ([]*salemongo.YearlyTopupRow, bool)
	SetYearlyTopupCache(ctx context.Context, year int, data []*salemongo.YearlyTopupRow)
}

type CardStatsWithdrawCache interface {
	GetMonthlyWithdrawCache(ctx context.Context, year int) ([]*salemongo.MonthlyWithdrawRow, bool)
	SetMonthlyWithdrawCache(ctx context.Context, year int, data []*salemongo.MonthlyWithdrawRow)

	GetYearlyWithdrawCache(ctx context.Context, year int) ([]*salemongo.YearlyWithdrawRow, bool)
	SetYearlyWithdrawCache(ctx context.Context, year int, data []*salemongo.YearlyWithdrawRow)
}

type CardStatsTransactionCache interface {
	GetMonthlyTransactionCache(ctx context.Context, year int) ([]*salemongo.MonthlyTransactionRow, bool)
	SetMonthlyTransactionCache(ctx context.Context, year int, data []*salemongo.MonthlyTransactionRow)

	GetYearlyTransactionCache(ctx context.Context, year int) ([]*salemongo.YearlyTransactionRow, bool)
	SetYearlyTransactionCache(ctx context.Context, year int, data []*salemongo.YearlyTransactionRow)
}

type CardStatsTransferCache interface {
	GetMonthlyTransferSenderCache(ctx context.Context, year int) ([]*salemongo.MonthlyTransferSentRow, bool)
	SetMonthlyTransferSenderCache(ctx context.Context, year int, data []*salemongo.MonthlyTransferSentRow)

	GetYearlyTransferSenderCache(ctx context.Context, year int) ([]*salemongo.YearlyTransferSentRow, bool)
	SetYearlyTransferSenderCache(ctx context.Context, year int, data []*salemongo.YearlyTransferSentRow)

	GetMonthlyTransferReceiverCache(ctx context.Context, year int) ([]*salemongo.MonthlyTransferReceivedRow, bool)
	SetMonthlyTransferReceiverCache(ctx context.Context, year int, data []*salemongo.MonthlyTransferReceivedRow)

	GetYearlyTransferReceiverCache(ctx context.Context, year int) ([]*salemongo.YearlyTransferReceivedRow, bool)
	SetYearlyTransferReceiverCache(ctx context.Context, year int, data []*salemongo.YearlyTransferReceivedRow)
}
