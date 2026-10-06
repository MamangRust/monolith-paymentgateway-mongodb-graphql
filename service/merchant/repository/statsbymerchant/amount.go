package merchantstatsmerchantrepository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	merchant_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/merchant_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type merchantStatsAmountByMerchantRepository struct {
	db *mongo.Database
}

func NewMerchantStatsAmountByMerchantRepository(db *mongo.Database) MerchantStatsAmountByMerchantRepository {
	return &merchantStatsAmountByMerchantRepository{db: db}
}

func (r *merchantStatsAmountByMerchantRepository) coll() *mongo.Collection {
	return r.db.Collection("transactions")
}

func monthName(m time.Month) string {
	return time.Date(0, m, 1, 0, 0, 0, 0, time.UTC).Format("Jan")
}

// aggregateMonthlyAmountsByMerchant sums the transaction amount per month for the
// given merchant inside a single calendar year.
func (r *merchantStatsAmountByMerchantRepository) aggregateMonthlyAmountsByMerchant(ctx context.Context, merchantID, year int) (map[int]int64, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":       nil,
			"merchant_id":      int32(merchantID),
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

// aggregateYearlyAmountsByMerchant sums the transaction amount per year for the
// given merchant across its whole history.
func (r *merchantStatsAmountByMerchantRepository) aggregateYearlyAmountsByMerchant(ctx context.Context, merchantID int) (map[int]int64, error) {
	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":  nil,
			"merchant_id": int32(merchantID),
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

func (r *merchantStatsAmountByMerchantRepository) GetMonthlyAmountByMerchants(ctx context.Context, req *requests.MonthYearAmountMerchant) ([]*salemongo.MerchantMonthlyAmountRow, error) {
	totals, err := r.aggregateMonthlyAmountsByMerchant(ctx, req.MerchantID, req.Year)
	if err != nil {
		return nil, merchant_errors.ErrGetMonthlyAmountByMerchantsFailed.WithInternal(err)
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

func (r *merchantStatsAmountByMerchantRepository) GetYearlyAmountByMerchants(ctx context.Context, req *requests.MonthYearAmountMerchant) ([]*salemongo.MerchantYearlyAmountRow, error) {
	totals, err := r.aggregateYearlyAmountsByMerchant(ctx, req.MerchantID)
	if err != nil {
		return nil, merchant_errors.ErrGetYearlyAmountByMerchantsFailed.WithInternal(err)
	}

	results := make([]*salemongo.MerchantYearlyAmountRow, 0, 5)
	for year := req.Year - 4; year <= req.Year; year++ {
		results = append(results, &salemongo.MerchantYearlyAmountRow{
			Year:        time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006"),
			TotalAmount: totals[year],
		})
	}
	return results, nil
}
