package repositorydashboard

import (
	"context"

	card_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/card_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type cardDashboardTransactionRepository struct {
	db *mongo.Database
}

func NewCardDashboardTransactionRepository(db *mongo.Database) CardDashboardTransactionRepository {
	return &cardDashboardTransactionRepository{db: db}
}

func (r *cardDashboardTransactionRepository) GetTotalTransactionAmount(ctx context.Context) (*int64, error) {
	total, err := sumAmount(ctx, r.db, "transactions", "amount", nil)
	if err != nil {
		return nil, card_errors.ErrGetTotalTransactionsFailed.WithInternal(err)
	}
	return total, nil
}

func (r *cardDashboardTransactionRepository) GetTotalTransactionAmountByCardNumber(ctx context.Context, cardNumber string) (*int64, error) {
	total, err := sumAmount(ctx, r.db, "transactions", "amount", bson.M{"card_number": cardNumber})
	if err != nil {
		return nil, card_errors.ErrGetTotalTransactionAmountByCardFailed.WithInternal(err)
	}
	return total, nil
}
