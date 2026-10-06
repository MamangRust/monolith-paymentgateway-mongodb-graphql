package repositorystats

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	card_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/card_errors/repository"
	"go.mongodb.org/mongo-driver/mongo"
)

type cardStatsTransactionRepository struct {
	db *mongo.Database
}

func NewCardStatsTransactionRepository(db *mongo.Database) CardStatsTransactionRepository {
	return &cardStatsTransactionRepository{db: db}
}

func (r *cardStatsTransactionRepository) GetMonthlyTransactionAmount(ctx context.Context, year int) ([]*salemongo.MonthlyTransactionRow, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	totals, err := sumByMonth(ctx, r.db, "transactions", "transaction_time", "amount", start, end)
	if err != nil {
		return nil, card_errors.ErrGetMonthlyTransactionAmountFailed.WithInternal(err)
	}

	results := make([]*salemongo.MonthlyTransactionRow, 0, 12)
	for month := 1; month <= 12; month++ {
		results = append(results, &salemongo.MonthlyTransactionRow{
			Month:                  monthName(time.Month(month)),
			TotalTransactionAmount: int32(totals[month]),
		})
	}
	return results, nil
}

func (r *cardStatsTransactionRepository) GetYearlyTransactionAmount(ctx context.Context, year int) ([]*salemongo.YearlyTransactionRow, error) {
	start := time.Date(year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	totals, err := sumByYear(ctx, r.db, "transactions", "transaction_time", "amount", start, end)
	if err != nil {
		return nil, card_errors.ErrGetYearlyTransactionAmountFailed.WithInternal(err)
	}

	results := make([]*salemongo.YearlyTransactionRow, 0, 5)
	for y := year - 4; y <= year; y++ {
		results = append(results, &salemongo.YearlyTransactionRow{
			Year:                   yearString(y),
			TotalTransactionAmount: totals[y],
		})
	}
	return results, nil
}
