package saldostatsrepository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	saldo_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/saldo_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type saldoStatsBalanceRepository struct {
	db *mongo.Database
}

func NewSaldoStatsBalanceRepository(db *mongo.Database) SaldoStatsBalanceRepository {
	return &saldoStatsBalanceRepository{db: db}
}

func monthLabel(m time.Month) string {
	return m.String()[:3]
}

func (r *saldoStatsBalanceRepository) monthlyBalances(ctx context.Context, start, end time.Time) (map[int]int64, error) {
	cursor, err := r.db.Collection("saldos").Aggregate(ctx, salemongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at": nil,
			"created_at": bson.M{"$gte": start, "$lt": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":           bson.M{"month": bson.M{"$month": "$created_at"}},
			"total_balance": bson.M{"$sum": "$total_balance"},
		}}},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	totals := make(map[int]int64, 12)
	for cursor.Next(ctx) {
		var item struct {
			ID struct {
				Month int `bson:"month"`
			} `bson:"_id"`
			Total int64 `bson:"total_balance"`
		}
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		totals[item.ID.Month] = item.Total
	}
	return totals, nil
}

func (r *saldoStatsBalanceRepository) GetMonthlySaldoBalances(ctx context.Context, year int) ([]*salemongo.MonthlySaldoBalanceRow, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	totals, err := r.monthlyBalances(ctx, start, end)
	if err != nil {
		return nil, saldo_errors.ErrGetMonthlySaldoBalancesFailed.WithInternal(err)
	}

	results := make([]*salemongo.MonthlySaldoBalanceRow, 0, 12)
	for month := 1; month <= 12; month++ {
		results = append(results, &salemongo.MonthlySaldoBalanceRow{
			Month:        monthLabel(time.Month(month)),
			TotalBalance: int32(totals[month]),
		})
	}
	return results, nil
}

func (r *saldoStatsBalanceRepository) GetYearlySaldoBalances(ctx context.Context, year int) ([]*salemongo.YearlySaldoBalanceRow, error) {
	start := time.Date(year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	cursor, err := r.db.Collection("saldos").Aggregate(ctx, salemongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at": nil,
			"created_at": bson.M{"$gte": start, "$lt": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":           bson.M{"year": bson.M{"$year": "$created_at"}},
			"total_balance": bson.M{"$sum": "$total_balance"},
		}}},
	})
	if err != nil {
		return nil, saldo_errors.ErrGetYearlySaldoBalancesFailed.WithInternal(err)
	}
	defer cursor.Close(ctx)

	totals := make(map[int]int64, 5)
	for cursor.Next(ctx) {
		var item struct {
			ID struct {
				Year int `bson:"year"`
			} `bson:"_id"`
			Total int64 `bson:"total_balance"`
		}
		if err := cursor.Decode(&item); err != nil {
			return nil, saldo_errors.ErrGetYearlySaldoBalancesFailed.WithInternal(err)
		}
		totals[item.ID.Year] = item.Total
	}

	results := make([]*salemongo.YearlySaldoBalanceRow, 0, 5)
	for offset := 4; offset >= 0; offset-- {
		current := year - offset
		results = append(results, &salemongo.YearlySaldoBalanceRow{
			Year:         time.Date(current, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006"),
			TotalBalance: totals[current],
		})
	}
	return results, nil
}
