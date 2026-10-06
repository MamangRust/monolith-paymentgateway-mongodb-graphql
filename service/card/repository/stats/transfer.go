package repositorystats

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	card_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/card_errors/repository"
	"go.mongodb.org/mongo-driver/mongo"
)

type cardStaticsTransferRepository struct {
	db *mongo.Database
}

func NewCardStatsTransferRepository(db *mongo.Database) CardStatsTransferRepository {
	return &cardStaticsTransferRepository{db: db}
}

func (r *cardStaticsTransferRepository) GetMonthlyTransferAmountSender(ctx context.Context, year int) ([]*salemongo.MonthlyTransferSentRow, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	totals, err := sumByMonth(ctx, r.db, "transfers", "transfer_time", "transfer_amount", start, end)
	if err != nil {
		return nil, card_errors.ErrGetMonthlyTransferAmountSenderFailed.WithInternal(err)
	}

	results := make([]*salemongo.MonthlyTransferSentRow, 0, 12)
	for month := 1; month <= 12; month++ {
		results = append(results, &salemongo.MonthlyTransferSentRow{
			Month:           monthName(time.Month(month)),
			TotalSentAmount: int32(totals[month]),
		})
	}
	return results, nil
}

func (r *cardStaticsTransferRepository) GetYearlyTransferAmountSender(ctx context.Context, year int) ([]*salemongo.YearlyTransferSentRow, error) {
	start := time.Date(year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	totals, err := sumByYear(ctx, r.db, "transfers", "transfer_time", "transfer_amount", start, end)
	if err != nil {
		return nil, card_errors.ErrGetYearlyTransferAmountSenderFailed.WithInternal(err)
	}

	results := make([]*salemongo.YearlyTransferSentRow, 0, 5)
	for y := year - 4; y <= year; y++ {
		results = append(results, &salemongo.YearlyTransferSentRow{
			Year:            yearString(y),
			TotalSentAmount: totals[y],
		})
	}
	return results, nil
}

func (r *cardStaticsTransferRepository) GetMonthlyTransferAmountReceiver(ctx context.Context, year int) ([]*salemongo.MonthlyTransferReceivedRow, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	totals, err := sumByMonth(ctx, r.db, "transfers", "transfer_time", "transfer_amount", start, end)
	if err != nil {
		return nil, card_errors.ErrGetMonthlyTransferAmountReceiverFailed.WithInternal(err)
	}

	results := make([]*salemongo.MonthlyTransferReceivedRow, 0, 12)
	for month := 1; month <= 12; month++ {
		results = append(results, &salemongo.MonthlyTransferReceivedRow{
			Month:               monthName(time.Month(month)),
			TotalReceivedAmount: int32(totals[month]),
		})
	}
	return results, nil
}

func (r *cardStaticsTransferRepository) GetYearlyTransferAmountReceiver(ctx context.Context, year int) ([]*salemongo.YearlyTransferReceivedRow, error) {
	start := time.Date(year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	totals, err := sumByYear(ctx, r.db, "transfers", "transfer_time", "transfer_amount", start, end)
	if err != nil {
		return nil, card_errors.ErrGetYearlyTransferAmountReceiverFailed.WithInternal(err)
	}

	results := make([]*salemongo.YearlyTransferReceivedRow, 0, 5)
	for y := year - 4; y <= year; y++ {
		results = append(results, &salemongo.YearlyTransferReceivedRow{
			Year:                yearString(y),
			TotalReceivedAmount: totals[y],
		})
	}
	return results, nil
}
