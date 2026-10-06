package merchantstatsapikeyrepository

import (
	"context"
	"sort"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	merchant_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/merchant_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type merchantStatsMethodByApiKeyRepository struct {
	db *mongo.Database
}

func NewMerchantStatsMethodByApiKeyRepository(db *mongo.Database) MerchantStatsMethodByApiKeyRepository {
	return &merchantStatsMethodByApiKeyRepository{db: db}
}

func (r *merchantStatsMethodByApiKeyRepository) coll() *mongo.Collection {
	return r.db.Collection("transactions")
}

func (r *merchantStatsMethodByApiKeyRepository) merchantIDByAPIKey(ctx context.Context, apiKey string) (int32, bool, error) {
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

func (r *merchantStatsMethodByApiKeyRepository) distinctPaymentMethods(ctx context.Context, merchantID int32, start, end time.Time) ([]string, error) {
	values, err := r.coll().Distinct(ctx, "payment_method", bson.M{
		"deleted_at":       nil,
		"merchant_id":      merchantID,
		"transaction_time": bson.M{"$gte": start, "$lt": end},
	})
	if err != nil {
		return nil, err
	}

	methods := make([]string, 0, len(values))
	for _, value := range values {
		if method, ok := value.(string); ok && method != "" {
			methods = append(methods, method)
		}
	}
	sort.Strings(methods)
	return methods, nil
}

func (r *merchantStatsMethodByApiKeyRepository) aggregateMonthlyMethodAmounts(ctx context.Context, merchantID int32, year int) (map[int]map[string]int64, error) {
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
				"month":  bson.M{"$month": "$transaction_time"},
				"method": "$payment_method",
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
			Month  int    `bson:"month"`
			Method string `bson:"method"`
		} `bson:"_id"`
		TotalAmount int64 `bson:"total_amount"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}

	totals := make(map[int]map[string]int64, 12)
	for _, row := range rows {
		if totals[row.ID.Month] == nil {
			totals[row.ID.Month] = make(map[string]int64)
		}
		totals[row.ID.Month][row.ID.Method] = row.TotalAmount
	}
	return totals, nil
}

func (r *merchantStatsMethodByApiKeyRepository) aggregateYearlyMethodAmounts(ctx context.Context, merchantID int32) (map[int]map[string]int64, error) {
	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":  nil,
			"merchant_id": merchantID,
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"year":   bson.M{"$year": "$transaction_time"},
				"method": "$payment_method",
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
			Year   int    `bson:"year"`
			Method string `bson:"method"`
		} `bson:"_id"`
		TotalAmount int64 `bson:"total_amount"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}

	totals := make(map[int]map[string]int64)
	for _, row := range rows {
		if totals[row.ID.Year] == nil {
			totals[row.ID.Year] = make(map[string]int64)
		}
		totals[row.ID.Year][row.ID.Method] = row.TotalAmount
	}
	return totals, nil
}

func (r *merchantStatsMethodByApiKeyRepository) GetMonthlyPaymentMethodByApikey(ctx context.Context, req *requests.MonthYearPaymentMethodApiKey) ([]*salemongo.MerchantMonthlyPaymentMethodRow, error) {
	merchantID, found, err := r.merchantIDByAPIKey(ctx, req.Apikey)
	if err != nil {
		return nil, merchant_errors.ErrGetMonthlyPaymentMethodByApikeyFailed.WithInternal(err)
	}
	if !found {
		return []*salemongo.MerchantMonthlyPaymentMethodRow{}, nil
	}

	start := time.Date(req.Year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	methods, err := r.distinctPaymentMethods(ctx, merchantID, start, end)
	if err != nil {
		return nil, merchant_errors.ErrGetMonthlyPaymentMethodByApikeyFailed.WithInternal(err)
	}
	totals, err := r.aggregateMonthlyMethodAmounts(ctx, merchantID, req.Year)
	if err != nil {
		return nil, merchant_errors.ErrGetMonthlyPaymentMethodByApikeyFailed.WithInternal(err)
	}

	var results []*salemongo.MerchantMonthlyPaymentMethodRow
	for month := 1; month <= 12; month++ {
		for _, method := range methods {
			results = append(results, &salemongo.MerchantMonthlyPaymentMethodRow{
				Month:         monthName(time.Month(month)),
				PaymentMethod: method,
				TotalAmount:   int32(totals[month][method]),
			})
		}
	}
	return results, nil
}

func (r *merchantStatsMethodByApiKeyRepository) GetYearlyPaymentMethodByApikey(ctx context.Context, req *requests.MonthYearPaymentMethodApiKey) ([]*salemongo.MerchantYearlyPaymentMethodRow, error) {
	merchantID, found, err := r.merchantIDByAPIKey(ctx, req.Apikey)
	if err != nil {
		return nil, merchant_errors.ErrGetYearlyPaymentMethodByApikeyFailed.WithInternal(err)
	}
	if !found {
		return []*salemongo.MerchantYearlyPaymentMethodRow{}, nil
	}

	start := time.Date(req.Year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(req.Year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	methods, err := r.distinctPaymentMethods(ctx, merchantID, start, end)
	if err != nil {
		return nil, merchant_errors.ErrGetYearlyPaymentMethodByApikeyFailed.WithInternal(err)
	}
	totals, err := r.aggregateYearlyMethodAmounts(ctx, merchantID)
	if err != nil {
		return nil, merchant_errors.ErrGetYearlyPaymentMethodByApikeyFailed.WithInternal(err)
	}

	var results []*salemongo.MerchantYearlyPaymentMethodRow
	for year := req.Year - 4; year <= req.Year; year++ {
		for _, method := range methods {
			results = append(results, &salemongo.MerchantYearlyPaymentMethodRow{
				Year:          time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006"),
				PaymentMethod: method,
				TotalAmount:   totals[year][method],
			})
		}
	}
	return results, nil
}
