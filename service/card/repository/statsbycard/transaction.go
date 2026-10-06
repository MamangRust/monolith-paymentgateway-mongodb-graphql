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

type cardStatsTransactionByCardRepository struct {
	db *mongo.Database
}

func NewCardStatsTransactionByCardRepository(db *mongo.Database) CardStatsTransactionByCardRepository {
	return &cardStatsTransactionByCardRepository{db: db}
}

func (r *cardStatsTransactionByCardRepository) GetMonthlyTransactionAmountByCardNumber(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyTransactionRow, error) {
	start := time.Date(req.Year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	totals, err := aggregateTotals(ctx, r.db, "transactions", "transaction_time", "amount", "month", bson.M{"card_number": req.CardNumber}, start, end)
	if err != nil {
		return nil, card_errors.ErrGetMonthlyTransactionAmountByCardFailed.WithInternal(err)
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

func (r *cardStatsTransactionByCardRepository) GetYearlyTransactionAmountByCardNumber(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyTransactionRow, error) {
	start := time.Date(req.Year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(req.Year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	totals, err := aggregateTotals(ctx, r.db, "transactions", "transaction_time", "amount", "year", bson.M{"card_number": req.CardNumber}, start, end)
	if err != nil {
		return nil, card_errors.ErrGetYearlyTransactionAmountByCardFailed.WithInternal(err)
	}

	results := make([]*salemongo.YearlyTransactionRow, 0, 5)
	for year := req.Year - 4; year <= req.Year; year++ {
		results = append(results, &salemongo.YearlyTransactionRow{
			Year:                   yearString(year),
			TotalTransactionAmount: totals[year],
		})
	}
	return results, nil
}
