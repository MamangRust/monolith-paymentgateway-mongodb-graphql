package topupstatsrepository

import (
	"context"
	"sort"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	topup_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/topup_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type topupStatsMethodRepository struct {
	db *mongo.Database
}

func NewTopupStatsMethodRepository(db *mongo.Database) TopupStatsMethodRepository {
	return &topupStatsMethodRepository{db: db}
}

func (r *topupStatsMethodRepository) coll() *mongo.Collection {
	return r.db.Collection("topups")
}

func (r *topupStatsMethodRepository) GetMonthlyTopupMethods(ctx context.Context, year int) ([]*salemongo.TopupMonthlyMethodRow, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at": nil,
			"topup_time": bson.M{"$gte": start, "$lt": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"month":  bson.M{"$month": "$topup_time"},
				"method": "$topup_method",
			},
			"total_topups": bson.M{"$sum": 1},
			"total_amount": bson.M{"$sum": "$topup_amount"},
		}}},
	})
	if err != nil {
		return nil, topup_errors.ErrGetMonthlyTopupMethodsFailed.WithInternal(err)
	}
	defer cursor.Close(ctx)

	var rows []struct {
		ID struct {
			Month  int    `bson:"month"`
			Method string `bson:"method"`
		} `bson:"_id"`
		TotalTopups int64 `bson:"total_topups"`
		TotalAmount int64 `bson:"total_amount"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, topup_errors.ErrGetMonthlyTopupMethodsFailed.WithInternal(err)
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].ID.Month != rows[j].ID.Month {
			return rows[i].ID.Month < rows[j].ID.Month
		}
		return rows[i].ID.Method < rows[j].ID.Method
	})

	results := make([]*salemongo.TopupMonthlyMethodRow, 0, len(rows))
	for _, row := range rows {
		results = append(results, &salemongo.TopupMonthlyMethodRow{
			Month:       monthName(time.Month(row.ID.Month)),
			TopupMethod: row.ID.Method,
			TotalTopups: int32(row.TotalTopups),
			TotalAmount: int32(row.TotalAmount),
		})
	}
	return results, nil
}

func (r *topupStatsMethodRepository) GetYearlyTopupMethods(ctx context.Context, year int) ([]*salemongo.TopupYearlyMethodRow, error) {
	start := time.Date(year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at": nil,
			"topup_time": bson.M{"$gte": start, "$lt": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"year":   bson.M{"$year": "$topup_time"},
				"method": "$topup_method",
			},
			"total_topups": bson.M{"$sum": 1},
			"total_amount": bson.M{"$sum": "$topup_amount"},
		}}},
		{{Key: "$sort", Value: bson.M{"_id.year": 1, "_id.method": 1}}},
	})
	if err != nil {
		return nil, topup_errors.ErrGetYearlyTopupMethodsFailed.WithInternal(err)
	}
	defer cursor.Close(ctx)

	var rows []struct {
		ID struct {
			Year   int    `bson:"year"`
			Method string `bson:"method"`
		} `bson:"_id"`
		TotalTopups int64 `bson:"total_topups"`
		TotalAmount int64 `bson:"total_amount"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, topup_errors.ErrGetYearlyTopupMethodsFailed.WithInternal(err)
	}

	results := make([]*salemongo.TopupYearlyMethodRow, 0, len(rows))
	for _, row := range rows {
		results = append(results, &salemongo.TopupYearlyMethodRow{
			Year:        yearString(row.ID.Year),
			TopupMethod: row.ID.Method,
			TotalTopups: row.TotalTopups,
			TotalAmount: row.TotalAmount,
		})
	}
	return results, nil
}
