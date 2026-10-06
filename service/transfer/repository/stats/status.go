package transferstatsrepository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	transfer_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/transfer_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type transferStatsStatusRepository struct {
	db *mongo.Database
}

func NewTransferStatsStatusRepository(db *mongo.Database) TransferStatsStatusRepository {
	return &transferStatsStatusRepository{db: db}
}

// aggregateStatus sums the transfer count and amount inside the given window.
func (r *transferStatsStatusRepository) aggregateStatus(ctx context.Context, filter bson.M) (int64, int64, error) {
	cursor, err := r.db.Collection("transfers").Aggregate(ctx, salemongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.M{
			"_id":          nil,
			"total_count":  bson.M{"$sum": 1},
			"total_amount": bson.M{"$sum": "$transfer_amount"},
		}}},
	})
	if err != nil {
		return 0, 0, err
	}
	defer cursor.Close(ctx)

	var row struct {
		TotalCount  int64 `bson:"total_count"`
		TotalAmount int64 `bson:"total_amount"`
	}
	if cursor.Next(ctx) {
		if err := cursor.Decode(&row); err != nil {
			return 0, 0, err
		}
	}
	return row.TotalCount, row.TotalAmount, nil
}

func monthWindow(year, month int) (time.Time, time.Time) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	return start, start.AddDate(0, 1, 0)
}

func (r *transferStatsStatusRepository) GetMonthTransferStatusSuccess(ctx context.Context, req *requests.MonthStatusTransfer) ([]*salemongo.TransferMonthlyStatusSuccessRow, error) {
	results := make([]*salemongo.TransferMonthlyStatusSuccessRow, 0, 2)

	for i := 0; i < 2; i++ {
		month := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -i, 0)
		start := month
		end := month.AddDate(0, 1, 0)

		total, amount, err := r.aggregateStatus(ctx, bson.M{
			"deleted_at":    nil,
			"status":        "success",
			"transfer_time": bson.M{"$gte": start, "$lt": end},
		})
		if err != nil {
			return nil, transfer_errors.ErrGetMonthTransferStatusSuccessFailed.WithInternal(err)
		}

		results = append(results, &salemongo.TransferMonthlyStatusSuccessRow{
			Year:         start.Format("2006"),
			Month:        monthLabel(start.Month()),
			TotalSuccess: total,
			TotalAmount:  int32(amount),
		})
	}

	return results, nil
}

func (r *transferStatsStatusRepository) GetYearlyTransferStatusSuccess(ctx context.Context, year int) ([]*salemongo.TransferYearlyStatusSuccessRow, error) {
	results := make([]*salemongo.TransferYearlyStatusSuccessRow, 0, 2)

	for i := 0; i < 2; i++ {
		start := time.Date(year-i, 1, 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(1, 0, 0)

		total, amount, err := r.aggregateStatus(ctx, bson.M{
			"deleted_at":    nil,
			"status":        "success",
			"transfer_time": bson.M{"$gte": start, "$lt": end},
		})
		if err != nil {
			return nil, transfer_errors.ErrGetYearlyTransferStatusSuccessFailed.WithInternal(err)
		}

		results = append(results, &salemongo.TransferYearlyStatusSuccessRow{
			Year:         start.Format("2006"),
			TotalSuccess: int32(total),
			TotalAmount:  int32(amount),
		})
	}

	return results, nil
}

func (r *transferStatsStatusRepository) GetMonthTransferStatusFailed(ctx context.Context, req *requests.MonthStatusTransfer) ([]*salemongo.TransferMonthlyStatusFailedRow, error) {
	results := make([]*salemongo.TransferMonthlyStatusFailedRow, 0, 2)

	for i := 0; i < 2; i++ {
		start, end := monthWindow(req.Year, req.Month)
		start = start.AddDate(0, -i, 0)
		end = end.AddDate(0, -i, 0)

		total, amount, err := r.aggregateStatus(ctx, bson.M{
			"deleted_at":    nil,
			"status":        "failed",
			"transfer_time": bson.M{"$gte": start, "$lt": end},
		})
		if err != nil {
			return nil, transfer_errors.ErrGetMonthTransferStatusFailedFailed.WithInternal(err)
		}

		results = append(results, &salemongo.TransferMonthlyStatusFailedRow{
			Year:        start.Format("2006"),
			Month:       monthLabel(start.Month()),
			TotalFailed: total,
			TotalAmount: int32(amount),
		})
	}

	return results, nil
}

func (r *transferStatsStatusRepository) GetYearlyTransferStatusFailed(ctx context.Context, year int) ([]*salemongo.TransferYearlyStatusFailedRow, error) {
	results := make([]*salemongo.TransferYearlyStatusFailedRow, 0, 2)

	for i := 0; i < 2; i++ {
		start := time.Date(year-i, 1, 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(1, 0, 0)

		total, amount, err := r.aggregateStatus(ctx, bson.M{
			"deleted_at":    nil,
			"status":        "failed",
			"transfer_time": bson.M{"$gte": start, "$lt": end},
		})
		if err != nil {
			return nil, transfer_errors.ErrGetYearlyTransferStatusFailedFailed.WithInternal(err)
		}

		results = append(results, &salemongo.TransferYearlyStatusFailedRow{
			Year:        start.Format("2006"),
			TotalFailed: int32(total),
			TotalAmount: int32(amount),
		})
	}

	return results, nil
}
