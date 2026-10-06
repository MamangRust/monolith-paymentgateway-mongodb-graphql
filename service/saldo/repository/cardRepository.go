package repository

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type cardRepository struct {
	db *mongo.Database
}

func NewCardRepository(db *mongo.Database) CardRepository {
	return &cardRepository{db: db}
}

func (r *cardRepository) FindCardByCardNumber(ctx context.Context, card_number string) (*salemongo.Card, error) {
	var card salemongo.Card
	if err := r.db.Collection("cards").FindOne(ctx, bson.M{
		"card_number": card_number,
		"deleted_at":  nil,
	}).Decode(&card); err != nil {
		return nil, err
	}
	return &card, nil
}
