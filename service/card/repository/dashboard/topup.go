package repositorydashboard

import (
	"context"

	card_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/card_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type cardDashboardTopupRepository struct {
	db *mongo.Database
}

func NewCardDashboardTopupRepository(db *mongo.Database) CardDashboardTopupRepository {
	return &cardDashboardTopupRepository{db: db}
}

func (r *cardDashboardTopupRepository) GetTotalTopAmount(ctx context.Context) (*int64, error) {
	total, err := sumAmount(ctx, r.db, "topups", "topup_amount", nil)
	if err != nil {
		return nil, card_errors.ErrGetTotalTopupsFailed.WithInternal(err)
	}
	return total, nil
}

func (r *cardDashboardTopupRepository) GetTotalTopupAmountByCardNumber(ctx context.Context, cardNumber string) (*int64, error) {
	total, err := sumAmount(ctx, r.db, "topups", "topup_amount", bson.M{"card_number": cardNumber})
	if err != nil {
		return nil, card_errors.ErrGetTotalTopupAmountByCardFailed.WithInternal(err)
	}
	return total, nil
}
