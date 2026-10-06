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

type topupStatsByCardStatusRepository struct {
	db *mongo.Database
}

func NewTopupStatsByCardStatusRepository(db *mongo.Database) TopupStatsByCardStatusRepository {
	return &topupStatsByCardStatusRepository{db: db}
}

func (r *topupStatsByCardStatusRepository) coll() *mongo.Collection {
	return r.db.Collection("topups")
}

// cardStatusTotal counts the rows of the given status for one card inside the window.
func (r *topupStatsByCardStatusRepository) cardStatusTotal(ctx context.Context, cardNumber, status string, start, end time.Time) (int64, int64, error) {
	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":  nil,
			"card_number": cardNumber,
			"status":      status,
			"topup_time":  bson.M{"$gte": start, "$lte": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":          nil,
			"total_status": bson.M{"$sum": 1},
			"total_amount": bson.M{"$sum": "$topup_amount"},
		}}},
	})
	if err != nil {
		return 0, 0, err
	}
	defer cursor.Close(ctx)

	var rows []struct {
		TotalStatus int64 `bson:"total_status"`
		TotalAmount int64 `bson:"total_amount"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return 0, 0, err
	}
	if len(rows) == 0 {
		return 0, 0, nil
	}
	return rows[0].TotalStatus, rows[0].TotalAmount, nil
}

func (r *topupStatsByCardStatusRepository) GetMonthTopupStatusSuccessByCardNumber(ctx context.Context, req *requests.MonthTopupStatusCardNumber) ([]*salemongo.TopupMonthlyStatusRow, error) {
	results := make([]*salemongo.TopupMonthlyStatusRow, 0, 2)

	currentDate := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC)
	prevDate := currentDate.AddDate(0, -1, 0)

	for _, periodStart := range []time.Time{currentDate, prevDate} {
		periodEnd := periodStart.AddDate(0, 1, -1)

		count, amount, err := r.cardStatusTotal(ctx, req.CardNumber, "success", periodStart, periodEnd)
		if err != nil {
			return nil, topup_errors.ErrGetMonthTopupStatusSuccessByCardFailed.WithInternal(err)
		}

		results = append(results, &salemongo.TopupMonthlyStatusRow{
			Year:         yearString(periodStart.Year()),
			Month:        monthName(periodStart.Month()),
			TotalSuccess: count,
			TotalAmount:  int32(amount),
		})
	}
	return results, nil
}

func (r *topupStatsByCardStatusRepository) GetMonthTopupStatusFailedByCardNumber(ctx context.Context, req *requests.MonthTopupStatusCardNumber) ([]*salemongo.TopupMonthlyStatusRow, error) {
	results := make([]*salemongo.TopupMonthlyStatusRow, 0, 2)

	currentDate := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC)
	prevDate := currentDate.AddDate(0, -1, 0)

	for _, periodStart := range []time.Time{currentDate, prevDate} {
		periodEnd := periodStart.AddDate(0, 1, -1)

		count, amount, err := r.cardStatusTotal(ctx, req.CardNumber, "failed", periodStart, periodEnd)
		if err != nil {
			return nil, topup_errors.ErrGetMonthTopupStatusFailedByCardFailed.WithInternal(err)
		}

		results = append(results, &salemongo.TopupMonthlyStatusRow{
			Year:        yearString(periodStart.Year()),
			Month:       monthName(periodStart.Month()),
			TotalFailed: count,
			TotalAmount: int32(amount),
		})
	}
	return results, nil
}

func (r *topupStatsByCardStatusRepository) GetYearlyTopupStatusSuccessByCardNumber(ctx context.Context, req *requests.YearTopupStatusCardNumber) ([]*salemongo.TopupYearlyStatusRow, error) {
	results := make([]*salemongo.TopupYearlyStatusRow, 0, 2)

	for _, year := range []int{req.Year, req.Year - 1} {
		count, amount, err := r.cardStatusTotal(ctx, req.CardNumber, "success",
			time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC))
		if err != nil {
			return nil, topup_errors.ErrGetYearlyTopupStatusSuccessByCardFailed.WithInternal(err)
		}

		results = append(results, &salemongo.TopupYearlyStatusRow{
			Year:         yearString(year),
			TotalSuccess: int32(count),
			TotalAmount:  int32(amount),
		})
	}
	return results, nil
}

func (r *topupStatsByCardStatusRepository) GetYearlyTopupStatusFailedByCardNumber(ctx context.Context, req *requests.YearTopupStatusCardNumber) ([]*salemongo.TopupYearlyStatusRow, error) {
	results := make([]*salemongo.TopupYearlyStatusRow, 0, 2)

	for _, year := range []int{req.Year, req.Year - 1} {
		count, amount, err := r.cardStatusTotal(ctx, req.CardNumber, "failed",
			time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC))
		if err != nil {
			return nil, topup_errors.ErrGetYearlyTopupStatusSuccessByCardFailed.WithInternal(err)
		}

		results = append(results, &salemongo.TopupYearlyStatusRow{
			Year:        yearString(year),
			TotalFailed: int32(count),
			TotalAmount: int32(amount),
		})
	}
	return results, nil
}
