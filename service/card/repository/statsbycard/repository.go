package repositorystatsbycard

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"go.mongodb.org/mongo-driver/mongo"
)

type CardStatsBalanceByCardRepository interface {
	GetMonthlyBalancesByCardNumber(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyBalanceRow, error)
	GetYearlyBalanceByCardNumber(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyBalanceRow, error)
}

type CardStatsTopupByCardRepository interface {
	GetMonthlyTopupAmountByCardNumber(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyTopupRow, error)
	GetYearlyTopupAmountByCardNumber(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyTopupRow, error)
}

type CardStatsWithdrawByCardRepository interface {
	GetMonthlyWithdrawAmountByCardNumber(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyWithdrawRow, error)
	GetYearlyWithdrawAmountByCardNumber(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyWithdrawRow, error)
}

type CardStatsTransactionByCardRepository interface {
	GetMonthlyTransactionAmountByCardNumber(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyTransactionRow, error)
	GetYearlyTransactionAmountByCardNumber(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyTransactionRow, error)
}

type CardStatsTransferByCardRepository interface {
	GetMonthlyTransferAmountBySender(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyTransferSentRow, error)
	GetYearlyTransferAmountBySender(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyTransferSentRow, error)
	GetMonthlyTransferAmountByReceiver(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyTransferReceivedRow, error)
	GetYearlyTransferAmountByReceiver(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyTransferReceivedRow, error)
}

type CardStatsByCardRepository interface {
	CardStatsBalanceByCardRepository
	CardStatsTopupByCardRepository
	CardStatsTransactionByCardRepository
	CardStatsTransferByCardRepository
	CardStatsWithdrawByCardRepository
}

type repository struct {
	CardStatsBalanceByCardRepository
	CardStatsTopupByCardRepository
	CardStatsTransactionByCardRepository
	CardStatsTransferByCardRepository
	CardStatsWithdrawByCardRepository
}

func NewCardStatsByCardRepository(db *mongo.Database) CardStatsByCardRepository {
	return &repository{
		CardStatsBalanceByCardRepository:     NewCardStatsBalanceByCardRepository(db),
		CardStatsTopupByCardRepository:       NewCardStatsTopupByCardRepository(db),
		CardStatsTransactionByCardRepository: NewCardStatsTransactionByCardRepository(db),
		CardStatsTransferByCardRepository:    NewCardStatsTransferByCardRepository(db),
		CardStatsWithdrawByCardRepository:    NewCardStatsWithdrawByCardRepository(db),
	}
}
