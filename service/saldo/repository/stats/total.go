package saldostatsrepository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	saldo_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/saldo_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type saldoStatsTotalBalanceRepository struct {
	db *mongo.Database
}

func NewSaldoStatsTotalBalanceRepository(db *mongo.Database) SaldoStatsTotalSaldoRepository {
	return &saldoStatsTotalBalanceRepository{db: db}
}

func (r *saldoStatsTotalBalanceRepository) totalBetween(ctx context.Context, start, end time.Time) (int64, error) {
	cursor, err := r.db.Collection("saldos").Aggregate(ctx, salemongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at": nil,
			"created_at": bson.M{"$gte": start, "$lt": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":           nil,
			"total_balance": bson.M{"$sum": "$total_balance"},
		}}},
	})
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var row struct {
		TotalBalance int64 `bson:"total_balance"`
	}
	if cursor.Next(ctx) {
		if err := cursor.Decode(&row); err != nil {
			return 0, err
		}
	}
	return row.TotalBalance, nil
}

func (r *saldoStatsTotalBalanceRepository) GetMonthlyTotalSaldoBalance(ctx context.Context, req *requests.MonthTotalSaldoBalance) ([]*salemongo.MonthlyTotalSaldoBalanceRow, error) {
	results := make([]*salemongo.MonthlyTotalSaldoBalanceRow, 0, 2)

	for i := 0; i < 2; i++ {
		start := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -i, 0)
		end := start.AddDate(0, 1, 0)

		total, err := r.totalBetween(ctx, start, end)
		if err != nil {
			return nil, saldo_errors.ErrGetMonthlyTotalSaldoBalanceFailed.WithInternal(err)
		}

		results = append(results, &salemongo.MonthlyTotalSaldoBalanceRow{
			Year:         start.Format("2006"),
			Month:        monthLabel(start.Month()),
			TotalBalance: int32(total),
		})
	}

	return results, nil
}

func (r *saldoStatsTotalBalanceRepository) GetYearTotalSaldoBalance(ctx context.Context, year int) ([]*salemongo.YearlyTotalSaldoBalancesRow, error) {
	results := make([]*salemongo.YearlyTotalSaldoBalancesRow, 0, 2)

	for i := 0; i < 2; i++ {
		start := time.Date(year-i, 1, 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(1, 0, 0)

		total, err := r.totalBetween(ctx, start, end)
		if err != nil {
			return nil, saldo_errors.ErrGetYearTotalSaldoBalanceFailed.WithInternal(err)
		}

		results = append(results, &salemongo.YearlyTotalSaldoBalancesRow{
			Year:         start.Format("2006"),
			TotalBalance: int32(total),
		})
	}

	return results, nil
}
