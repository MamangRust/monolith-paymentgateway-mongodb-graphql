package transactionstatsrepository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	transaction_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/transaction_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type transactionStatsAmountRepository struct {
	db *mongo.Database
}

func NewTransactionStatsAmountRepository(db *mongo.Database) TransactionStatsAmountRepository {
	return &transactionStatsAmountRepository{db: db}
}

func (r *transactionStatsAmountRepository) coll() *mongo.Collection {
	return r.db.Collection("transactions")
}

func monthName(m time.Month) string {
	return time.Date(0, m, 1, 0, 0, 0, 0, time.UTC).Format("Jan")
}

func yearString(year int) string {
	return time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006")
}

func (r *transactionStatsAmountRepository) GetMonthlyAmounts(ctx context.Context, year int) ([]*salemongo.TransactionMonthlyAmountRow, error) {
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
		return nil, transaction_errors.ErrGetMonthlyAmountsFailed.WithInternal(err)
	}
	defer cursor.Close(ctx)

	var rows []struct {
		ID struct {
			Month int `bson:"month"`
		} `bson:"_id"`
		TotalAmount int64 `bson:"total_amount"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, transaction_errors.ErrGetMonthlyAmountsFailed.WithInternal(err)
	}

	totals := make(map[int]int64, len(rows))
	for _, row := range rows {
		totals[row.ID.Month] = row.TotalAmount
	}

	results := make([]*salemongo.TransactionMonthlyAmountRow, 0, 12)
	for month := 1; month <= 12; month++ {
		results = append(results, &salemongo.TransactionMonthlyAmountRow{
			Month:       monthName(time.Month(month)),
			TotalAmount: int32(totals[month]),
		})
	}
	return results, nil
}

func (r *transactionStatsAmountRepository) GetYearlyAmounts(ctx context.Context, year int) ([]*salemongo.TransactionYearlyAmountRow, error) {
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
		return nil, transaction_errors.ErrGetYearlyAmountsFailed.WithInternal(err)
	}
	defer cursor.Close(ctx)

	var rows []struct {
		ID struct {
			Year int `bson:"year"`
		} `bson:"_id"`
		TotalAmount int64 `bson:"total_amount"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, transaction_errors.ErrGetYearlyAmountsFailed.WithInternal(err)
	}

	totals := make(map[int]int64, len(rows))
	for _, row := range rows {
		totals[row.ID.Year] = row.TotalAmount
	}

	results := make([]*salemongo.TransactionYearlyAmountRow, 0, 5)
	for y := year - 4; y <= year; y++ {
		results = append(results, &salemongo.TransactionYearlyAmountRow{
			Year:        yearString(y),
			TotalAmount: totals[y],
		})
	}
	return results, nil
}
