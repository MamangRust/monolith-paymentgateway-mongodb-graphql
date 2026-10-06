package seeder

import (
	"context"
	"fmt"

	apikey "github.com/MamangRust/monolith-payment-gateway-pkg/api-key"
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// merchantSeeder is a struct that represents a seeder for the merchants collection.
type merchantSeeder struct {
	db     *mongo.Database
	ctx    context.Context
	logger logger.LoggerInterface
}

// NewMerchantSeeder creates a new instance of the merchantSeeder.
func NewMerchantSeeder(db *mongo.Database, ctx context.Context, logger logger.LoggerInterface) *merchantSeeder {
	return &merchantSeeder{
		db:     db,
		ctx:    ctx,
		logger: logger,
	}
}

// Seed populates the merchants collection with 10 merchants, 5 active and
// 5 trashed.
func (r *merchantSeeder) Seed() error {
	adjectives := []string{"Blue", "Green", "Red", "Yellow", "Fast"}
	nouns := []string{"Shop", "Store", "Mart", "Market", "Hub"}

	totalMerchants := 10
	activeMerchants := 5
	trashedMerchants := 5

	for i := 0; i < totalMerchants; i++ {
		adjective := adjectives[i%len(adjectives)]
		noun := nouns[i%len(nouns)]
		merchantName := fmt.Sprintf("%s %s", adjective, noun)

		apiKey, _ := apikey.GenerateApiKey()

		merchantID, err := salemongo.NextID(r.ctx, r.db, "merchants")
		if err != nil {
			r.logger.Error("failed to allocate merchant id", zap.Error(err))
			return fmt.Errorf("failed to allocate merchant id: %w", err)
		}

		status := "active"
		if i >= activeMerchants {
			status = "deactive"
		}

		now := nowPtr()
		merchant := &salemongo.Merchant{
			MerchantID: merchantID,
			MerchantNo: uuid.New(),
			Name:       merchantName,
			ApiKey:     apiKey,
			UserID:     int32((i % 5) + 1),
			Status:     status,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		if err := insertOne(r.ctx, r.db, "merchants", merchant, r.logger); err != nil {
			r.logger.Error("failed to seed merchant", zap.Int("merchant", i+1), zap.Error(err))
			return fmt.Errorf("failed to seed merchant %d: %w", i+1, err)
		}

		if status != "active" {
			if err := trashDocument(r.ctx, r.db, "merchants", "merchant_id", merchant.MerchantID); err != nil {
				r.logger.Error("failed to trash merchant", zap.Int("merchant", i+1), zap.Error(err))
				return fmt.Errorf("failed to trash merchant %d: %w", i+1, err)
			}
		}
	}

	r.logger.Info("merchant seeded successfully",
		zap.Int("totalMerchants", totalMerchants),
		zap.Int("activeMerchants", activeMerchants),
		zap.Int("trashedMerchants", trashedMerchants))

	return nil
}
