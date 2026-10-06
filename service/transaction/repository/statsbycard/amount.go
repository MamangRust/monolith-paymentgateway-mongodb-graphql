package transactionbycardrepository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	transaction_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/transaction_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type transactionStatsByCardAmountRepository struct {
	db *mongo.Database
}

func NewTransactionStatsByCardAmountRepository(db *mongo.Database) TransactionStatsByCardAmountRepository {
	return &transactionStatsByCardAmountRepository{db: db}
}

func (r *transactionStatsByCardAmountRepository) coll() *mongo.Collection {
	return r.db.Collection("transactions")
}

func monthName(m time.Month) string {
	return time.Date(0, m, 1, 0, 0, 0, 0, time.UTC).Format("Jan")
}

func yearString(year int) string {
	return time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006")
}

func (r *transactionStatsByCardAmountRepository) GetMonthlyAmountsByCardNumber(ctx context.Context, req *requests.MonthYearPaymentMethod) ([]*salemongo.TransactionMonthlyAmountByCardRow, error) {
	start := time.Date(req.Year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":       nil,
			"card_number":      req.CardNumber,
			"transaction_time": bson.M{"$gte": start, "$lt": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":          bson.M{"month": bson.M{"$month": "$transaction_time"}},
			"total_amount": bson.M{"$sum": "$amount"},
		}}},
	})
	if err != nil {
		return nil, transaction_errors.ErrGetMonthlyAmountsByCardFailed.WithInternal(err)
	}
	defer cursor.Close(ctx)

	var rows []struct {
		ID struct {
			Month int `bson:"month"`
		} `bson:"_id"`
		TotalAmount int64 `bson:"total_amount"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, transaction_errors.ErrGetMonthlyAmountsByCardFailed.WithInternal(err)
	}

	totals := make(map[int]int64, len(rows))
	for _, row := range rows {
		totals[row.ID.Month] = row.TotalAmount
	}

	results := make([]*salemongo.TransactionMonthlyAmountByCardRow, 0, 12)
	for month := 1; month <= 12; month++ {
		results = append(results, &salemongo.TransactionMonthlyAmountByCardRow{
			Month:       monthName(time.Month(month)),
			TotalAmount: int32(totals[month]),
		})
	}
	return results, nil
}

func (r *transactionStatsByCardAmountRepository) GetYearlyAmountsByCardNumber(ctx context.Context, req *requests.MonthYearPaymentMethod) ([]*salemongo.TransactionYearlyAmountByCardRow, error) {
	start := time.Date(req.Year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(req.Year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":       nil,
			"card_number":      req.CardNumber,
			"transaction_time": bson.M{"$gte": start, "$lt": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":          bson.M{"year": bson.M{"$year": "$transaction_time"}},
			"total_amount": bson.M{"$sum": "$amount"},
		}}},
	})
	if err != nil {
		return nil, transaction_errors.ErrGetYearlyAmountsByCardFailed.WithInternal(err)
	}
	defer cursor.Close(ctx)

	var rows []struct {
		ID struct {
			Year int `bson:"year"`
		} `bson:"_id"`
		TotalAmount int64 `bson:"total_amount"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, transaction_errors.ErrGetYearlyAmountsByCardFailed.WithInternal(err)
	}

	totals := make(map[int]int64, len(rows))
	for _, row := range rows {
		totals[row.ID.Year] = row.TotalAmount
	}

	results := make([]*salemongo.TransactionYearlyAmountByCardRow, 0, 5)
	for year := req.Year - 4; year <= req.Year; year++ {
		results = append(results, &salemongo.TransactionYearlyAmountByCardRow{
			Year:        yearString(year),
			TotalAmount: totals[year],
		})
	}
	return results, nil
}
