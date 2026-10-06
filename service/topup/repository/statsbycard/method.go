package topupstatsbycardrepository

import (
	"context"
	"sort"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	topup_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/topup_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type topupStatsByCardMethodRepository struct {
	db *mongo.Database
}

func NewTopupStatsByCardMethodRepository(db *mongo.Database) TopupStatsByCardMethodRepository {
	return &topupStatsByCardMethodRepository{db: db}
}

func (r *topupStatsByCardMethodRepository) coll() *mongo.Collection {
	return r.db.Collection("topups")
}

// distinctMethods returns every topup method seen so far.
func (r *topupStatsByCardMethodRepository) distinctMethods(ctx context.Context) ([]string, error) {
	values, err := r.coll().Distinct(ctx, "topup_method", bson.M{"deleted_at": nil})
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

type monthlyMethodTotal struct {
	Month  int
	Method string
	Count  int64
	Amount int64
}

func (r *topupStatsByCardMethodRepository) monthlyMethodTotals(ctx context.Context, cardNumber string, start, end time.Time) ([]monthlyMethodTotal, error) {
	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":  nil,
			"card_number": cardNumber,
			"topup_time":  bson.M{"$gte": start, "$lt": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"month":  bson.M{"$month": "$topup_time"},
				"method": "$topup_method",
			},
			"total_topups": bson.M{"$sum": 1},
			"total_amount": bson.M{"$sum": "$topup_amount"},
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
		TotalTopups int64 `bson:"total_topups"`
		TotalAmount int64 `bson:"total_amount"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}

	totals := make([]monthlyMethodTotal, 0, len(rows))
	for _, row := range rows {
		totals = append(totals, monthlyMethodTotal{
			Month:  row.ID.Month,
			Method: row.ID.Method,
			Count:  row.TotalTopups,
			Amount: row.TotalAmount,
		})
	}
	return totals, nil
}

func (r *topupStatsByCardMethodRepository) GetMonthlyTopupMethodsByCardNumber(ctx context.Context, req *requests.YearMonthMethod) ([]*salemongo.TopupMonthlyMethodByCardRow, error) {
	start := time.Date(req.Year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	methods, err := r.distinctMethods(ctx)
	if err != nil {
		return nil, topup_errors.ErrGetMonthlyTopupMethodsByCardFailed.WithInternal(err)
	}

	totals, err := r.monthlyMethodTotals(ctx, req.CardNumber, start, end)
	if err != nil {
		return nil, topup_errors.ErrGetMonthlyTopupMethodsByCardFailed.WithInternal(err)
	}

	index := make(map[[2]interface{}]monthlyMethodTotal, len(totals))
	for _, total := range totals {
		index[[2]interface{}{total.Month, total.Method}] = total
	}

	results := make([]*salemongo.TopupMonthlyMethodByCardRow, 0, 12*len(methods))
	for month := 1; month <= 12; month++ {
		for _, method := range methods {
			total := index[[2]interface{}{month, method}]
			results = append(results, &salemongo.TopupMonthlyMethodByCardRow{
				Month:       monthName(time.Month(month)),
				TopupMethod: method,
				TotalTopups: int32(total.Count),
				TotalAmount: int32(total.Amount),
			})
		}
	}
	return results, nil
}

func (r *topupStatsByCardMethodRepository) GetYearlyTopupMethodsByCardNumber(ctx context.Context, req *requests.YearMonthMethod) ([]*salemongo.TopupYearlyMethodByCardRow, error) {
	start := time.Date(req.Year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(req.Year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":  nil,
			"card_number": req.CardNumber,
			"topup_time":  bson.M{"$gte": start, "$lt": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"year":   bson.M{"$year": "$topup_time"},
				"method": "$topup_method",
			},
			"total_topups": bson.M{"$sum": 1},
			"total_amount": bson.M{"$sum": "$topup_amount"},
		}}},
		{{Key: "$sort", Value: bson.M{"_id.year": 1, "_id.method": 1}}},
	})
	if err != nil {
		return nil, topup_errors.ErrGetYearlyTopupMethodsByCardFailed.WithInternal(err)
	}
	defer cursor.Close(ctx)

	var rows []struct {
		ID struct {
			Year   int    `bson:"year"`
			Method string `bson:"method"`
		} `bson:"_id"`
		TotalTopups int64 `bson:"total_topups"`
		TotalAmount int64 `bson:"total_amount"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, topup_errors.ErrGetYearlyTopupMethodsByCardFailed.WithInternal(err)
	}

	results := make([]*salemongo.TopupYearlyMethodByCardRow, 0, len(rows))
	for _, row := range rows {
		results = append(results, &salemongo.TopupYearlyMethodByCardRow{
			Year:        yearString(row.ID.Year),
			TopupMethod: row.ID.Method,
			TotalTopups: row.TotalTopups,
			TotalAmount: row.TotalAmount,
		})
	}
	return results, nil
}
