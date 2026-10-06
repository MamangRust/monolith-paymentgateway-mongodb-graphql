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

type merchantStatsTotalAmountByMerchantRepository struct {
	db *mongo.Database
}

func NewMerchantStatsTotalAmountByMerchantRepository(db *mongo.Database) MerchantStatsTotalAmountByMerchantRepository {
	return &merchantStatsTotalAmountByMerchantRepository{db: db}
}

func (r *merchantStatsTotalAmountByMerchantRepository) coll() *mongo.Collection {
	return r.db.Collection("transactions")
}

type merchantMonthTotal struct {
	Year        int
	Month       int
	TotalAmount int64
}

// aggregateMonthlyTotalsByMerchant totals the transaction amount per calendar
// month for a merchant across its whole history.
func (r *merchantStatsTotalAmountByMerchantRepository) aggregateMonthlyTotalsByMerchant(ctx context.Context, merchantID int) ([]merchantMonthTotal, error) {
	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":  nil,
			"merchant_id": int32(merchantID),
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

	totals := make([]merchantMonthTotal, 0, len(rows))
	for _, row := range rows {
		totals = append(totals, merchantMonthTotal{
			Year:        row.ID.Year,
			Month:       row.ID.Month,
			TotalAmount: row.TotalAmount,
		})
	}
	return totals, nil
}

// aggregateYearlyTotalsByMerchant totals the transaction amount per year for a
// merchant across its whole history.
func (r *merchantStatsTotalAmountByMerchantRepository) aggregateYearlyTotalsByMerchant(ctx context.Context, merchantID, year int) (map[int]int64, error) {
	start := time.Date(year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":       nil,
			"merchant_id":      int32(merchantID),
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

func (r *merchantStatsTotalAmountByMerchantRepository) GetMonthlyTotalAmountByMerchants(ctx context.Context, req *requests.MonthYearTotalAmountMerchant) ([]*salemongo.MerchantMonthlyTotalAmountRow, error) {
	totals, err := r.aggregateMonthlyTotalsByMerchant(ctx, req.MerchantID)
	if err != nil {
		return nil, merchant_errors.ErrGetMonthlyTotalAmountByMerchantsFailed.WithInternal(err)
	}

	byMonth := make(map[[2]int]int64, len(totals))
	for _, total := range totals {
		byMonth[[2]int{total.Year, total.Month}] = total.TotalAmount
	}

	results := make([]*salemongo.MerchantMonthlyTotalAmountRow, 0, 12)
	for month := 1; month <= 12; month++ {
		results = append(results, &salemongo.MerchantMonthlyTotalAmountRow{
			Year:        time.Date(req.Year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006"),
			Month:       monthName(time.Month(month)),
			TotalAmount: int32(byMonth[[2]int{req.Year, month}]),
		})
	}
	return results, nil
}

func (r *merchantStatsTotalAmountByMerchantRepository) GetYearlyTotalAmountByMerchants(ctx context.Context, req *requests.MonthYearTotalAmountMerchant) ([]*salemongo.MerchantYearlyTotalAmountRow, error) {
	totals, err := r.aggregateYearlyTotalsByMerchant(ctx, req.MerchantID, req.Year)
	if err != nil {
		return nil, merchant_errors.ErrGetYearlyTotalAmountByMerchantsFailed.WithInternal(err)
	}

	results := make([]*salemongo.MerchantYearlyTotalAmountRow, 0, 5)
	for year := req.Year - 4; year <= req.Year; year++ {
		results = append(results, &salemongo.MerchantYearlyTotalAmountRow{
			Year:        time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006"),
			TotalAmount: int32(totals[year]),
		})
	}
	return results, nil
}
