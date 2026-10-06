package transferstatsbycardrepository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	transfer_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/transfer_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type transferStatsByCardAmountReceiverRepository struct {
	db *mongo.Database
}

func NewTransferStatsByCardAmountReceiverRepository(db *mongo.Database) TransferStatsByCardAmountReceiverRepository {
	return &transferStatsByCardAmountReceiverRepository{db: db}
}

func timeLabel(m time.Month) string {
	return m.String()[:3]
}

func (r *transferStatsByCardAmountReceiverRepository) GetMonthlyTransferAmountsByReceiverCardNumber(ctx context.Context, req *requests.MonthYearCardNumber) ([]*salemongo.TransferMonthlyAmountByReceiverCardRow, error) {
	start := time.Date(req.Year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(req.Year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	cursor, err := r.db.Collection("transfers").Aggregate(ctx, salemongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":    nil,
			"transfer_to":   req.CardNumber,
			"transfer_time": bson.M{"$gte": start, "$lt": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":                   bson.M{"month": bson.M{"$month": "$transfer_time"}},
			"total_transfer_amount": bson.M{"$sum": "$transfer_amount"},
		}}},
	})
	if err != nil {
		return nil, transfer_errors.ErrGetMonthlyTransferAmountsByReceiverCardFailed.WithInternal(err)
	}
	defer cursor.Close(ctx)

	totals := make(map[int]int32, 12)
	for cursor.Next(ctx) {
		var item struct {
			ID struct {
				Month int `bson:"month"`
			} `bson:"_id"`
			Total int64 `bson:"total_transfer_amount"`
		}
		if err := cursor.Decode(&item); err != nil {
			return nil, transfer_errors.ErrGetMonthlyTransferAmountsByReceiverCardFailed.WithInternal(err)
		}
		totals[item.ID.Month] = int32(item.Total)
	}

	results := make([]*salemongo.TransferMonthlyAmountByReceiverCardRow, 0, 12)
	for month := 1; month <= 12; month++ {
		results = append(results, &salemongo.TransferMonthlyAmountByReceiverCardRow{
			Month:               timeLabel(time.Month(month)),
			TotalTransferAmount: totals[month],
		})
	}
	return results, nil
}

func (r *transferStatsByCardAmountReceiverRepository) GetYearlyTransferAmountsByReceiverCardNumber(ctx context.Context, req *requests.MonthYearCardNumber) ([]*salemongo.TransferYearlyAmountByReceiverCardRow, error) {
	start := time.Date(req.Year-4, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(req.Year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	cursor, err := r.db.Collection("transfers").Aggregate(ctx, salemongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at":    nil,
			"transfer_to":   req.CardNumber,
			"transfer_time": bson.M{"$gte": start, "$lt": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":                   bson.M{"year": bson.M{"$year": "$transfer_time"}},
			"total_transfer_amount": bson.M{"$sum": "$transfer_amount"},
		}}},
		{{Key: "$sort", Value: bson.M{"_id.year": 1}}},
	})
	if err != nil {
		return nil, transfer_errors.ErrGetYearlyTransferAmountsByReceiverCardFailed.WithInternal(err)
	}
	defer cursor.Close(ctx)

	results := make([]*salemongo.TransferYearlyAmountByReceiverCardRow, 0, 5)
	for cursor.Next(ctx) {
		var item struct {
			ID struct {
				Year int `bson:"year"`
			} `bson:"_id"`
			Total int64 `bson:"total_transfer_amount"`
		}
		if err := cursor.Decode(&item); err != nil {
			return nil, transfer_errors.ErrGetYearlyTransferAmountsByReceiverCardFailed.WithInternal(err)
		}
		results = append(results, &salemongo.TransferYearlyAmountByReceiverCardRow{
			Year:                time.Date(item.ID.Year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006"),
			TotalTransferAmount: item.Total,
		})
	}
	return results, nil
}
