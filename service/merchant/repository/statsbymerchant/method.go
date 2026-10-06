package merchantstatsmerchantrepository

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

type merchantStatsMethodByMerchantRepository struct {
	db *mongo.Database
}

func NewMerchantStatsMethodByMerchantRepository(db *mongo.Database) MerchantStatsMethodByMerchantRepository {
	return &merchantStatsMethodByMerchantRepository{db: db}
}

func (r *merchantStatsMethodByMerchantRepository) coll() *mongo.Collection {
	return r.db.Collection("transactions")
}

// distinctPaymentMethods lists the payment methods ever used by a merchant inside
// the given time window.
func (r *merchantStatsMethodByMerchantRepository) distinctPaymentMethods(ctx context.Context, merchantID int, start, end time.Time) ([]string, error) {
	values, err := r.coll().Distinct(ctx, "payment_method", bson.M{
		"deleted_at":       nil,
		"merchant_id":      int32(merchantID),
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

// aggregateMonthlyMethodAmountsByMerchant totals the transaction amount per month
// and payment method for a single calendar year.
func (r *merchantStatsMethodByMerchantRepository) aggregateMonthlyMethodAmountsByMerchant(ctx context.Context, merchantID, year int) (map[int]map[string]int64, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":       nil,
			"merchant_id":      int32(merchantID),
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

// aggregateYearlyMethodAmountsByMerchant totals the transaction amount per year
// and payment method for a merchant's whole history.
func (r *merchantStatsMethodByMerchantRepository) aggregateYearlyMethodAmountsByMerchant(ctx context.Context, merchantID int) (map[int]map[string]int64, error) {
	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":  nil,
			"merchant_id": int32(merchantID),
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

func (r *merchantStatsMethodByMerchantRepository) GetMonthlyPaymentMethodByMerchants(ctx context.Context, req *requests.MonthYearPaymentMethodMerchant) ([]*salemongo.MerchantMonthlyPaymentMethodRow, error) {
	start := time.Date(req.Year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	methods, err := r.distinctPaymentMethods(ctx, req.MerchantID, start, end)
	if err != nil {
		return nil, merchant_errors.ErrGetMonthlyPaymentMethodByMerchantsFailed.WithInternal(err)
	}
	totals, err := r.aggregateMonthlyMethodAmountsByMerchant(ctx, req.MerchantID, req.Year)
	if err != nil {
		return nil, merchant_errors.ErrGetMonthlyPaymentMethodByMerchantsFailed.WithInternal(err)
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

func (r *merchantStatsMethodByMerchantRepository) GetYearlyPaymentMethodByMerchants(ctx context.Context, req *requests.MonthYearPaymentMethodMerchant) ([]*salemongo.MerchantYearlyPaymentMethodRow, error) {
	start := time.Date(req.Year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(req.Year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	methods, err := r.distinctPaymentMethods(ctx, req.MerchantID, start, end)
	if err != nil {
		return nil, merchant_errors.ErrGetYearlyPaymentMethodByMerchantsFailed.WithInternal(err)
	}
	totals, err := r.aggregateYearlyMethodAmountsByMerchant(ctx, req.MerchantID)
	if err != nil {
		return nil, merchant_errors.ErrGetYearlyPaymentMethodByMerchantsFailed.WithInternal(err)
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
