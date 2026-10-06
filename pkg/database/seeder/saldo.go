package seeder

import (
	"context"
	"fmt"
	"math/rand"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// saldoSeeder is a struct that represents a seeder for the saldos collection.
type saldoSeeder struct {
	db     *mongo.Database
	ctx    context.Context
	logger logger.LoggerInterface
}

// NewSaldoSeeder creates a new instance of the saldoSeeder.
func NewSaldoSeeder(db *mongo.Database, ctx context.Context, logger logger.LoggerInterface) *saldoSeeder {
	return &saldoSeeder{
		db:     db,
		ctx:    ctx,
		logger: logger,
	}
}

// Seed populates the saldos collection with 10 saldos, 5 active and 5 trashed.
func (r *saldoSeeder) Seed() error {
	totalSaldos := 10
	activeSaldos := 5
	trashedSaldos := 5

	cards := make([]*salemongo.Card, 0, totalSaldos)
	for i := 1; i <= totalSaldos; i++ {
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

	if len(cards) < totalSaldos {
		r.logger.Error("not enough cards to seed saldo", zap.Int("required", totalSaldos), zap.Int("available", len(cards)))
		return fmt.Errorf("not enough cards to seed saldo: required %d, got %d", totalSaldos, len(cards))
	}

	for i, card := range cards {
		saldoID, err := salemongo.NextID(r.ctx, r.db, "saldos")
		if err != nil {
			r.logger.Error("failed to allocate saldo id", zap.Error(err))
			return fmt.Errorf("failed to allocate saldo id: %w", err)
		}

		now := nowPtr()
		saldo := &salemongo.Saldo{
			SaldoID:      saldoID,
			CardNumber:   card.CardNumber,
			TotalBalance: int64(rand.Intn(9_000_000) + 1_000_000),
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if err := insertOne(r.ctx, r.db, "saldos", saldo, r.logger); err != nil {
			r.logger.Error("failed to seed saldo", zap.Int("index", i), zap.String("card", card.CardNumber), zap.Error(err))
			return fmt.Errorf("failed to seed saldo for card %s: %w", card.CardNumber, err)
		}

		if i >= activeSaldos {
			if err := trashDocument(r.ctx, r.db, "saldos", "saldo_id", saldo.SaldoID); err != nil {
				r.logger.Error("failed to trash saldo", zap.Int("index", i), zap.String("card", card.CardNumber), zap.Error(err))
				return fmt.Errorf("failed to trash saldo %d for card %s: %w", i+1, card.CardNumber, err)
			}
		}
	}

	r.logger.Info("saldo seeded successfully",
		zap.Int("totalSaldos", totalSaldos),
		zap.Int("activeSaldos", activeSaldos),
		zap.Int("trashedSaldos", trashedSaldos))

	return nil
}
