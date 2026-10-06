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

// withdrawSeeder is a struct that represents a seeder for the withdraws collection.
type withdrawSeeder struct {
	db     *mongo.Database
	ctx    context.Context
	logger logger.LoggerInterface
}

// NewWithdrawSeeder creates a new instance of the withdrawSeeder.
func NewWithdrawSeeder(db *mongo.Database, ctx context.Context, logger logger.LoggerInterface) *withdrawSeeder {
	return &withdrawSeeder{
		db:     db,
		ctx:    ctx,
		logger: logger,
	}
}

// Seed populates the withdraws collection with 10 withdraws, 5 active and
// 5 trashed.
func (r *withdrawSeeder) Seed() error {
	total := 10
	active := 5
	trashed := total - active

	cards := make([]*salemongo.Card, 0, total)
	for i := 1; i <= total; i++ {
		card, found, err := findCardByUserID(r.ctx, r.db, int32(i))
		if err != nil {
			r.logger.Error("failed to get card for user", zap.Int("userID", i), zap.Error(err))
			return fmt.Errorf("failed to get card for user %d: %w", i, err)
		}
		if !found {
			r.logger.Debug("no card found for user", zap.Int("userID", i))
			continue
		}
		cards = append(cards, card)
	}

	if len(cards) < total {
		r.logger.Error("not enough cards for withdraw seeding", zap.Int("required", total), zap.Int("available", len(cards)))
		return fmt.Errorf("not enough cards: required %d, got %d", total, len(cards))
	}

	statusOptions := []string{"pending", "success", "failed"}

	months := make([]time.Time, 12)
	currentYear := time.Now().Year()
	for i := 0; i < 12; i++ {
		months[i] = time.Date(currentYear, time.Month(i+1), 1, 0, 0, 0, 0, time.UTC)
	}

	for i := 0; i < total; i++ {
		card := cards[i]
		status := statusOptions[rand.Intn(len(statusOptions))]

		monthIndex := i % 12
		withdrawTime := months[monthIndex].Add(time.Duration(rand.Intn(28)) * 24 * time.Hour)

		withdrawID, err := salemongo.NextID(r.ctx, r.db, "withdraws")
		if err != nil {
			r.logger.Error("failed to allocate withdraw id", zap.Error(err))
			return fmt.Errorf("failed to allocate withdraw id: %w", err)
		}

		now := nowPtr()
		withdraw := &salemongo.Withdraw{
			WithdrawID:     withdrawID,
			WithdrawNo:     uuid.New(),
			CardNumber:     card.CardNumber,
			WithdrawAmount: int64(rand.Intn(1000000) + 50000),
			WithdrawTime:   withdrawTime,
			Status:         status,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		if err := insertOne(r.ctx, r.db, "withdraws", withdraw, r.logger); err != nil {
			r.logger.Error("failed to seed withdraw", zap.Int("index", i), zap.Error(err))
			return fmt.Errorf("failed to create withdraw %d: %w", i, err)
		}

		if i >= active {
			if err := trashDocument(r.ctx, r.db, "withdraws", "withdraw_id", withdraw.WithdrawID); err != nil {
				r.logger.Error("failed to trash withdraw", zap.Int("withdraw.id", int(withdraw.WithdrawID)), zap.Error(err))
				return fmt.Errorf("failed to trash withdraw %d: %w", i, err)
			}
		}
	}

	r.logger.Info("withdraw seeding completed",
		zap.Int("total", total),
		zap.Int("active", active),
		zap.Int("trashed", trashed))

	return nil
}
