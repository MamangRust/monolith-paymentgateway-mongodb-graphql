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

type cardStatsTransferByCardRepository struct {
	db *mongo.Database
}

func NewCardStatsTransferByCardRepository(db *mongo.Database) CardStatsTransferByCardRepository {
	return &cardStatsTransferByCardRepository{db: db}
}

func (r *cardStatsTransferByCardRepository) GetMonthlyTransferAmountBySender(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyTransferSentRow, error) {
	start := time.Date(req.Year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	totals, err := aggregateTotals(ctx, r.db, "transfers", "transfer_time", "transfer_amount", "month", bson.M{"transfer_from": req.CardNumber}, start, end)
	if err != nil {
		return nil, card_errors.ErrGetMonthlyTransferAmountBySenderFailed.WithInternal(err)
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

func (r *cardStatsTransferByCardRepository) GetYearlyTransferAmountBySender(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyTransferSentRow, error) {
	start := time.Date(req.Year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(req.Year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	totals, err := aggregateTotals(ctx, r.db, "transfers", "transfer_time", "transfer_amount", "year", bson.M{"transfer_from": req.CardNumber}, start, end)
	if err != nil {
		return nil, card_errors.ErrGetYearlyTransferAmountBySenderFailed.WithInternal(err)
	}

	results := make([]*salemongo.YearlyTransferSentRow, 0, 5)
	for year := req.Year - 4; year <= req.Year; year++ {
		results = append(results, &salemongo.YearlyTransferSentRow{
			Year:            yearString(year),
			TotalSentAmount: totals[year],
		})
	}
	return results, nil
}

func (r *cardStatsTransferByCardRepository) GetMonthlyTransferAmountByReceiver(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.MonthlyTransferReceivedRow, error) {
	start := time.Date(req.Year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	totals, err := aggregateTotals(ctx, r.db, "transfers", "transfer_time", "transfer_amount", "month", bson.M{"transfer_to": req.CardNumber}, start, end)
	if err != nil {
		return nil, card_errors.ErrGetMonthlyTransferAmountByReceiverFailed.WithInternal(err)
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

func (r *cardStatsTransferByCardRepository) GetYearlyTransferAmountByReceiver(ctx context.Context, req *requests.MonthYearCardNumberCard) ([]*salemongo.YearlyTransferReceivedRow, error) {
	start := time.Date(req.Year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(req.Year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	totals, err := aggregateTotals(ctx, r.db, "transfers", "transfer_time", "transfer_amount", "year", bson.M{"transfer_to": req.CardNumber}, start, end)
	if err != nil {
		return nil, card_errors.ErrGetYearlyTransferAmountByReceiverFailed.WithInternal(err)
	}

	results := make([]*salemongo.YearlyTransferReceivedRow, 0, 5)
	for year := req.Year - 4; year <= req.Year; year++ {
		results = append(results, &salemongo.YearlyTransferReceivedRow{
			Year:                yearString(year),
			TotalReceivedAmount: totals[year],
		})
	}
	return results, nil
}
