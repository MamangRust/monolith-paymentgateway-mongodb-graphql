package transactionbycardrepository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	transaction_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/transaction_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type transactionStatsByCardStatusRepository struct {
	db *mongo.Database
}

func NewTransactionStatsByCardStatusRepository(db *mongo.Database) TransactionStatsByCardStatusRepository {
	return &transactionStatsByCardStatusRepository{db: db}
}

func (r *transactionStatsByCardStatusRepository) coll() *mongo.Collection {
	return r.db.Collection("transactions")
}

type cardStatusTotal struct {
	Count  int64
	Amount int64
}

// cardStatusTotalFor totals one status for one card inside the [start, end] window.
func (r *transactionStatsByCardStatusRepository) cardStatusTotalFor(ctx context.Context, cardNumber, status string, start, end time.Time) (cardStatusTotal, error) {
	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":       nil,
			"card_number":      cardNumber,
			"status":           status,
			"transaction_time": bson.M{"$gte": start, "$lte": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":          nil,
			"total_status": bson.M{"$sum": 1},
			"total_amount": bson.M{"$sum": "$amount"},
		}}},
	})
	if err != nil {
		return cardStatusTotal{}, err
	}
	defer cursor.Close(ctx)

	var rows []struct {
		TotalStatus int64 `bson:"total_status"`
		TotalAmount int64 `bson:"total_amount"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return cardStatusTotal{}, err
	}
	if len(rows) == 0 {
		return cardStatusTotal{}, nil
	}
	return cardStatusTotal{Count: rows[0].TotalStatus, Amount: rows[0].TotalAmount}, nil
}

func (r *transactionStatsByCardStatusRepository) GetMonthTransactionStatusSuccessByCardNumber(ctx context.Context, req *requests.MonthStatusTransactionCardNumber) ([]*salemongo.TransactionMonthlyStatusSuccessByCardRow, error) {
	currentDate := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC)
	prevDate := currentDate.AddDate(0, -1, 0)

	results := make([]*salemongo.TransactionMonthlyStatusSuccessByCardRow, 0, 2)
	for _, periodStart := range []time.Time{currentDate, prevDate} {
		periodEnd := periodStart.AddDate(0, 1, -1)

		total, err := r.cardStatusTotalFor(ctx, req.CardNumber, "success", periodStart, periodEnd)
		if err != nil {
			return nil, transaction_errors.ErrGetMonthTransactionStatusSuccessByCardFailed.WithInternal(err)
		}

		results = append(results, &salemongo.TransactionMonthlyStatusSuccessByCardRow{
			Year:         yearString(periodStart.Year()),
			Month:        monthName(periodStart.Month()),
			TotalSuccess: total.Count,
			TotalAmount:  int32(total.Amount),
		})
	}
	return results, nil
}

func (r *transactionStatsByCardStatusRepository) GetMonthTransactionStatusFailedByCardNumber(ctx context.Context, req *requests.MonthStatusTransactionCardNumber) ([]*salemongo.TransactionMonthlyStatusFailedByCardRow, error) {
	currentDate := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC)
	prevDate := currentDate.AddDate(0, -1, 0)

	results := make([]*salemongo.TransactionMonthlyStatusFailedByCardRow, 0, 2)
	for _, periodStart := range []time.Time{currentDate, prevDate} {
		periodEnd := periodStart.AddDate(0, 1, -1)

		total, err := r.cardStatusTotalFor(ctx, req.CardNumber, "failed", periodStart, periodEnd)
		if err != nil {
			return nil, transaction_errors.ErrGetMonthTransactionStatusFailedByCardFailed.WithInternal(err)
		}

		results = append(results, &salemongo.TransactionMonthlyStatusFailedByCardRow{
			Year:        yearString(periodStart.Year()),
			Month:       monthName(periodStart.Month()),
			TotalFailed: total.Count,
			TotalAmount: int32(total.Amount),
		})
	}
	return results, nil
}

func (r *transactionStatsByCardStatusRepository) GetYearlyTransactionStatusSuccessByCardNumber(ctx context.Context, req *requests.YearStatusTransactionCardNumber) ([]*salemongo.TransactionYearlyStatusSuccessByCardRow, error) {
	results := make([]*salemongo.TransactionYearlyStatusSuccessByCardRow, 0, 2)

	for _, year := range []int{req.Year, req.Year - 1} {
		total, err := r.cardStatusTotalFor(ctx, req.CardNumber, "success",
			time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC))
		if err != nil {
			return nil, transaction_errors.ErrGetYearlyTransactionStatusSuccessByCardFailed.WithInternal(err)
		}

		results = append(results, &salemongo.TransactionYearlyStatusSuccessByCardRow{
			Year:         yearString(year),
			TotalSuccess: int32(total.Count),
			TotalAmount:  int32(total.Amount),
		})
	}
	return results, nil
}

func (r *transactionStatsByCardStatusRepository) GetYearlyTransactionStatusFailedByCardNumber(ctx context.Context, req *requests.YearStatusTransactionCardNumber) ([]*salemongo.TransactionYearlyStatusFailedByCardRow, error) {
	results := make([]*salemongo.TransactionYearlyStatusFailedByCardRow, 0, 2)

	for _, year := range []int{req.Year, req.Year - 1} {
		total, err := r.cardStatusTotalFor(ctx, req.CardNumber, "failed",
			time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC))
		if err != nil {
			return nil, transaction_errors.ErrGetYearlyTransactionStatusFailedByCardFailed.WithInternal(err)
		}

		results = append(results, &salemongo.TransactionYearlyStatusFailedByCardRow{
			Year:        yearString(year),
			TotalFailed: int32(total.Count),
			TotalAmount: int32(total.Amount),
		})
	}
	return results, nil
}
