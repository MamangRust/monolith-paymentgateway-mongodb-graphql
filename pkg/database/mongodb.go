package database

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// NewMongoDatabase opens the MongoDB connection used by every service and
// makes sure the collection indexes exist.
func NewMongoDatabase(log logger.LoggerInterface) (*mongo.Database, error) {
	client, err := salemongo.NewMongoClient(log)
	if err != nil {
		return nil, err
	}

	db := salemongo.NewDatabase(client)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := salemongo.CreateIndexes(ctx, db); err != nil {
		log.Error("Failed to create MongoDB indexes", zap.Error(err))
		return nil, err
	}

	return db, nil
}
