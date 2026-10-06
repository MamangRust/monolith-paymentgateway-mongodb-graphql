package topupstatsbycardrepository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	topup_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/topup_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type topupStatsByCardAmountRepository struct {
	db *mongo.Database
}

func NewTopupStatsByCardAmountRepository(db *mongo.Database) TopupStatsByCardAmountRepository {
	return &topupStatsByCardAmountRepository{db: db}
}

func (r *topupStatsByCardAmountRepository) coll() *mongo.Collection {
	return r.db.Collection("topups")
}

func monthName(m time.Month) string {
	return time.Date(0, m, 1, 0, 0, 0, 0, time.UTC).Format("Jan")
}

func yearString(year int) string {
	return time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006")
}

func aggregateMonthlyTotals(ctx context.Context, coll *mongo.Collection, filter bson.M, start, end time.Time) (map[int]int64, error) {
	match := bson.M{}
	for k, v := range filter {
		match[k] = v
	}
	match["deleted_at"] = nil
	match["topup_time"] = bson.M{"$gte": start, "$lt": end}

	cursor, err := coll.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: match}},
		{{Key: "$group", Value: bson.M{
			"_id":          bson.M{"month": bson.M{"$month": "$topup_time"}},
			"total_amount": bson.M{"$sum": "$topup_amount"},
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

func aggregateYearlyTotals(ctx context.Context, coll *mongo.Collection, filter bson.M, start, end time.Time) (map[int]int64, error) {
	match := bson.M{}
	for k, v := range filter {
		match[k] = v
	}
	match["deleted_at"] = nil
	match["topup_time"] = bson.M{"$gte": start, "$lt": end}

	cursor, err := coll.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: match}},
		{{Key: "$group", Value: bson.M{
			"_id":          bson.M{"year": bson.M{"$year": "$topup_time"}},
			"total_amount": bson.M{"$sum": "$topup_amount"},
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

func (r *topupStatsByCardAmountRepository) GetMonthlyTopupAmountsByCardNumber(ctx context.Context, req *requests.YearMonthMethod) ([]*salemongo.TopupMonthlyAmountRow, error) {
	start := time.Date(req.Year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	totals, err := aggregateMonthlyTotals(ctx, r.coll(), bson.M{"card_number": req.CardNumber}, start, end)
	if err != nil {
		return nil, topup_errors.ErrGetMonthlyTopupAmountsByCardFailed.WithInternal(err)
	}

	results := make([]*salemongo.TopupMonthlyAmountRow, 0, 12)
	for month := 1; month <= 12; month++ {
		results = append(results, &salemongo.TopupMonthlyAmountRow{
			Month:       monthName(time.Month(month)),
			TotalAmount: int32(totals[month]),
		})
	}
	return results, nil
}

func (r *topupStatsByCardAmountRepository) GetYearlyTopupAmountsByCardNumber(ctx context.Context, req *requests.YearMonthMethod) ([]*salemongo.TopupYearlyAmountRow, error) {
	start := time.Date(req.Year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(req.Year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	totals, err := aggregateYearlyTotals(ctx, r.coll(), bson.M{"card_number": req.CardNumber}, start, end)
	if err != nil {
		return nil, topup_errors.ErrGetYearlyTopupAmountsByCardFailed.WithInternal(err)
	}

	results := make([]*salemongo.TopupYearlyAmountRow, 0, 5)
	for year := req.Year - 4; year <= req.Year; year++ {
		results = append(results, &salemongo.TopupYearlyAmountRow{
			Year:        yearString(year),
			TotalAmount: totals[year],
		})
	}
	return results, nil
}
