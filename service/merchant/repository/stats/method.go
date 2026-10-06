package merchantstatsrepository

import (
	"context"
	"sort"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	merchant_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/merchant_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type merchantStatsMethodRepository struct {
	db *mongo.Database
}

func NewMerchantStatsMethodRepository(db *mongo.Database) MerchantStatsMethodRepository {
	return &merchantStatsMethodRepository{db: db}
}

func (r *merchantStatsMethodRepository) coll() *mongo.Collection {
	return r.db.Collection("transactions")
}

func (r *merchantStatsMethodRepository) distinctPaymentMethods(ctx context.Context, start, end time.Time) ([]string, error) {
	values, err := r.coll().Distinct(ctx, "payment_method", bson.M{
		"deleted_at":       nil,
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

func (r *merchantStatsMethodRepository) aggregateMonthlyMethodAmounts(ctx context.Context, year int) (map[int]map[string]int64, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":       nil,
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

func (r *merchantStatsMethodRepository) aggregateYearlyMethodAmounts(ctx context.Context) (map[int]map[string]int64, error) {
	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"deleted_at": nil}}},
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

func (r *merchantStatsMethodRepository) GetMonthlyPaymentMethodsMerchant(ctx context.Context, year int) ([]*salemongo.MerchantMonthlyPaymentMethodRow, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	methods, err := r.distinctPaymentMethods(ctx, start, end)
	if err != nil {
		return nil, merchant_errors.ErrGetMonthlyPaymentMethodsMerchantFailed.WithInternal(err)
	}
	totals, err := r.aggregateMonthlyMethodAmounts(ctx, year)
	if err != nil {
		return nil, merchant_errors.ErrGetMonthlyPaymentMethodsMerchantFailed.WithInternal(err)
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

func (r *merchantStatsMethodRepository) GetYearlyPaymentMethodMerchant(ctx context.Context, year int) ([]*salemongo.MerchantYearlyPaymentMethodRow, error) {
	start := time.Date(year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	methods, err := r.distinctPaymentMethods(ctx, start, end)
	if err != nil {
		return nil, merchant_errors.ErrGetYearlyPaymentMethodMerchantFailed.WithInternal(err)
	}
	totals, err := r.aggregateYearlyMethodAmounts(ctx)
	if err != nil {
		return nil, merchant_errors.ErrGetYearlyPaymentMethodMerchantFailed.WithInternal(err)
	}

	var results []*salemongo.MerchantYearlyPaymentMethodRow
	for y := year - 4; y <= year; y++ {
		for _, method := range methods {
			results = append(results, &salemongo.MerchantYearlyPaymentMethodRow{
				Year:          time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006"),
				PaymentMethod: method,
				TotalAmount:   totals[y][method],
			})
		}
	}
	return results, nil
}
