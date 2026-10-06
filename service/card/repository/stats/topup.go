package repositorystats

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	card_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/card_errors/repository"
	"go.mongodb.org/mongo-driver/mongo"
)

type cardStatsTopupRepository struct {
	db *mongo.Database
}

func NewCardStatsTopupRepository(db *mongo.Database) CardStatsTopupRepository {
	return &cardStatsTopupRepository{db: db}
}

func (r *cardStatsTopupRepository) GetMonthlyTopupAmount(ctx context.Context, year int) ([]*salemongo.MonthlyTopupRow, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	totals, err := sumByMonth(ctx, r.db, "topups", "topup_time", "topup_amount", start, end)
	if err != nil {
		return nil, card_errors.ErrGetMonthlyTopupAmountFailed.WithInternal(err)
	}

	results := make([]*salemongo.MonthlyTopupRow, 0, 12)
	for month := 1; month <= 12; month++ {
		results = append(results, &salemongo.MonthlyTopupRow{
			Month:            monthName(time.Month(month)),
			TotalTopupAmount: int32(totals[month]),
		})
	}
	return results, nil
}

func (r *cardStatsTopupRepository) GetYearlyTopupAmount(ctx context.Context, year int) ([]*salemongo.YearlyTopupRow, error) {
	start := time.Date(year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	totals, err := sumByYear(ctx, r.db, "topups", "topup_time", "topup_amount", start, end)
	if err != nil {
		return nil, card_errors.ErrGetYearlyTopupAmountFailed.WithInternal(err)
	}

	results := make([]*salemongo.YearlyTopupRow, 0, 5)
	for y := year - 4; y <= year; y++ {
		results = append(results, &salemongo.YearlyTopupRow{
			Year:             yearString(y),
			TotalTopupAmount: totals[y],
		})
	}
	return results, nil
}
