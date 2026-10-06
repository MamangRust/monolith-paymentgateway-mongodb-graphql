package cardstatsbycard

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

// CardStatsBalanceByCardService provides methods for retrieving card balance statistics by card number.
type CardStatsBalanceByCardService interface {
	FindMonthlyBalancesByCardNumber(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyBalanceRow, error)
	FindYearlyBalanceByCardNumber(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyBalanceRow, error)
}

type CardStatsTopupByCardService interface {
	FindMonthlyTopupAmountByCardNumber(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyTopupRow, error)
	FindYearlyTopupAmountByCardNumber(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyTopupRow, error)
}

type CardStatsWithdrawByCardService interface {
	FindMonthlyWithdrawAmountByCardNumber(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyWithdrawRow, error)
	FindYearlyWithdrawAmountByCardNumber(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyWithdrawRow, error)
}

type CardStatsTransactionByCardService interface {
	FindMonthlyTransactionAmountByCardNumber(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyTransactionRow, error)
	FindYearlyTransactionAmountByCardNumber(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyTransactionRow, error)
}

type CardStatsTransferByCardService interface {
	FindMonthlyTransferAmountBySender(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyTransferSentRow, error)
	FindYearlyTransferAmountBySender(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyTransferSentRow, error)
	FindMonthlyTransferAmountByReceiver(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyTransferReceivedRow, error)
	FindYearlyTransferAmountByReceiver(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyTransferReceivedRow, error)
}
