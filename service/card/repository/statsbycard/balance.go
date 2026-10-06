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

type cardStatsBalanceByCardRepository struct {
	db *mongo.Database
}

func NewCardStatsBalanceByCardRepository(db *mongo.Database) CardStatsBalanceByCardRepository {
	return &cardStatsBalanceByCardRepository{db: db}
}

func (r *cardStatsBalanceByCardRepository) GetMonthlyBalancesByCardNumber(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyBalanceRow, error) {
	start := time.Date(req.Year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	totals, err := aggregateTotals(ctx, r.db, "saldos", "created_at", "total_balance", "month", bson.M{"card_number": req.CardNumber}, start, end)
	if err != nil {
		return nil, card_errors.ErrGetMonthlyBalanceByCardFailed.WithInternal(err)
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

func (r *cardStatsBalanceByCardRepository) GetYearlyBalanceByCardNumber(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyBalanceRow, error) {
	start := time.Date(req.Year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(req.Year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	totals, err := aggregateTotals(ctx, r.db, "saldos", "created_at", "total_balance", "year", bson.M{"card_number": req.CardNumber}, start, end)
	if err != nil {
		return nil, card_errors.ErrGetYearlyBalanceByCardFailed.WithInternal(err)
	}

	results := make([]*salemongo.YearlyBalanceRow, 0, 5)
	for year := req.Year - 4; year <= req.Year; year++ {
		results = append(results, &salemongo.YearlyBalanceRow{
			Year:         yearString(year),
			TotalBalance: totals[year],
		})
	}
	return results, nil
}
