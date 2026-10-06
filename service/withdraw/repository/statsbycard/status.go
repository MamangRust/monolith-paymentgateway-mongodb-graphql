package withdrawstatsbycardrepository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	withdraw_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/withdraw_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type withdrawStatsByCardStatusRepository struct {
	db *mongo.Database
}

type withdrawStatsByCardAmountRepository struct {
	db *mongo.Database
}

type repository struct {
	WithdrawStatsByCardStatusRepository
	WithdrawStatsByCardAmountRepository
}

func NewWithdrawStatsByCardRepository(db *mongo.Database) WithdrawStatsByCardRepository {
	return &repository{
		WithdrawStatsByCardStatusRepository: NewWithdrawStatsByCardStatusRepository(db),
		WithdrawStatsByCardAmountRepository: NewWithdrawStatsByCardAmountRepository(db),
	}
}

func NewWithdrawStatsByCardStatusRepository(db *mongo.Database) WithdrawStatsByCardStatusRepository {
	return &withdrawStatsByCardStatusRepository{db: db}
}

func NewWithdrawStatsByCardAmountRepository(db *mongo.Database) WithdrawStatsByCardAmountRepository {
	return &withdrawStatsByCardAmountRepository{db: db}
}

func monthName(m time.Month) string {
	return time.Date(0, m, 1, 0, 0, 0, 0, time.UTC).Format("Jan")
}

func yearString(year int) string {
	return time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006")
}

func (r *withdrawStatsByCardAmountRepository) coll() *mongo.Collection {
	return r.db.Collection("withdraws")
}

func (r *withdrawStatsByCardStatusRepository) coll() *mongo.Collection {
	return r.db.Collection("withdraws")
}

type cardStatusTotal struct {
	Count  int64
	Amount int64
}

// cardStatusTotalFor totals one status for one card inside the [start, end] window.
func (r *withdrawStatsByCardStatusRepository) cardStatusTotalFor(ctx context.Context, cardNumber, status string, start, end time.Time) (cardStatusTotal, error) {
	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":    nil,
			"card_number":   cardNumber,
			"status":        status,
			"withdraw_time": bson.M{"$gte": start, "$lte": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":          nil,
			"total_status": bson.M{"$sum": 1},
			"total_amount": bson.M{"$sum": "$withdraw_amount"},
		}}},
	})
	if err != nil {
		return cardStatusTotal{}, err
	}
	defer cursor.Close(ctx)

	var rows []struct {
		TotalStatus int64 `bson:"total_status"`
		TotalAmount int64 `bson:"total_amount"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return cardStatusTotal{}, err
	}
	if len(rows) == 0 {
		return cardStatusTotal{}, nil
	}
	return cardStatusTotal{Count: rows[0].TotalStatus, Amount: rows[0].TotalAmount}, nil
}

func (r *withdrawStatsByCardStatusRepository) GetMonthWithdrawStatusSuccessByCardNumber(ctx context.Context, req *requests.MonthStatusWithdrawCardNumber) ([]*salemongo.WithdrawMonthlyStatusSuccessByCardRow, error) {
	currentDate := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC)
	prevDate := currentDate.AddDate(0, -1, 0)

	results := make([]*salemongo.WithdrawMonthlyStatusSuccessByCardRow, 0, 2)
	for _, periodStart := range []time.Time{currentDate, prevDate} {
		periodEnd := periodStart.AddDate(0, 1, -1)

		total, err := r.cardStatusTotalFor(ctx, req.CardNumber, "success", periodStart, periodEnd)
		if err != nil {
			return nil, withdraw_errors.ErrGetMonthWithdrawStatusSuccessByCardFailed.WithInternal(err)
		}

		results = append(results, &salemongo.WithdrawMonthlyStatusSuccessByCardRow{
			Year:         yearString(periodStart.Year()),
			Month:        monthName(periodStart.Month()),
			TotalSuccess: total.Count,
			TotalAmount:  int32(total.Amount),
		})
	}
	return results, nil
}

func (r *withdrawStatsByCardStatusRepository) GetMonthWithdrawStatusFailedByCardNumber(ctx context.Context, req *requests.MonthStatusWithdrawCardNumber) ([]*salemongo.WithdrawMonthlyStatusFailedByCardRow, error) {
	currentDate := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC)
	prevDate := currentDate.AddDate(0, -1, 0)

	results := make([]*salemongo.WithdrawMonthlyStatusFailedByCardRow, 0, 2)
	for _, periodStart := range []time.Time{currentDate, prevDate} {
		periodEnd := periodStart.AddDate(0, 1, -1)

		total, err := r.cardStatusTotalFor(ctx, req.CardNumber, "failed", periodStart, periodEnd)
		if err != nil {
			return nil, withdraw_errors.ErrGetMonthWithdrawStatusFailedByCardFailed.WithInternal(err)
		}

		results = append(results, &salemongo.WithdrawMonthlyStatusFailedByCardRow{
			Year:        yearString(periodStart.Year()),
			Month:       monthName(periodStart.Month()),
			TotalFailed: total.Count,
			TotalAmount: int32(total.Amount),
		})
	}
	return results, nil
}

func (r *withdrawStatsByCardStatusRepository) GetYearlyWithdrawStatusSuccessByCardNumber(ctx context.Context, req *requests.YearStatusWithdrawCardNumber) ([]*salemongo.WithdrawYearlyStatusSuccessByCardRow, error) {
	results := make([]*salemongo.WithdrawYearlyStatusSuccessByCardRow, 0, 2)

	for _, year := range []int{req.Year, req.Year - 1} {
		total, err := r.cardStatusTotalFor(ctx, req.CardNumber, "success",
			time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC))
		if err != nil {
			return nil, withdraw_errors.ErrGetYearlyWithdrawStatusSuccessByCardFailed.WithInternal(err)
		}

		results = append(results, &salemongo.WithdrawYearlyStatusSuccessByCardRow{
			Year:         yearString(year),
			TotalSuccess: int32(total.Count),
			TotalAmount:  int32(total.Amount),
		})
	}
	return results, nil
}

func (r *withdrawStatsByCardStatusRepository) GetYearlyWithdrawStatusFailedByCardNumber(ctx context.Context, req *requests.YearStatusWithdrawCardNumber) ([]*salemongo.WithdrawYearlyStatusFailedByCardRow, error) {
	results := make([]*salemongo.WithdrawYearlyStatusFailedByCardRow, 0, 2)

	for _, year := range []int{req.Year, req.Year - 1} {
		total, err := r.cardStatusTotalFor(ctx, req.CardNumber, "failed",
			time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC))
		if err != nil {
			return nil, withdraw_errors.ErrGetYearlyWithdrawStatusFailedByCardFailed.WithInternal(err)
		}

		results = append(results, &salemongo.WithdrawYearlyStatusFailedByCardRow{
			Year:        yearString(year),
			TotalFailed: int32(total.Count),
			TotalAmount: int32(total.Amount),
		})
	}
	return results, nil
}

func (r *withdrawStatsByCardAmountRepository) GetMonthlyWithdrawsByCardNumber(ctx context.Context, req *requests.YearMonthCardNumber) ([]*salemongo.WithdrawMonthlyAmountByCardRow, error) {
	start := time.Date(req.Year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":    nil,
			"card_number":   req.CardNumber,
			"withdraw_time": bson.M{"$gte": start, "$lt": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":                   bson.M{"month": bson.M{"$month": "$withdraw_time"}},
			"total_withdraw_amount": bson.M{"$sum": "$withdraw_amount"},
		}}},
	})
	if err != nil {
		return nil, withdraw_errors.ErrGetMonthlyWithdrawsByCardFailed.WithInternal(err)
	}
	defer cursor.Close(ctx)

	var rows []struct {
		ID struct {
			Month int `bson:"month"`
		} `bson:"_id"`
		TotalWithdrawAmount int64 `bson:"total_withdraw_amount"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, withdraw_errors.ErrGetMonthlyWithdrawsByCardFailed.WithInternal(err)
	}

	totals := make(map[int]int64, len(rows))
	for _, row := range rows {
		totals[row.ID.Month] = row.TotalWithdrawAmount
	}

	results := make([]*salemongo.WithdrawMonthlyAmountByCardRow, 0, 12)
	for month := 1; month <= 12; month++ {
		results = append(results, &salemongo.WithdrawMonthlyAmountByCardRow{
			Month:               monthName(time.Month(month)),
			TotalWithdrawAmount: int32(totals[month]),
		})
	}
	return results, nil
}

func (r *withdrawStatsByCardAmountRepository) GetYearlyWithdrawsByCardNumber(ctx context.Context, req *requests.YearMonthCardNumber) ([]*salemongo.WithdrawYearlyAmountByCardRow, error) {
	start := time.Date(req.Year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(req.Year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	cursor, err := r.coll().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":    nil,
			"card_number":   req.CardNumber,
			"withdraw_time": bson.M{"$gte": start, "$lt": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":                   bson.M{"year": bson.M{"$year": "$withdraw_time"}},
			"total_withdraw_amount": bson.M{"$sum": "$withdraw_amount"},
		}}},
	})
	if err != nil {
		return nil, withdraw_errors.ErrGetYearlyWithdrawsByCardFailed.WithInternal(err)
	}
	defer cursor.Close(ctx)

	var rows []struct {
		ID struct {
			Year int `bson:"year"`
		} `bson:"_id"`
		TotalWithdrawAmount int64 `bson:"total_withdraw_amount"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, withdraw_errors.ErrGetYearlyWithdrawsByCardFailed.WithInternal(err)
	}

	totals := make(map[int]int64, len(rows))
	for _, row := range rows {
		totals[row.ID.Year] = row.TotalWithdrawAmount
	}

	results := make([]*salemongo.WithdrawYearlyAmountByCardRow, 0, 5)
	for year := req.Year - 4; year <= req.Year; year++ {
		results = append(results, &salemongo.WithdrawYearlyAmountByCardRow{
			Year:                yearString(year),
			TotalWithdrawAmount: totals[year],
		})
	}
	return results, nil
}
