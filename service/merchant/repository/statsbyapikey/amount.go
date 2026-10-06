package merchantstatsapikeyrepository

import (
	"context"
	"errors"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	merchant_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/merchant_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type merchantStatsAmountByApiKeyRepository struct {
	db *mongo.Database
}

func NewMerchantStatsAmountByApiKeyRepository(db *mongo.Database) MerchantStatsAmountByApiKeyRepository {
	return &merchantStatsAmountByApiKeyRepository{db: db}
}

func (r *merchantStatsAmountByApiKeyRepository) coll() *mongo.Collection {
	return r.db.Collection("transactions")
}

func monthName(m time.Month) string {
	return time.Date(0, m, 1, 0, 0, 0, 0, time.UTC).Format("Jan")
}

// merchantIDByAPIKey resolves the owning merchant of an API key.
func (r *merchantStatsAmountByApiKeyRepository) merchantIDByAPIKey(ctx context.Context, apiKey string) (int32, bool, error) {
	var merchant salemongo.Merchant
	err := r.db.Collection("merchants").FindOne(ctx, bson.M{
		"api_key":    apiKey,
		"deleted_at": nil,
	}).Decode(&merchant)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return merchant.MerchantID, true, nil
}

func (r *merchantStatsAmountByApiKeyRepository) aggregateMonthlyAmounts(ctx context.Context, merchantID int32, year int) (map[int]int64, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":       nil,
			"merchant_id":      merchantID,
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

func (r *merchantStatsAmountByApiKeyRepository) aggregateYearlyAmounts(ctx context.Context, merchantID int32) (map[int]int64, error) {
	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":  nil,
			"merchant_id": merchantID,
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

func (r *merchantStatsAmountByApiKeyRepository) GetMonthlyAmountByApikey(ctx context.Context, req *requests.MonthYearAmountApiKey) ([]*salemongo.MerchantMonthlyAmountRow, error) {
	results := make([]*salemongo.MerchantMonthlyAmountRow, 0, 12)

	merchantID, found, err := r.merchantIDByAPIKey(ctx, req.Apikey)
	if err != nil {
		return nil, merchant_errors.ErrGetMonthlyAmountByApikeyFailed.WithInternal(err)
	}

	totals := map[int]int64{}
	if found {
		totals, err = r.aggregateMonthlyAmounts(ctx, merchantID, req.Year)
		if err != nil {
			return nil, merchant_errors.ErrGetMonthlyAmountByApikeyFailed.WithInternal(err)
		}
	}

	for month := 1; month <= 12; month++ {
		results = append(results, &salemongo.MerchantMonthlyAmountRow{
			Month:       monthName(time.Month(month)),
			TotalAmount: int32(totals[month]),
		})
	}
	return results, nil
}

func (r *merchantStatsAmountByApiKeyRepository) GetYearlyAmountByApikey(ctx context.Context, req *requests.MonthYearAmountApiKey) ([]*salemongo.MerchantYearlyAmountRow, error) {
	results := make([]*salemongo.MerchantYearlyAmountRow, 0, 5)

	merchantID, found, err := r.merchantIDByAPIKey(ctx, req.Apikey)
	if err != nil {
		return nil, merchant_errors.ErrGetYearlyAmountByApikeyFailed.WithInternal(err)
	}

	totals := map[int]int64{}
	if found {
		totals, err = r.aggregateYearlyAmounts(ctx, merchantID)
		if err != nil {
			return nil, merchant_errors.ErrGetYearlyAmountByApikeyFailed.WithInternal(err)
		}
	}

	for year := req.Year - 4; year <= req.Year; year++ {
		results = append(results, &salemongo.MerchantYearlyAmountRow{
			Year:        time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006"),
			TotalAmount: totals[year],
		})
	}
	return results, nil
}
