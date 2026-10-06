package cardstatsbycardmencache

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type CardStatsBalanceByCardCache interface {
	GetMonthlyBalanceByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyBalanceRow, bool)
	GetYearlyBalanceByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyBalanceRow, bool)

	SetMonthlyBalanceByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard, data []*salemongo.MonthlyBalanceRow)
	SetYearlyBalanceByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard, data []*salemongo.YearlyBalanceRow)
}

type CardStatsTopupByCardCache interface {
	GetMonthlyTopupByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyTopupRow, bool)
	GetYearlyTopupByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyTopupRow, bool)

	SetMonthlyTopupByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard, data []*salemongo.MonthlyTopupRow)
	SetYearlyTopupByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard, data []*salemongo.YearlyTopupRow)
}

type CardStatsWithdrawByCardCache interface {
	GetMonthlyWithdrawByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyWithdrawRow, bool)
	GetYearlyWithdrawByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyWithdrawRow, bool)

	SetMonthlyWithdrawByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard, data []*salemongo.MonthlyWithdrawRow)
	SetYearlyWithdrawByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard, data []*salemongo.YearlyWithdrawRow)
}

type CardStatsTransactionByCardCache interface {
	GetMonthlyTransactionByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyTransactionRow, bool)
	GetYearlyTransactionByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyTransactionRow, bool)

	SetMonthlyTransactionByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard, data []*salemongo.MonthlyTransactionRow)
	SetYearlyTransactionByNumberCache(ctx context.Context, req *requests.MonthYearCardNumberCard, data []*salemongo.YearlyTransactionRow)
}

type CardStatsTransferByCardCache interface {
	GetMonthlyTransferBySenderCache(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyTransferSentRow, bool)
	GetYearlyTransferBySenderCache(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyTransferSentRow, bool)

	SetMonthlyTransferBySenderCache(ctx context.Context, req *requests.MonthYearCardNumberCard, data []*salemongo.MonthlyTransferSentRow)
	SetYearlyTransferBySenderCache(ctx context.Context, req *requests.MonthYearCardNumberCard, data []*salemongo.YearlyTransferSentRow)

	GetMonthlyTransferByReceiverCache(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyTransferReceivedRow, bool)
	GetYearlyTransferByReceiverCache(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyTransferReceivedRow, bool)

	SetMonthlyTransferByReceiverCache(ctx context.Context, req *requests.MonthYearCardNumberCard, data []*salemongo.MonthlyTransferReceivedRow)
	SetYearlyTransferByReceiverCache(ctx context.Context, req *requests.MonthYearCardNumberCard, data []*salemongo.YearlyTransferReceivedRow)
}
