package repositorydashboard

import (
	"context"

	card_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/card_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type cardDashboardTransferRepository struct {
	db *mongo.Database
}

func NewCardDashboardTransferRepository(db *mongo.Database) CardDashboardTransferRepository {
	return &cardDashboardTransferRepository{db: db}
}

func (r *cardDashboardTransferRepository) GetTotalTransferAmount(ctx context.Context) (*int64, error) {
	total, err := sumAmount(ctx, r.db, "transfers", "transfer_amount", nil)
	if err != nil {
		return nil, card_errors.ErrGetTotalTransfersFailed.WithInternal(err)
	}
	return total, nil
}

func (r *cardDashboardTransferRepository) GetTotalTransferAmountBySender(ctx context.Context, senderCardNumber string) (*int64, error) {
	total, err := sumAmount(ctx, r.db, "transfers", "transfer_amount", bson.M{"transfer_from": senderCardNumber})
	if err != nil {
		return nil, card_errors.ErrGetTotalTransferSenderByCardFailed.WithInternal(err)
	}
	return total, nil
}

func (r *cardDashboardTransferRepository) GetTotalTransferAmountByReceiver(ctx context.Context, receiverCardNumber string) (*int64, error) {
	total, err := sumAmount(ctx, r.db, "transfers", "transfer_amount", bson.M{"transfer_to": receiverCardNumber})
	if err != nil {
		return nil, card_errors.ErrGetTotalTransferSenderByCardFailed.WithInternal(err)
	}
	return total, nil
}
