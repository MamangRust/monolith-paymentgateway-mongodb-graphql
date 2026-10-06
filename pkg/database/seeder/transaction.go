package seeder

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// transactionSeeder is a struct that represents a seeder for the transactions collection.
type transactionSeeder struct {
	db     *mongo.Database
	ctx    context.Context
	logger logger.LoggerInterface
}

// NewTransactionSeeder creates a new instance of the transactionSeeder.
func NewTransactionSeeder(db *mongo.Database, ctx context.Context, logger logger.LoggerInterface) *transactionSeeder {
	return &transactionSeeder{
		db:     db,
		ctx:    ctx,
		logger: logger,
	}
}

// Seed populates the transactions collection with 10 transactions, 5 active
// and 5 trashed.
func (r *transactionSeeder) Seed() error {
	total := 10
	active := 5
	trashed := 5

	paymentMethods := []string{"Bank Alpha", "Bank Beta", "Bank Gamma"}
	statusOptions := []string{"pending", "success", "failed"}

	cards := make([]*salemongo.Card, 0, total)
	for i := 1; i <= total; i++ {
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

	if len(cards) < total {
		r.logger.Error("not enough cards for transaction seeding", zap.Int("required", total), zap.Int("available", len(cards)))
		return fmt.Errorf("not enough cards for transaction seeding: required %d, got %d", total, len(cards))
	}

	cursor, err := r.db.Collection("merchants").Find(r.ctx,
		bson.M{"deleted_at": nil},
		options.Find().SetLimit(int64(total)))
	if err != nil {
		r.logger.Error("failed to get merchant list", zap.Error(err))
		return fmt.Errorf("failed to get merchant list: %w", err)
	}
	defer cursor.Close(r.ctx)

	var merchants []salemongo.Merchant
	if err := cursor.All(r.ctx, &merchants); err != nil {
		r.logger.Error("failed to decode merchant list", zap.Error(err))
		return fmt.Errorf("failed to decode merchant list: %w", err)
	}

	if len(merchants) < total {
		r.logger.Error("not enough merchants for transaction seeding", zap.Int("required", total), zap.Int("available", len(merchants)))
		return fmt.Errorf("not enough merchants: required %d, got %d", total, len(merchants))
	}

	months := make([]time.Time, 12)
	currentYear := time.Now().Year()
	for i := 0; i < 12; i++ {
		months[i] = time.Date(currentYear, time.Month(i+1), 1, 0, 0, 0, 0, time.UTC)
	}

	for i := 0; i < total; i++ {
		card := cards[i]
		merchant := merchants[i%len(merchants)]
		paymentMethod := paymentMethods[i%len(paymentMethods)]
		status := statusOptions[i%len(statusOptions)]

		monthIndex := i % 12
		transactionTime := months[monthIndex].Add(time.Duration(rand.Intn(28)) * 24 * time.Hour)

		transactionID, err := salemongo.NextID(r.ctx, r.db, "transactions")
		if err != nil {
			r.logger.Error("failed to allocate transaction id", zap.Int("index", i), zap.Error(err))
			return fmt.Errorf("failed to allocate transaction id: %w", err)
		}

		now := nowPtr()
		transaction := &salemongo.Transaction{
			TransactionID:   transactionID,
			TransactionNo:   uuid.New(),
			CardNumber:      card.CardNumber,
			Amount:          int64(rand.Intn(1000000-50000) + 50000),
			PaymentMethod:   paymentMethod,
			MerchantID:      merchant.MerchantID,
			TransactionTime: transactionTime,
			Status:          status,
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		if err := insertOne(r.ctx, r.db, "transactions", transaction, r.logger); err != nil {
			r.logger.Error("failed to seed transaction", zap.Int("index", i), zap.Error(err))
			return fmt.Errorf("failed to seed transaction %d: %w", i, err)
		}

		if i >= active {
			if err := trashDocument(r.ctx, r.db, "transactions", "transaction_id", transaction.TransactionID); err != nil {
				r.logger.Error("failed to trash transaction", zap.Int("index", i), zap.Error(err))
				return fmt.Errorf("failed to trash transaction %d: %w", i, err)
			}
		}
	}

	r.logger.Info("transaction seeded successfully", zap.Int("total", total), zap.Int("active", active), zap.Int("trashed", trashed))
	return nil
}
