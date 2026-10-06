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

// transferSeeder is a struct that represents a seeder for the transfers collection.
type transferSeeder struct {
	db     *mongo.Database
	ctx    context.Context
	logger logger.LoggerInterface
}

// NewTransferSeeder creates a new instance of the transferSeeder.
func NewTransferSeeder(db *mongo.Database, ctx context.Context, logger logger.LoggerInterface) *transferSeeder {
	return &transferSeeder{
		db:     db,
		ctx:    ctx,
		logger: logger,
	}
}

// Seed populates the transfers collection with 10 transfers, 5 active and
// 5 trashed.
func (r *transferSeeder) Seed() error {
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

	if len(cards) < 2 {
		r.logger.Error("not enough cards available for transfer seeding", zap.Int("available", len(cards)))
		return fmt.Errorf("need at least 2 cards, got %d", len(cards))
	}

	statusOptions := []string{"pending", "success", "failed"}

	months := make([]time.Time, 12)
	currentYear := time.Now().Year()
	for i := 0; i < 12; i++ {
		months[i] = time.Date(currentYear, time.Month(i+1), 1, 0, 0, 0, 0, time.UTC)
	}

	for i := 0; i < total; i++ {
		fromIndex := rand.Intn(len(cards))
		toIndex := rand.Intn(len(cards))
		for fromIndex == toIndex {
			toIndex = rand.Intn(len(cards))
		}

		transferFrom := cards[fromIndex].CardNumber
		transferTo := cards[toIndex].CardNumber
		amount := int64(rand.Intn(1000000) + 50000)
		status := statusOptions[rand.Intn(len(statusOptions))]

		monthIndex := i % 12
		transferTime := months[monthIndex].Add(time.Duration(rand.Intn(28)) * 24 * time.Hour)

		transferID, err := salemongo.NextID(r.ctx, r.db, "transfers")
		if err != nil {
			r.logger.Error("failed to allocate transfer id", zap.Int("transfer", i+1), zap.Error(err))
			return fmt.Errorf("failed to allocate transfer id: %w", err)
		}

		now := nowPtr()
		transfer := &salemongo.Transfer{
			TransferID:     transferID,
			TransferNo:     uuid.New(),
			TransferFrom:   transferFrom,
			TransferTo:     transferTo,
			TransferAmount: amount,
			TransferTime:   transferTime,
			Status:         status,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		if err := insertOne(r.ctx, r.db, "transfers", transfer, r.logger); err != nil {
			r.logger.Error("failed to seed transfer", zap.Int("transfer", i+1), zap.Error(err))
			return fmt.Errorf("failed to seed transfer %d: %w", i+1, err)
		}

		if i >= active {
			if err := trashDocument(r.ctx, r.db, "transfers", "transfer_id", transfer.TransferID); err != nil {
				r.logger.Error("failed to trash transfer", zap.Int("transfer", i+1), zap.Error(err))
				return fmt.Errorf("failed to trash transfer %d: %w", i+1, err)
			}
		}
	}

	r.logger.Info("transfer seeded successfully",
		zap.Int("total", total),
		zap.Int("active", active),
		zap.Int("trashed", trashed))

	return nil
}
