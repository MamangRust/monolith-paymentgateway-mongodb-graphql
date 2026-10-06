package repositorystats

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	card_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/card_errors/repository"
	"go.mongodb.org/mongo-driver/mongo"
)

type cardStatsWithdrawRepository struct {
	db *mongo.Database
}

func NewCardStatsWithdrawRepository(db *mongo.Database) CardStatsWithdrawRepository {
	return &cardStatsWithdrawRepository{db: db}
}

func (r *cardStatsWithdrawRepository) GetMonthlyWithdrawAmount(ctx context.Context, year int) ([]*salemongo.MonthlyWithdrawRow, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	totals, err := sumByMonth(ctx, r.db, "withdraws", "withdraw_time", "withdraw_amount", start, end)
	if err != nil {
		return nil, card_errors.ErrGetMonthlyWithdrawAmountFailed.WithInternal(err)
	}

	results := make([]*salemongo.MonthlyWithdrawRow, 0, 12)
	for month := 1; month <= 12; month++ {
		results = append(results, &salemongo.MonthlyWithdrawRow{
			Month:               monthName(time.Month(month)),
			TotalWithdrawAmount: int32(totals[month]),
		})
	}
	return results, nil
}

func (r *cardStatsWithdrawRepository) GetYearlyWithdrawAmount(ctx context.Context, year int) ([]*salemongo.YearlyWithdrawRow, error) {
	start := time.Date(year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	totals, err := sumByYear(ctx, r.db, "withdraws", "withdraw_time", "withdraw_amount", start, end)
	if err != nil {
		return nil, card_errors.ErrGetYearlyWithdrawAmountFailed.WithInternal(err)
	}

	results := make([]*salemongo.YearlyWithdrawRow, 0, 5)
	for y := year - 4; y <= year; y++ {
		results = append(results, &salemongo.YearlyWithdrawRow{
			Year:                yearString(y),
			TotalWithdrawAmount: totals[y],
		})
	}
	return results, nil
}
