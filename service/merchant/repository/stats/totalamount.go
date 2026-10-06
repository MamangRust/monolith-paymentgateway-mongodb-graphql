package merchantstatsrepository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	merchant_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/merchant_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type merchantStatsTotalAmountRepository struct {
	db *mongo.Database
}

func NewMerchantStatsTotalAmountRepository(db *mongo.Database) MerchantStatsTotalAmountRepository {
	return &merchantStatsTotalAmountRepository{db: db}
}

func (r *merchantStatsTotalAmountRepository) coll() *mongo.Collection {
	return r.db.Collection("transactions")
}

func (r *merchantStatsTotalAmountRepository) aggregateMonthlyTotals(ctx context.Context, year int) (map[[2]int]int64, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":       nil,
			"transaction_time": bson.M{"$gte": start, "$lt": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"year":  bson.M{"$year": "$transaction_time"},
				"month": bson.M{"$month": "$transaction_time"},
			},
			"total_amount": bson.M{"$sum": "$amount"},
		}}},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rows []struct {
		ID struct {
			Year  int `bson:"year"`
			Month int `bson:"month"`
		} `bson:"_id"`
		TotalAmount int64 `bson:"total_amount"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}

	totals := make(map[[2]int]int64, len(rows))
	for _, row := range rows {
		totals[[2]int{row.ID.Year, row.ID.Month}] = row.TotalAmount
	}
	return totals, nil
}

func (r *merchantStatsTotalAmountRepository) aggregateYearlyTotals(ctx context.Context, year int) (map[int]int64, error) {
	start := time.Date(year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":       nil,
			"transaction_time": bson.M{"$gte": start, "$lt": end},
		}}},
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

func (r *merchantStatsTotalAmountRepository) GetMonthlyTotalAmountMerchant(ctx context.Context, year int) ([]*salemongo.MerchantMonthlyTotalAmountRow, error) {
	totals, err := r.aggregateMonthlyTotals(ctx, year)
	if err != nil {
		return nil, merchant_errors.ErrGetMonthlyTotalAmountMerchantFailed.WithInternal(err)
	}

	results := make([]*salemongo.MerchantMonthlyTotalAmountRow, 0, 12)
	for month := 1; month <= 12; month++ {
		results = append(results, &salemongo.MerchantMonthlyTotalAmountRow{
			Year:        time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006"),
			Month:       monthName(time.Month(month)),
			TotalAmount: int32(totals[[2]int{year, month}]),
		})
	}
	return results, nil
}

func (r *merchantStatsTotalAmountRepository) GetYearlyTotalAmountMerchant(ctx context.Context, year int) ([]*salemongo.MerchantYearlyTotalAmountRow, error) {
	totals, err := r.aggregateYearlyTotals(ctx, year)
	if err != nil {
		return nil, merchant_errors.ErrGetYearlyTotalAmountMerchantFailed.WithInternal(err)
	}

	results := make([]*salemongo.MerchantYearlyTotalAmountRow, 0, 5)
	for y := year - 4; y <= year; y++ {
		results = append(results, &salemongo.MerchantYearlyTotalAmountRow{
			Year:        time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006"),
			TotalAmount: int32(totals[y]),
		})
	}
	return results, nil
}
