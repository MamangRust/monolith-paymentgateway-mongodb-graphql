package transactionstatsrepository

import (
	"context"
	"sort"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	transaction_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/transaction_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type transactionStatsMethodRepository struct {
	db *mongo.Database
}

func NewTransactionStatsMethodRepository(db *mongo.Database) TransactionStatsMethodRepository {
	return &transactionStatsMethodRepository{db: db}
}

func (r *transactionStatsMethodRepository) coll() *mongo.Collection {
	return r.db.Collection("transactions")
}

// distinctPaymentMethods returns every payment method seen so far.
func (r *transactionStatsMethodRepository) distinctPaymentMethods(ctx context.Context) ([]string, error) {
	values, err := r.coll().Distinct(ctx, "payment_method", bson.M{"deleted_at": nil})
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

type methodTotal struct {
	Key    int
	Method string
	Count  int64
	Amount int64
}

func (r *transactionStatsMethodRepository) GetMonthlyPaymentMethods(ctx context.Context, year int) ([]*salemongo.TransactionMonthlyPaymentMethodRow, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	methods, err := r.distinctPaymentMethods(ctx)
	if err != nil {
		return nil, transaction_errors.ErrGetMonthlyPaymentMethodsFailed.WithInternal(err)
	}

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
			"total_count":  bson.M{"$sum": 1},
			"total_amount": bson.M{"$sum": "$amount"},
		}}},
	})
	if err != nil {
		return nil, transaction_errors.ErrGetMonthlyPaymentMethodsFailed.WithInternal(err)
	}
	defer cursor.Close(ctx)

	var rows []struct {
		ID struct {
			Month  int    `bson:"month"`
			Method string `bson:"method"`
		} `bson:"_id"`
		TotalCount  int64 `bson:"total_count"`
		TotalAmount int64 `bson:"total_amount"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, transaction_errors.ErrGetMonthlyPaymentMethodsFailed.WithInternal(err)
	}

	index := make(map[[2]interface{}][2]int64, len(rows))
	for _, row := range rows {
		index[[2]interface{}{row.ID.Month, row.ID.Method}] = [2]int64{row.TotalCount, row.TotalAmount}
	}

	results := make([]*salemongo.TransactionMonthlyPaymentMethodRow, 0, 12*len(methods))
	for month := 1; month <= 12; month++ {
		for _, method := range methods {
			total := index[[2]interface{}{month, method}]
			results = append(results, &salemongo.TransactionMonthlyPaymentMethodRow{
				Month:         monthName(time.Month(month)),
				PaymentMethod: method,
				TotalCount:    int32(total[0]),
				TotalAmount:   int32(total[1]),
			})
		}
	}
	return results, nil
}

func (r *transactionStatsMethodRepository) GetYearlyPaymentMethods(ctx context.Context, year int) ([]*salemongo.TransactionYearlyPaymentMethodRow, error) {
	start := time.Date(year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":       nil,
			"transaction_time": bson.M{"$gte": start, "$lt": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"year":   bson.M{"$year": "$transaction_time"},
				"method": "$payment_method",
			},
			"total_count":  bson.M{"$sum": 1},
			"total_amount": bson.M{"$sum": "$amount"},
		}}},
		{{Key: "$sort", Value: bson.M{"_id.year": 1, "_id.method": 1}}},
	})
	if err != nil {
		return nil, transaction_errors.ErrGetYearlyPaymentMethodsFailed.WithInternal(err)
	}
	defer cursor.Close(ctx)

	var rows []struct {
		ID struct {
			Year   int    `bson:"year"`
			Method string `bson:"method"`
		} `bson:"_id"`
		TotalCount  int64 `bson:"total_count"`
		TotalAmount int64 `bson:"total_amount"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, transaction_errors.ErrGetYearlyPaymentMethodsFailed.WithInternal(err)
	}

	results := make([]*salemongo.TransactionYearlyPaymentMethodRow, 0, len(rows))
	for _, row := range rows {
		results = append(results, &salemongo.TransactionYearlyPaymentMethodRow{
			Year:          yearString(row.ID.Year),
			PaymentMethod: row.ID.Method,
			TotalCount:    row.TotalCount,
			TotalAmount:   row.TotalAmount,
		})
	}
	return results, nil
}
