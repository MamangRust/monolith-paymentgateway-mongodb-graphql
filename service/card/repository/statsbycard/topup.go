package repositorystatsbycard

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	card_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/card_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type cardStatsTopupByCardRepository struct {
	db *mongo.Database
}

func NewCardStatsTopupByCardRepository(db *mongo.Database) CardStatsTopupByCardRepository {
	return &cardStatsTopupByCardRepository{db: db}
}

func (r *cardStatsTopupByCardRepository) GetMonthlyTopupAmountByCardNumber(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyTopupRow, error) {
	start := time.Date(req.Year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	totals, err := aggregateTotals(ctx, r.db, "topups", "topup_time", "topup_amount", "month", bson.M{"card_number": req.CardNumber}, start, end)
	if err != nil {
		return nil, card_errors.ErrGetMonthlyTopupAmountByCardFailed.WithInternal(err)
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

func (r *cardStatsTopupByCardRepository) GetYearlyTopupAmountByCardNumber(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyTopupRow, error) {
	start := time.Date(req.Year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(req.Year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	totals, err := aggregateTotals(ctx, r.db, "topups", "topup_time", "topup_amount", "year", bson.M{"card_number": req.CardNumber}, start, end)
	if err != nil {
		return nil, card_errors.ErrGetYearlyTopupAmountByCardFailed.WithInternal(err)
	}

	results := make([]*salemongo.YearlyTopupRow, 0, 5)
	for year := req.Year - 4; year <= req.Year; year++ {
		results = append(results, &salemongo.YearlyTopupRow{
			Year:             yearString(year),
			TotalTopupAmount: totals[year],
		})
	}
	return results, nil
}
