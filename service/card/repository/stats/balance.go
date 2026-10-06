package repositorystats

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	card_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/card_errors/repository"
	"go.mongodb.org/mongo-driver/mongo"
)

type CardStatsBalanceRepository interface {
	GetMonthlyBalance(ctx context.Context, year int) ([]*salemongo.MonthlyBalanceRow, error)
	GetYearlyBalance(ctx context.Context, year int) ([]*salemongo.YearlyBalanceRow, error)
}

type CardStatsTopupRepository interface {
	GetMonthlyTopupAmount(ctx context.Context, year int) ([]*salemongo.MonthlyTopupRow, error)
	GetYearlyTopupAmount(ctx context.Context, year int) ([]*salemongo.YearlyTopupRow, error)
}

type CardStatsWithdrawRepository interface {
	GetMonthlyWithdrawAmount(ctx context.Context, year int) ([]*salemongo.MonthlyWithdrawRow, error)
	GetYearlyWithdrawAmount(ctx context.Context, year int) ([]*salemongo.YearlyWithdrawRow, error)
}

type CardStatsTransactionRepository interface {
	GetMonthlyTransactionAmount(ctx context.Context, year int) ([]*salemongo.MonthlyTransactionRow, error)
	GetYearlyTransactionAmount(ctx context.Context, year int) ([]*salemongo.YearlyTransactionRow, error)
}

type CardStatsTransferRepository interface {
	GetMonthlyTransferAmountSender(ctx context.Context, year int) ([]*salemongo.MonthlyTransferSentRow, error)
	GetYearlyTransferAmountSender(ctx context.Context, year int) ([]*salemongo.YearlyTransferSentRow, error)
	GetMonthlyTransferAmountReceiver(ctx context.Context, year int) ([]*salemongo.MonthlyTransferReceivedRow, error)
	GetYearlyTransferAmountReceiver(ctx context.Context, year int) ([]*salemongo.YearlyTransferReceivedRow, error)
}

type cardStatsBalanceRepository struct {
	db *mongo.Database
}

func NewCardStatsBalanceRepository(db *mongo.Database) CardStatsBalanceRepository {
	return &cardStatsBalanceRepository{db: db}
}

func (r *cardStatsBalanceRepository) GetMonthlyBalance(ctx context.Context, year int) ([]*salemongo.MonthlyBalanceRow, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	totals, err := sumByMonth(ctx, r.db, "saldos", "created_at", "total_balance", start, end)
	if err != nil {
		return nil, card_errors.ErrGetMonthlyBalanceFailed.WithInternal(err)
	}

	results := make([]*salemongo.MonthlyBalanceRow, 0, 12)
	for month := 1; month <= 12; month++ {
		results = append(results, &salemongo.MonthlyBalanceRow{
			Month:        monthName(time.Month(month)),
			TotalBalance: int32(totals[month]),
		})
	}
	return results, nil
}

func (r *cardStatsBalanceRepository) GetYearlyBalance(ctx context.Context, year int) ([]*salemongo.YearlyBalanceRow, error) {
	start := time.Date(year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	totals, err := sumByYear(ctx, r.db, "saldos", "created_at", "total_balance", start, end)
	if err != nil {
		return nil, card_errors.ErrGetYearlyBalanceFailed.WithInternal(err)
	}

	results := make([]*salemongo.YearlyBalanceRow, 0, 5)
	for y := year - 4; y <= year; y++ {
		results = append(results, &salemongo.YearlyBalanceRow{
			Year:         yearString(y),
			TotalBalance: totals[y],
		})
	}
	return results, nil
}
