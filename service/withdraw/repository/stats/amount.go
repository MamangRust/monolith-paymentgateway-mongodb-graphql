package withdrawstatsrepository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	withdraw_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/withdraw_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type withdrawStatsAmountRepository struct {
	db *mongo.Database
}

func NewWithdrawStatsAmountRepository(db *mongo.Database) WithdrawStatsAmountRepository {
	return &withdrawStatsAmountRepository{db: db}
}

func (r *withdrawStatsAmountRepository) coll() *mongo.Collection {
	return r.db.Collection("withdraws")
}

func monthName(m time.Month) string {
	return time.Date(0, m, 1, 0, 0, 0, 0, time.UTC).Format("Jan")
}

func yearString(year int) string {
	return time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006")
}

func (r *withdrawStatsAmountRepository) GetMonthlyWithdraws(ctx context.Context, year int) ([]*salemongo.WithdrawMonthlyAmountRow, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":    nil,
			"withdraw_time": bson.M{"$gte": start, "$lt": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":                   bson.M{"month": bson.M{"$month": "$withdraw_time"}},
			"total_withdraw_amount": bson.M{"$sum": "$withdraw_amount"},
		}}},
	})
	if err != nil {
		return nil, withdraw_errors.ErrGetMonthlyWithdrawsFailed.WithInternal(err)
	}
	defer cursor.Close(ctx)

	var rows []struct {
		ID struct {
			Month int `bson:"month"`
		} `bson:"_id"`
		TotalWithdrawAmount int64 `bson:"total_withdraw_amount"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, withdraw_errors.ErrGetMonthlyWithdrawsFailed.WithInternal(err)
	}

	totals := make(map[int]int64, len(rows))
	for _, row := range rows {
		totals[row.ID.Month] = row.TotalWithdrawAmount
	}

	results := make([]*salemongo.WithdrawMonthlyAmountRow, 0, 12)
	for month := 1; month <= 12; month++ {
		results = append(results, &salemongo.WithdrawMonthlyAmountRow{
			Month:               monthName(time.Month(month)),
			TotalWithdrawAmount: int32(totals[month]),
		})
	}
	return results, nil
}

func (r *withdrawStatsAmountRepository) GetYearlyWithdraws(ctx context.Context, year int) ([]*salemongo.WithdrawYearlyAmountRow, error) {
	start := time.Date(year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":    nil,
			"withdraw_time": bson.M{"$gte": start, "$lt": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":                   bson.M{"year": bson.M{"$year": "$withdraw_time"}},
			"total_withdraw_amount": bson.M{"$sum": "$withdraw_amount"},
		}}},
	})
	if err != nil {
		return nil, withdraw_errors.ErrGetYearlyWithdrawsFailed.WithInternal(err)
	}
	defer cursor.Close(ctx)

	var rows []struct {
		ID struct {
			Year int `bson:"year"`
		} `bson:"_id"`
		TotalWithdrawAmount int64 `bson:"total_withdraw_amount"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, withdraw_errors.ErrGetYearlyWithdrawsFailed.WithInternal(err)
	}

	totals := make(map[int]int64, len(rows))
	for _, row := range rows {
		totals[row.ID.Year] = row.TotalWithdrawAmount
	}

	results := make([]*salemongo.WithdrawYearlyAmountRow, 0, 5)
	for y := year - 4; y <= year; y++ {
		results = append(results, &salemongo.WithdrawYearlyAmountRow{
			Year:                yearString(y),
			TotalWithdrawAmount: totals[y],
		})
	}
	return results, nil
}
