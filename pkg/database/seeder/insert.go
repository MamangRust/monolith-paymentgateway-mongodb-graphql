package seeder

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func nowPtr() *time.Time {
	now := time.Now()
	return &now
}

func insertOne(ctx context.Context, db *mongo.Database, collName string, doc interface{}, log logger.LoggerInterface) error {
	if _, err := db.Collection(collName).InsertOne(ctx, doc); err != nil {
		log.Error("Failed to insert document", zap.String("collection", collName), zap.Error(err))
		return err
	}
	return nil
}

func insertMany(ctx context.Context, db *mongo.Database, collName string, docs []interface{}, log logger.LoggerInterface) error {
	if len(docs) == 0 {
		return nil
	}
	if _, err := db.Collection(collName).InsertMany(ctx, docs); err != nil {
		log.Error("Failed to bulk insert", zap.String("collection", collName), zap.Error(err))
		return err
	}
	log.Info("Seeded collection", zap.String("collection", collName), zap.Int("count", len(docs)))
	return nil
}

// findCardByUserID returns the first active card owned by the given user.
func findCardByUserID(ctx context.Context, db *mongo.Database, userID int32) (*salemongo.Card, bool, error) {
	var result salemongo.Card
	err := db.Collection("cards").FindOne(ctx, bson.M{
		"user_id":    userID,
		"deleted_at": nil,
	}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return &result, true, nil
}

// trashDocument soft-deletes a document by its sequence field.
func trashDocument(ctx context.Context, db *mongo.Database, collName, idField string, id interface{}) error {
	_, err := db.Collection(collName).UpdateOne(ctx,
		bson.M{idField: id, "deleted_at": nil},
		bson.M{"$set": bson.M{"deleted_at": nowPtr()}},
	)
	return err
}

// setStatus updates the status field of a document by its sequence field.
func setStatus(ctx context.Context, db *mongo.Database, collName, idField string, id interface{}, status string) error {
	_, err := db.Collection(collName).UpdateOne(ctx,
		bson.M{idField: id},
		bson.M{"$set": bson.M{"status": status, "updated_at": nowPtr()}},
	)
	return err
}
