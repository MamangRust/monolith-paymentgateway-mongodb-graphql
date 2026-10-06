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

type transferStatsByCardStatusRepository struct {
	db *mongo.Database
}

func NewTransferStatsByCardStatusRepository(db *mongo.Database) TransferStatsByCardStatusRepository {
	return &transferStatsByCardStatusRepository{db: db}
}

// cardStatusTotals sums the transfer count and amount for a status window,
// counting the card both as sender and receiver.
func (r *transferStatsByCardStatusRepository) cardStatusTotals(ctx context.Context, cardNumber, status string, start, end time.Time) (int64, int64, error) {
	cursor, err := r.db.Collection("transfers").Aggregate(ctx, salemongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"deleted_at": nil,
			"status":     status,
			"$or": []bson.M{
				{"transfer_from": cardNumber},
				{"transfer_to": cardNumber},
			},
			"transfer_time": bson.M{"$gte": start, "$lt": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":          nil,
			"total_count":  bson.M{"$sum": 1},
			"total_amount": bson.M{"$sum": "$transfer_amount"},
		}}},
	})
	if err != nil {
		return 0, 0, err
	}
	defer cursor.Close(ctx)

	var row struct {
		TotalCount  int64 `bson:"total_count"`
		TotalAmount int64 `bson:"total_amount"`
	}
	if cursor.Next(ctx) {
		if err := cursor.Decode(&row); err != nil {
			return 0, 0, err
		}
	}
	return row.TotalCount, row.TotalAmount, nil
}

func (r *transferStatsByCardStatusRepository) GetMonthTransferStatusSuccessByCardNumber(ctx context.Context, req *requests.MonthStatusTransferCardNumber) ([]*salemongo.TransferMonthlyStatusSuccessByCardRow, error) {
	results := make([]*salemongo.TransferMonthlyStatusSuccessByCardRow, 0, 2)

	for i := 0; i < 2; i++ {
		start := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -i, 0)
		end := start.AddDate(0, 1, 0)

		total, amount, err := r.cardStatusTotals(ctx, req.CardNumber, "success", start, end)
		if err != nil {
			return nil, transfer_errors.ErrGetMonthTransferStatusSuccessByCardFailed.WithInternal(err)
		}

		results = append(results, &salemongo.TransferMonthlyStatusSuccessByCardRow{
			Year:         start.Format("2006"),
			Month:        monthLabel(start.Month()),
			TotalSuccess: total,
			TotalAmount:  int32(amount),
		})
	}

	return results, nil
}

func (r *transferStatsByCardStatusRepository) GetYearlyTransferStatusSuccessByCardNumber(ctx context.Context, req *requests.YearStatusTransferCardNumber) ([]*salemongo.TransferYearlyStatusSuccessByCardRow, error) {
	results := make([]*salemongo.TransferYearlyStatusSuccessByCardRow, 0, 2)

	for i := 0; i < 2; i++ {
		start := time.Date(req.Year-i, 1, 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(1, 0, 0)

		total, amount, err := r.cardStatusTotals(ctx, req.CardNumber, "success", start, end)
		if err != nil {
			return nil, transfer_errors.ErrGetYearlyTransferStatusSuccessByCardFailed.WithInternal(err)
		}

		results = append(results, &salemongo.TransferYearlyStatusSuccessByCardRow{
			Year:         start.Format("2006"),
			TotalSuccess: int32(total),
			TotalAmount:  int32(amount),
		})
	}

	return results, nil
}

func (r *transferStatsByCardStatusRepository) GetMonthTransferStatusFailedByCardNumber(ctx context.Context, req *requests.MonthStatusTransferCardNumber) ([]*salemongo.TransferMonthlyStatusFailedByCardRow, error) {
	results := make([]*salemongo.TransferMonthlyStatusFailedByCardRow, 0, 2)

	for i := 0; i < 2; i++ {
		start := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -i, 0)
		end := start.AddDate(0, 1, 0)

		total, amount, err := r.cardStatusTotals(ctx, req.CardNumber, "failed", start, end)
		if err != nil {
			return nil, transfer_errors.ErrGetMonthTransferStatusFailedByCardFailed.WithInternal(err)
		}

		results = append(results, &salemongo.TransferMonthlyStatusFailedByCardRow{
			Year:        start.Format("2006"),
			Month:       monthLabel(start.Month()),
			TotalFailed: total,
			TotalAmount: int32(amount),
		})
	}

	return results, nil
}

func (r *transferStatsByCardStatusRepository) GetYearlyTransferStatusFailedByCardNumber(ctx context.Context, req *requests.YearStatusTransferCardNumber) ([]*salemongo.TransferYearlyStatusFailedByCardRow, error) {
	results := make([]*salemongo.TransferYearlyStatusFailedByCardRow, 0, 2)

	for i := 0; i < 2; i++ {
		start := time.Date(req.Year-i, 1, 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(1, 0, 0)

		total, amount, err := r.cardStatusTotals(ctx, req.CardNumber, "failed", start, end)
		if err != nil {
			return nil, transfer_errors.ErrGetYearlyTransferStatusFailedByCardFailed.WithInternal(err)
		}

		results = append(results, &salemongo.TransferYearlyStatusFailedByCardRow{
			Year:        start.Format("2006"),
			TotalFailed: int32(total),
			TotalAmount: int32(amount),
		})
	}

	return results, nil
}
