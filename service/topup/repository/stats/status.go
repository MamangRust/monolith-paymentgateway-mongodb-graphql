package topupstatsrepository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	topup_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/topup_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type topupStatsStatusRepository struct {
	db *mongo.Database
}

func NewTopupStatsStatusRepository(db *mongo.Database) TopupStatsStatusRepository {
	return &topupStatsStatusRepository{db: db}
}

func (r *topupStatsStatusRepository) coll() *mongo.Collection {
	return r.db.Collection("topups")
}

// aggregateStatus totals the given status over the [start, end] window.
func (r *topupStatsStatusRepository) aggregateStatus(ctx context.Context, status string, start, end time.Time) (int64, int64, error) {
	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at": nil,
			"status":     status,
			"topup_time": bson.M{"$gte": start, "$lte": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":          nil,
			"total_status": bson.M{"$sum": 1},
			"total_amount": bson.M{"$sum": "$topup_amount"},
		}}},
	})
	if err != nil {
		return 0, 0, err
	}
	defer cursor.Close(ctx)

	var rows []struct {
		TotalStatus int64 `bson:"total_status"`
		TotalAmount int64 `bson:"total_amount"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return 0, 0, err
	}
	if len(rows) == 0 {
		return 0, 0, nil
	}
	return rows[0].TotalStatus, rows[0].TotalAmount, nil
}

// aggregateYearlyStatus totals the given status grouped by year.
func (r *topupStatsStatusRepository) aggregateYearlyStatus(ctx context.Context, status string, start, end time.Time) (map[int]statusTotal, error) {
	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at": nil,
			"status":     status,
			"topup_time": bson.M{"$gte": start, "$lt": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":          bson.M{"year": bson.M{"$year": "$topup_time"}},
			"total_status": bson.M{"$sum": 1},
			"total_amount": bson.M{"$sum": "$topup_amount"},
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

type statusTotal struct {
	Count  int64
	Amount int64
}

func (r *topupStatsStatusRepository) GetMonthTopupStatusSuccess(ctx context.Context, req *requests.MonthTopupStatus) ([]*salemongo.TopupMonthlyStatusRow, error) {
	rows, err := r.monthlyStatusRows(ctx, req, "success")
	if err != nil {
		return nil, topup_errors.ErrGetMonthTopupStatusSuccessFailed.WithInternal(err)
	}
	return rows, nil
}

func (r *topupStatsStatusRepository) GetMonthTopupStatusFailed(ctx context.Context, req *requests.MonthTopupStatus) ([]*salemongo.TopupMonthlyStatusRow, error) {
	rows, err := r.monthlyStatusRows(ctx, req, "failed")
	if err != nil {
		return nil, topup_errors.ErrGetMonthTopupStatusFailedFailed.WithInternal(err)
	}
	return rows, nil
}

// monthlyStatusRows returns the previous and the current month, ordered oldest first.
func (r *topupStatsStatusRepository) monthlyStatusRows(ctx context.Context, req *requests.MonthTopupStatus, status string) ([]*salemongo.TopupMonthlyStatusRow, error) {
	currentDate := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC)
	prevDate := currentDate.AddDate(0, -1, 0)

	periods := []time.Time{prevDate, currentDate}

	results := make([]*salemongo.TopupMonthlyStatusRow, 0, len(periods))
	for _, periodStart := range periods {
		periodEnd := periodStart.AddDate(0, 1, -1)

		count, amount, err := r.aggregateStatus(ctx, status, periodStart, periodEnd)
		if err != nil {
			return nil, err
		}

		row := &salemongo.TopupMonthlyStatusRow{
			Year:        yearString(periodStart.Year()),
			Month:       monthName(periodStart.Month()),
			TotalAmount: int32(amount),
		}
		if status == "success" {
			row.TotalSuccess = count
		} else {
			row.TotalFailed = count
		}
		results = append(results, row)
	}
	return results, nil
}

func (r *topupStatsStatusRepository) GetYearlyTopupStatusSuccess(ctx context.Context, year int) ([]*salemongo.TopupYearlyStatusRow, error) {
	totals, err := r.aggregateYearlyStatus(ctx, "success", time.Date(year-1, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		return nil, topup_errors.ErrGetYearlyTopupStatusSuccessFailed.WithInternal(err)
	}

	results := make([]*salemongo.TopupYearlyStatusRow, 0, 2)
	for y := year - 1; y <= year; y++ {
		results = append(results, &salemongo.TopupYearlyStatusRow{
			Year:         yearString(y),
			TotalSuccess: int32(totals[y].Count),
			TotalAmount:  int32(totals[y].Amount),
		})
	}
	return results, nil
}

func (r *topupStatsStatusRepository) GetYearlyTopupStatusFailed(ctx context.Context, year int) ([]*salemongo.TopupYearlyStatusRow, error) {
	totals, err := r.aggregateYearlyStatus(ctx, "failed", time.Date(year-1, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		return nil, topup_errors.ErrGetYearlyTopupStatusFailedFailed.WithInternal(err)
	}

	results := make([]*salemongo.TopupYearlyStatusRow, 0, 2)
	for y := year - 1; y <= year; y++ {
		results = append(results, &salemongo.TopupYearlyStatusRow{
			Year:        yearString(y),
			TotalFailed: int32(totals[y].Count),
			TotalAmount: int32(totals[y].Amount),
		})
	}
	return results, nil
}
