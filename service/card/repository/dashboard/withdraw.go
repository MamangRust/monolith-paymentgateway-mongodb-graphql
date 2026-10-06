package repositorydashboard

import (
	"context"

	card_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/card_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type cardDashboardWithdrawRepository struct {
	db *mongo.Database
}

func NewCardDashboardWithdrawRepository(db *mongo.Database) CardDashboardWithdrawRepository {
	return &cardDashboardWithdrawRepository{db: db}
}

func (r *cardDashboardWithdrawRepository) GetTotalWithdrawAmount(ctx context.Context) (*int64, error) {
	total, err := sumAmount(ctx, r.db, "withdraws", "withdraw_amount", nil)
	if err != nil {
		return nil, card_errors.ErrGetTotalWithdrawsFailed.WithInternal(err)
	}
	return total, nil
}

func (r *cardDashboardWithdrawRepository) GetTotalWithdrawAmountByCardNumber(ctx context.Context, cardNumber string) (*int64, error) {
	total, err := sumAmount(ctx, r.db, "withdraws", "withdraw_amount", bson.M{"card_number": cardNumber})
	if err != nil {
		return nil, card_errors.ErrGetTotalWithdrawAmountByCardFailed.WithInternal(err)
	}
	return total, nil
}
