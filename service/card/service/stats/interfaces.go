package cardstatsservice

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
)

// CardStatsBalanceService handles balance statistics globally and per specific card number.
type CardStatsBalanceService interface {
	FindMonthlyBalance(ctx context.Context, year int) ([]*salemongo.MonthlyBalanceRow, error)
	FindYearlyBalance(ctx context.Context, year int) ([]*salemongo.YearlyBalanceRow, error)
}

// CardStatsTopupService handles top-up statistics globally and per specific card number.
type CardStatsTopupService interface {
	FindMonthlyTopupAmount(ctx context.Context, year int) ([]*salemongo.MonthlyTopupRow, error)
	FindYearlyTopupAmount(ctx context.Context, year int) ([]*salemongo.YearlyTopupRow, error)
}

// CardStatsWithdrawService handles withdraw statistics globally and per specific card number.
type CardStatsWithdrawService interface {
	FindMonthlyWithdrawAmount(ctx context.Context, year int) ([]*salemongo.MonthlyWithdrawRow, error)
	FindYearlyWithdrawAmount(ctx context.Context, year int) ([]*salemongo.YearlyWithdrawRow, error)
}

// CardStatsTransactionService handles transaction statistics globally and per specific card number.
type CardStatsTransactionService interface {
	FindMonthlyTransactionAmount(ctx context.Context, year int) ([]*salemongo.MonthlyTransactionRow, error)
	FindYearlyTransactionAmount(ctx context.Context, year int) ([]*salemongo.YearlyTransactionRow, error)
}

// CardStatsTransferService handles transfer statistics globally and per specific card number (as sender or receiver).
type CardStatsTransferService interface {
	FindMonthlyTransferAmountSender(ctx context.Context, year int) ([]*salemongo.MonthlyTransferSentRow, error)
	FindYearlyTransferAmountSender(ctx context.Context, year int) ([]*salemongo.YearlyTransferSentRow, error)
	FindMonthlyTransferAmountReceiver(ctx context.Context, year int) ([]*salemongo.MonthlyTransferReceivedRow, error)
	FindYearlyTransferAmountReceiver(ctx context.Context, year int) ([]*salemongo.YearlyTransferReceivedRow, error)
}
