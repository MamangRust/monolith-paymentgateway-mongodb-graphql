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

type cardStatsWithdrawByCardRepository struct {
	db *mongo.Database
}

func NewCardStatsWithdrawByCardRepository(db *mongo.Database) CardStatsWithdrawByCardRepository {
	return &cardStatsWithdrawByCardRepository{db: db}
}

func (r *cardStatsWithdrawByCardRepository) GetMonthlyWithdrawAmountByCardNumber(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyWithdrawRow, error) {
	start := time.Date(req.Year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	totals, err := aggregateTotals(ctx, r.db, "withdraws", "withdraw_time", "withdraw_amount", "month", bson.M{"card_number": req.CardNumber}, start, end)
	if err != nil {
		return nil, card_errors.ErrGetMonthlyWithdrawAmountByCardFailed.WithInternal(err)
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

func (r *cardStatsWithdrawByCardRepository) GetYearlyWithdrawAmountByCardNumber(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyWithdrawRow, error) {
	start := time.Date(req.Year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(req.Year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	totals, err := aggregateTotals(ctx, r.db, "withdraws", "withdraw_time", "withdraw_amount", "year", bson.M{"card_number": req.CardNumber}, start, end)
	if err != nil {
		return nil, card_errors.ErrGetYearlyWithdrawAmountByCardFailed.WithInternal(err)
	}

	results := make([]*salemongo.YearlyWithdrawRow, 0, 5)
	for year := req.Year - 4; year <= req.Year; year++ {
		results = append(results, &salemongo.YearlyWithdrawRow{
			Year:                yearString(year),
			TotalWithdrawAmount: totals[year],
		})
	}
	return results, nil
}
