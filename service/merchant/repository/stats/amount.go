package merchantstatsrepository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	merchant_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/merchant_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type merchantStatsAmountRepository struct {
	db *mongo.Database
}

func NewMerchantStatsAmountRepository(db *mongo.Database) MerchantStatsAmountRepository {
	return &merchantStatsAmountRepository{db: db}
}

func (r *merchantStatsAmountRepository) coll() *mongo.Collection {
	return r.db.Collection("transactions")
}

func monthName(m time.Month) string {
	return time.Date(0, m, 1, 0, 0, 0, 0, time.UTC).Format("Jan")
}

func (r *merchantStatsAmountRepository) aggregateMonthlyAmounts(ctx context.Context, year int) (map[int]int64, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":       nil,
			"transaction_time": bson.M{"$gte": start, "$lt": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":          bson.M{"month": bson.M{"$month": "$transaction_time"}},
			"total_amount": bson.M{"$sum": "$amount"},
		}}},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

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

func (r *merchantStatsAmountRepository) aggregateYearlyAmounts(ctx context.Context) (map[int]int64, error) {
	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"deleted_at": nil}}},
		{{Key: "$group", Value: bson.M{
			"_id":          bson.M{"year": bson.M{"$year": "$transaction_time"}},
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

func (r *merchantStatsAmountRepository) GetMonthlyAmountMerchant(ctx context.Context, year int) ([]*salemongo.MerchantMonthlyAmountRow, error) {
	totals, err := r.aggregateMonthlyAmounts(ctx, year)
	if err != nil {
		return nil, merchant_errors.ErrGetMonthlyAmountMerchantFailed.WithInternal(err)
	}

	results := make([]*salemongo.MerchantMonthlyAmountRow, 0, 12)
	for month := 1; month <= 12; month++ {
		results = append(results, &salemongo.MerchantMonthlyAmountRow{
			Month:       monthName(time.Month(month)),
			TotalAmount: int32(totals[month]),
		})
	}
	return results, nil
}

func (r *merchantStatsAmountRepository) GetYearlyAmountMerchant(ctx context.Context, year int) ([]*salemongo.MerchantYearlyAmountRow, error) {
	totals, err := r.aggregateYearlyAmounts(ctx)
	if err != nil {
		return nil, merchant_errors.ErrGetYearlyAmountMerchantFailed.WithInternal(err)
	}

	results := make([]*salemongo.MerchantYearlyAmountRow, 0, 5)
	for y := year - 4; y <= year; y++ {
		results = append(results, &salemongo.MerchantYearlyAmountRow{
			Year:        time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006"),
			TotalAmount: totals[y],
		})
	}
	return results, nil
}
