package transactionstatsrepository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	transaction_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/transaction_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type transactionStatsStatusRepository struct {
	db *mongo.Database
}

func NewTransactionStatsStatusRepository(db *mongo.Database) TransactionStatsStatusRepository {
	return &transactionStatsStatusRepository{db: db}
}

func (r *transactionStatsStatusRepository) coll() *mongo.Collection {
	return r.db.Collection("transactions")
}

type statusTotal struct {
	Count  int64
	Amount int64
}

// statusTotalFor totals one status inside the [start, end] window.
func (r *transactionStatsStatusRepository) statusTotalFor(ctx context.Context, status string, start, end time.Time) (statusTotal, error) {
	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":       nil,
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
		return statusTotal{}, err
	}
	defer cursor.Close(ctx)

	var rows []struct {
		TotalStatus int64 `bson:"total_status"`
		TotalAmount int64 `bson:"total_amount"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return statusTotal{}, err
	}
	if len(rows) == 0 {
		return statusTotal{}, nil
	}
	return statusTotal{Count: rows[0].TotalStatus, Amount: rows[0].TotalAmount}, nil
}

// statusTotalPerYear groups one status by year inside the [start, end) window.
func (r *transactionStatsStatusRepository) statusTotalPerYear(ctx context.Context, status string, start, end time.Time) (map[int]statusTotal, error) {
	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":       nil,
			"status":           status,
			"transaction_time": bson.M{"$gte": start, "$lt": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":          bson.M{"year": bson.M{"$year": "$transaction_time"}},
			"total_status": bson.M{"$sum": 1},
			"total_amount": bson.M{"$sum": "$amount"},
		}}},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rows []struct {
		ID struct {
			Year int `bson:"year"`
		} `bson:"_id"`
		TotalStatus int64 `bson:"total_status"`
		TotalAmount int64 `bson:"total_amount"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}

	totals := make(map[int]statusTotal, len(rows))
	for _, row := range rows {
		totals[row.ID.Year] = statusTotal{Count: row.TotalStatus, Amount: row.TotalAmount}
	}
	return totals, nil
}

func (r *transactionStatsStatusRepository) GetMonthTransactionStatusSuccess(ctx context.Context, req *requests.MonthStatusTransaction) ([]*salemongo.TransactionMonthlyStatusSuccessRow, error) {
	currentDate := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC)
	prevDate := currentDate.AddDate(0, -1, 0)

	results := make([]*salemongo.TransactionMonthlyStatusSuccessRow, 0, 2)
	for _, periodStart := range []time.Time{currentDate, prevDate} {
		periodEnd := periodStart.AddDate(0, 1, -1)

		total, err := r.statusTotalFor(ctx, "success", periodStart, periodEnd)
		if err != nil {
			return nil, transaction_errors.ErrGetMonthTransactionStatusSuccessFailed.WithInternal(err)
		}

		results = append(results, &salemongo.TransactionMonthlyStatusSuccessRow{
			Year:         yearString(periodStart.Year()),
			Month:        monthName(periodStart.Month()),
			TotalSuccess: total.Count,
			TotalAmount:  int32(total.Amount),
		})
	}
	return results, nil
}

func (r *transactionStatsStatusRepository) GetMonthTransactionStatusFailed(ctx context.Context, req *requests.MonthStatusTransaction) ([]*salemongo.TransactionMonthlyStatusFailedRow, error) {
	currentDate := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC)
	prevDate := currentDate.AddDate(0, -1, 0)

	results := make([]*salemongo.TransactionMonthlyStatusFailedRow, 0, 2)
	for _, periodStart := range []time.Time{currentDate, prevDate} {
		periodEnd := periodStart.AddDate(0, 1, -1)

		total, err := r.statusTotalFor(ctx, "failed", periodStart, periodEnd)
		if err != nil {
			return nil, transaction_errors.ErrGetMonthTransactionStatusFailedFailed.WithInternal(err)
		}

		results = append(results, &salemongo.TransactionMonthlyStatusFailedRow{
			Year:        yearString(periodStart.Year()),
			Month:       monthName(periodStart.Month()),
			TotalFailed: total.Count,
			TotalAmount: int32(total.Amount),
		})
	}
	return results, nil
}

func (r *transactionStatsStatusRepository) GetYearlyTransactionStatusSuccess(ctx context.Context, year int) ([]*salemongo.TransactionYearlyStatusSuccessRow, error) {
	totals, err := r.statusTotalPerYear(ctx, "success",
		time.Date(year-1, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		return nil, transaction_errors.ErrGetYearlyTransactionStatusSuccessFailed.WithInternal(err)
	}

	results := make([]*salemongo.TransactionYearlyStatusSuccessRow, 0, 2)
	for _, y := range []int{year, year - 1} {
		results = append(results, &salemongo.TransactionYearlyStatusSuccessRow{
			Year:         yearString(y),
			TotalSuccess: int32(totals[y].Count),
			TotalAmount:  int32(totals[y].Amount),
		})
	}
	return results, nil
}

func (r *transactionStatsStatusRepository) GetYearlyTransactionStatusFailed(ctx context.Context, year int) ([]*salemongo.TransactionYearlyStatusFailedRow, error) {
	totals, err := r.statusTotalPerYear(ctx, "failed",
		time.Date(year-1, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		return nil, transaction_errors.ErrGetYearlyTransactionStatusFailedFailed.WithInternal(err)
	}

	results := make([]*salemongo.TransactionYearlyStatusFailedRow, 0, 2)
	for _, y := range []int{year, year - 1} {
		results = append(results, &salemongo.TransactionYearlyStatusFailedRow{
			Year:        yearString(y),
			TotalFailed: int32(totals[y].Count),
			TotalAmount: int32(totals[y].Amount),
		})
	}
	return results, nil
}
