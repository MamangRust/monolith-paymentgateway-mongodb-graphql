package topupstatsrepository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	topup_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/topup_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type topupStatsAmountRepository struct {
	db *mongo.Database
}

func NewTopupStatsAmountRepository(db *mongo.Database) TopupStatsAmountRepository {
	return &topupStatsAmountRepository{db: db}
}

func (r *topupStatsAmountRepository) coll() *mongo.Collection {
	return r.db.Collection("topups")
}

func monthName(m time.Month) string {
	return time.Date(0, m, 1, 0, 0, 0, 0, time.UTC).Format("Jan")
}

func (r *topupStatsAmountRepository) aggregateMonthlyAmounts(ctx context.Context, year int) (map[int]int64, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at": nil,
			"topup_time": bson.M{"$gte": start, "$lt": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":          bson.M{"month": bson.M{"$month": "$topup_time"}},
			"total_amount": bson.M{"$sum": "$topup_amount"},
		}}},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	rows, err := decodeMonthTotals(ctx, cursor)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *topupStatsAmountRepository) aggregateYearlyAmounts(ctx context.Context) (map[int]int64, error) {
	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"deleted_at": nil}}},
		{{Key: "$group", Value: bson.M{
			"_id":          bson.M{"year": bson.M{"$year": "$topup_time"}},
			"total_amount": bson.M{"$sum": "$topup_amount"},
		}}},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	return decodeYearTotals(ctx, cursor)
}

func (r *topupStatsAmountRepository) GetMonthlyTopupAmounts(ctx context.Context, year int) ([]*salemongo.TopupMonthlyAmountRow, error) {
	totals, err := r.aggregateMonthlyAmounts(ctx, year)
	if err != nil {
		return nil, topup_errors.ErrGetMonthlyTopupAmountsFailed.WithInternal(err)
	}

	results := make([]*salemongo.TopupMonthlyAmountRow, 0, 12)
	for month := 1; month <= 12; month++ {
		results = append(results, &salemongo.TopupMonthlyAmountRow{
			Month:       monthName(time.Month(month)),
			TotalAmount: int32(totals[month]),
		})
	}
	return results, nil
}

func (r *topupStatsAmountRepository) GetYearlyTopupAmounts(ctx context.Context, year int) ([]*salemongo.TopupYearlyAmountRow, error) {
	totals, err := r.aggregateYearlyAmounts(ctx)
	if err != nil {
		return nil, topup_errors.ErrGetYearlyTopupAmountsFailed.WithInternal(err)
	}

	results := make([]*salemongo.TopupYearlyAmountRow, 0, 5)
	for y := year - 4; y <= year; y++ {
		results = append(results, &salemongo.TopupYearlyAmountRow{
			Year:        yearString(y),
			TotalAmount: totals[y],
		})
	}
	return results, nil
}

func yearString(year int) string {
	return time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006")
}

func decodeMonthTotals(ctx context.Context, cursor *mongo.Cursor) (map[int]int64, error) {
	var rows []struct {
		ID struct {
			Month int `bson:"month"`
		} `bson:"_id"`
		TotalAmount int64 `bson:"total_amount"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}

	totals := make(map[int]int64, len(rows))
	for _, row := range rows {
		totals[row.ID.Month] = row.TotalAmount
	}
	return totals, nil
}

func decodeYearTotals(ctx context.Context, cursor *mongo.Cursor) (map[int]int64, error) {
	var rows []struct {
		ID struct {
			Year int `bson:"year"`
		} `bson:"_id"`
		TotalAmount int64 `bson:"total_amount"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}

	totals := make(map[int]int64, len(rows))
	for _, row := range rows {
		totals[row.ID.Year] = row.TotalAmount
	}
	return totals, nil
}
