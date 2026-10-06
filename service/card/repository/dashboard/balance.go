package repositorydashboard

import (
	"context"

	card_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/card_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type cardDashboardBalanceRepository struct {
	db *mongo.Database
}

func NewCardDashboardBalanceRepository(db *mongo.Database) CardDashboardBalanceRepository {
	return &cardDashboardBalanceRepository{db: db}
}

func (r *cardDashboardBalanceRepository) GetTotalBalances(ctx context.Context) (*int64, error) {
	total, err := sumAmount(ctx, r.db, "saldos", "total_balance", nil)
	if err != nil {
		return nil, card_errors.ErrGetTotalBalancesFailed.WithInternal(err)
	}
	return total, nil
}

func (r *cardDashboardBalanceRepository) GetTotalBalanceByCardNumber(ctx context.Context, cardNumber string) (*int64, error) {
	total, err := sumAmount(ctx, r.db, "saldos", "total_balance", bson.M{"card_number": cardNumber})
	if err != nil {
		return nil, card_errors.ErrGetTotalBalanceByCardFailed.WithInternal(err)
	}
	return total, nil
}
