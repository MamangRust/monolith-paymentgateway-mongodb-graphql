package merchantstatsapikeyrepository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	merchant_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/merchant_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type merchantStatsTotalAmountByApiKeyRepository struct {
	db *mongo.Database
}

func NewMerchantStatsTotalAmountByApiKeyRepository(db *mongo.Database) MerchantStatsTotalAmountByApiKeyRepository {
	return &merchantStatsTotalAmountByApiKeyRepository{db: db}
}

func (r *merchantStatsTotalAmountByApiKeyRepository) coll() *mongo.Collection {
	return r.db.Collection("transactions")
}

func (r *merchantStatsTotalAmountByApiKeyRepository) merchantIDByAPIKey(ctx context.Context, apiKey string) (int32, bool, error) {
	var merchant salemongo.Merchant
	err := r.db.Collection("merchants").FindOne(ctx, bson.M{
		"api_key":    apiKey,
		"deleted_at": nil,
	}).Decode(&merchant)
	if err == mongo.ErrNoDocuments {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return merchant.MerchantID, true, nil
}

func (r *merchantStatsTotalAmountByApiKeyRepository) aggregateMonthlyTotals(ctx context.Context, merchantID int32, year int) (map[[2]int]int64, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":       nil,
			"merchant_id":      merchantID,
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

func (r *merchantStatsTotalAmountByApiKeyRepository) aggregateYearlyTotals(ctx context.Context, merchantID int32, year int) (map[int]int64, error) {
	start := time.Date(year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":       nil,
			"merchant_id":      merchantID,
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

func (r *merchantStatsTotalAmountByApiKeyRepository) GetMonthlyTotalAmountByApikey(ctx context.Context, req *requests.MonthYearTotalAmountApiKey) ([]*salemongo.MerchantMonthlyTotalAmountRow, error) {
	results := make([]*salemongo.MerchantMonthlyTotalAmountRow, 0, 12)

	merchantID, found, err := r.merchantIDByAPIKey(ctx, req.Apikey)
	if err != nil {
		return nil, merchant_errors.ErrGetMonthlyTotalAmountByApikeyFailed.WithInternal(err)
	}

	totals := map[[2]int]int64{}
	if found {
		totals, err = r.aggregateMonthlyTotals(ctx, merchantID, req.Year)
		if err != nil {
			return nil, merchant_errors.ErrGetMonthlyTotalAmountByApikeyFailed.WithInternal(err)
		}
	}

	for month := 1; month <= 12; month++ {
		results = append(results, &salemongo.MerchantMonthlyTotalAmountRow{
			Year:        time.Date(req.Year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006"),
			Month:       monthName(time.Month(month)),
			TotalAmount: int32(totals[[2]int{req.Year, month}]),
		})
	}
	return results, nil
}

func (r *merchantStatsTotalAmountByApiKeyRepository) GetYearlyTotalAmountByApikey(ctx context.Context, req *requests.MonthYearTotalAmountApiKey) ([]*salemongo.MerchantYearlyTotalAmountRow, error) {
	results := make([]*salemongo.MerchantYearlyTotalAmountRow, 0, 5)

	merchantID, found, err := r.merchantIDByAPIKey(ctx, req.Apikey)
	if err != nil {
		return nil, merchant_errors.ErrGetYearlyTotalAmountByApikeyFailed.WithInternal(err)
	}

	totals := map[int]int64{}
	if found {
		totals, err = r.aggregateYearlyTotals(ctx, merchantID, req.Year)
		if err != nil {
			return nil, merchant_errors.ErrGetYearlyTotalAmountByApikeyFailed.WithInternal(err)
		}
	}

	for year := req.Year - 4; year <= req.Year; year++ {
		results = append(results, &salemongo.MerchantYearlyTotalAmountRow{
			Year:        time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006"),
			TotalAmount: int32(totals[year]),
		})
	}
	return results, nil
}
