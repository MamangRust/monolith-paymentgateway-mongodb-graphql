package withdrawstatsrepository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	withdraw_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/withdraw_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type withdrawStatsStatusRepository struct {
	db *mongo.Database
}

func NewWithdrawStatsStatusRepository(db *mongo.Database) WithdrawStatsStatusRepository {
	return &withdrawStatsStatusRepository{db: db}
}

func (r *withdrawStatsStatusRepository) coll() *mongo.Collection {
	return r.db.Collection("withdraws")
}

type statusTotal struct {
	Count  int64
	Amount int64
}

// statusTotalFor totals one status inside the [start, end] window.
func (r *withdrawStatsStatusRepository) statusTotalFor(ctx context.Context, status string, start, end time.Time) (statusTotal, error) {
	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":    nil,
			"status":        status,
			"withdraw_time": bson.M{"$gte": start, "$lte": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":          nil,
			"total_status": bson.M{"$sum": 1},
			"total_amount": bson.M{"$sum": "$withdraw_amount"},
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

// statusTotalPerYear groups one status by year, restricted to the two years
// that the yearly statistics report on (the requested year and the one before).
func (r *withdrawStatsStatusRepository) statusTotalPerYear(ctx context.Context, status string, year int) (map[int]statusTotal, error) {
	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":    nil,
			"status":        status,
			"withdraw_time": bson.M{"$gte": time.Date(year-1, 1, 1, 0, 0, 0, 0, time.UTC), "$lt": time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":          bson.M{"year": bson.M{"$year": "$withdraw_time"}},
			"total_status": bson.M{"$sum": 1},
			"total_amount": bson.M{"$sum": "$withdraw_amount"},
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

func (r *withdrawStatsStatusRepository) GetMonthWithdrawStatusSuccess(ctx context.Context, req *requests.MonthStatusWithdraw) ([]*salemongo.WithdrawMonthlyStatusSuccessRow, error) {
	currentDate := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC)
	prevDate := currentDate.AddDate(0, -1, 0)

	results := make([]*salemongo.WithdrawMonthlyStatusSuccessRow, 0, 2)
	for _, periodStart := range []time.Time{currentDate, prevDate} {
		periodEnd := periodStart.AddDate(0, 1, -1)

		total, err := r.statusTotalFor(ctx, "success", periodStart, periodEnd)
		if err != nil {
			return nil, withdraw_errors.ErrGetMonthWithdrawStatusSuccessFailed.WithInternal(err)
		}

		results = append(results, &salemongo.WithdrawMonthlyStatusSuccessRow{
			Year:         yearString(periodStart.Year()),
			Month:        monthName(periodStart.Month()),
			TotalSuccess: total.Count,
			TotalAmount:  int32(total.Amount),
		})
	}
	return results, nil
}

func (r *withdrawStatsStatusRepository) GetMonthWithdrawStatusFailed(ctx context.Context, req *requests.MonthStatusWithdraw) ([]*salemongo.WithdrawMonthlyStatusFailedRow, error) {
	currentDate := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC)
	prevDate := currentDate.AddDate(0, -1, 0)

	results := make([]*salemongo.WithdrawMonthlyStatusFailedRow, 0, 2)
	for _, periodStart := range []time.Time{currentDate, prevDate} {
		periodEnd := periodStart.AddDate(0, 1, -1)

		total, err := r.statusTotalFor(ctx, "failed", periodStart, periodEnd)
		if err != nil {
			return nil, withdraw_errors.ErrGetMonthWithdrawStatusFailedFailed.WithInternal(err)
		}

		results = append(results, &salemongo.WithdrawMonthlyStatusFailedRow{
			Year:        yearString(periodStart.Year()),
			Month:       monthName(periodStart.Month()),
			TotalFailed: total.Count,
			TotalAmount: int32(total.Amount),
		})
	}
	return results, nil
}

func (r *withdrawStatsStatusRepository) GetYearlyWithdrawStatusSuccess(ctx context.Context, year int) ([]*salemongo.WithdrawYearlyStatusSuccessRow, error) {
	totals, err := r.statusTotalPerYear(ctx, "success", year)
	if err != nil {
		return nil, withdraw_errors.ErrGetYearlyWithdrawStatusSuccessFailed.WithInternal(err)
	}

	results := make([]*salemongo.WithdrawYearlyStatusSuccessRow, 0, 2)
	for _, y := range []int{year, year - 1} {
		results = append(results, &salemongo.WithdrawYearlyStatusSuccessRow{
			Year:         yearString(y),
			TotalSuccess: int32(totals[y].Count),
			TotalAmount:  int32(totals[y].Amount),
		})
	}
	return results, nil
}

func (r *withdrawStatsStatusRepository) GetYearlyWithdrawStatusFailed(ctx context.Context, year int) ([]*salemongo.WithdrawYearlyStatusFailedRow, error) {
	totals, err := r.statusTotalPerYear(ctx, "failed", year)
	if err != nil {
		return nil, withdraw_errors.ErrGetYearlyWithdrawStatusFailedFailed.WithInternal(err)
	}

	results := make([]*salemongo.WithdrawYearlyStatusFailedRow, 0, 2)
	for _, y := range []int{year, year - 1} {
		results = append(results, &salemongo.WithdrawYearlyStatusFailedRow{
			Year:        yearString(y),
			TotalFailed: int32(totals[y].Count),
			TotalAmount: int32(totals[y].Amount),
		})
	}
	return results, nil
}
