package seeder

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// topupSeeder is a struct that represents a seeder for the topups collection.
type topupSeeder struct {
	db     *mongo.Database
	ctx    context.Context
	logger logger.LoggerInterface
}

// NewTopupSeeder creates a new instance of the topupSeeder.
func NewTopupSeeder(db *mongo.Database, ctx context.Context, logger logger.LoggerInterface) *topupSeeder {
	return &topupSeeder{
		db:     db,
		ctx:    ctx,
		logger: logger,
	}
}

// Seed populates the topups collection with 10 topups, 5 active and 5 trashed.
func (r *topupSeeder) Seed() error {
	totalTopups := 10
	activeTopups := 5
	trashedTopups := 5

	cards := make([]*salemongo.Card, 0, totalTopups)
	for i := 1; i <= totalTopups; i++ {
		card, found, err := findCardByUserID(r.ctx, r.db, int32(i))
		if err != nil {
			r.logger.Error("failed to get card for user", zap.Int("userID", i), zap.Error(err))
			return fmt.Errorf("failed to get card for user %d: %w", i, err)
		}
		if !found {
			r.logger.Error("no card found for user", zap.Int("userID", i))
			continue
		}
		cards = append(cards, card)
	}

	if len(cards) < totalTopups {
		r.logger.Error("not enough cards found to seed topups", zap.Int("found", len(cards)))
		return fmt.Errorf("not enough cards to seed topups")
	}

	topupMethods := []string{"Bank Alpha", "Bank Beta", "Bank Gamma"}
	statusOptions := []string{"pending", "success", "failed"}

	months := make([]time.Time, 12)
	currentYear := time.Now().Year()
	for i := 0; i < 12; i++ {
		months[i] = time.Date(currentYear, time.Month(i+1), 1, 0, 0, 0, 0, time.UTC)
	}

	for i := 0; i < totalTopups; i++ {
		card := cards[i]
		cardNumber := card.CardNumber

		monthIndex := i % 12
		topupTime := months[monthIndex].Add(time.Duration(rand.Intn(28)) * 24 * time.Hour)

		topupID, err := salemongo.NextID(r.ctx, r.db, "topups")
		if err != nil {
			r.logger.Error("failed to allocate topup id", zap.String("card", cardNumber), zap.Error(err))
			return fmt.Errorf("failed to allocate topup id: %w", err)
		}

		status := statusOptions[rand.Intn(len(statusOptions))]
		now := nowPtr()
		topup := &salemongo.Topup{
			TopupID:     topupID,
			TopupNo:     uuid.New(),
			CardNumber:  cardNumber,
			TopupAmount: int64(rand.Intn(10000000) + 1000000),
			TopupMethod: topupMethods[rand.Intn(len(topupMethods))],
			TopupTime:   topupTime,
			Status:      status,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if err := insertOne(r.ctx, r.db, "topups", topup, r.logger); err != nil {
			r.logger.Error("failed to seed topup", zap.String("card", cardNumber), zap.Error(err))
			return fmt.Errorf("failed to seed topup for card %s: %w", cardNumber, err)
		}

		if i >= activeTopups {
			if err := trashDocument(r.ctx, r.db, "topups", "topup_id", topup.TopupID); err != nil {
				r.logger.Error("failed to trash topup", zap.Int("topup", i+1), zap.String("card", cardNumber), zap.Error(err))
				return fmt.Errorf("failed to trash topup %d for card %s: %w", i+1, cardNumber, err)
			}
		}
	}

	r.logger.Info("topup seeded successfully", zap.Int("totalTopups", totalTopups), zap.Int("activeTopups", activeTopups), zap.Int("trashedTopups", trashedTopups))
	return nil
}
