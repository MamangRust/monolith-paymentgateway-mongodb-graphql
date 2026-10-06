package seeder

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/date"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-pkg/randomvcc"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// cardSeeder is a struct that represents a seeder for the cards collection.
type cardSeeder struct {
	db     *mongo.Database
	ctx    context.Context
	logger logger.LoggerInterface
}

// NewCardSeeder creates a new instance of the cardSeeder.
func NewCardSeeder(db *mongo.Database, ctx context.Context, logger logger.LoggerInterface) *cardSeeder {
	return &cardSeeder{
		db:     db,
		ctx:    ctx,
		logger: logger,
	}
}

// Seed populates the cards collection with 10 cards, 5 active and 5 trashed.
func (r *cardSeeder) Seed() error {
	cardTypes := []string{"credit", "debit"}
	cardProviders := []string{"mandiri", "bni", "bri"}

	totalCards := 10
	activeCards := 5
	trashedCards := 5

	cardNumbers := make([]string, totalCards)
	for i := 0; i < totalCards; i++ {
		cardNumber, err := randomvcc.RandomCardNumber()
		if err != nil {
			r.logger.Error("failed to generate card number", zap.Int("index", i), zap.Error(err))
			return fmt.Errorf("failed to generate card number: %w", err)
		}
		cardNumbers[i] = cardNumber
	}

	for i := 0; i < totalCards; i++ {
		cardID, err := salemongo.NextID(r.ctx, r.db, "cards")
		if err != nil {
			r.logger.Error("failed to allocate card id", zap.Error(err))
			return fmt.Errorf("failed to allocate card id: %w", err)
		}

		now := nowPtr()
		card := &salemongo.Card{
			CardID:       cardID,
			UserID:       int32(i + 1),
			CardNumber:   cardNumbers[i],
			CardType:     cardTypes[i%len(cardTypes)],
			ExpireDate:   date.GenerateExpireDate(),
			Cvv:          fmt.Sprintf("%03d", i%1000),
			CardProvider: cardProviders[i%len(cardProviders)],
			Status:       "active",
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if err := insertOne(r.ctx, r.db, "cards", card, r.logger); err != nil {
			r.logger.Error("failed to seed card", zap.Int("card", i+1), zap.Error(err))
			return fmt.Errorf("failed to seed card %d: %w", i+1, err)
		}

		if i >= activeCards {
			if err := trashDocument(r.ctx, r.db, "cards", "card_id", card.CardID); err != nil {
				r.logger.Error("failed to trash card", zap.Int("card", i+1), zap.Error(err))
				return fmt.Errorf("failed to trash card %d: %w", i+1, err)
			}
		}
	}

	r.logger.Info("card seeded successfully", zap.Int("totalCards", totalCards), zap.Int("activeCards", activeCards), zap.Int("trashedCards", trashedCards))
	return nil
}
