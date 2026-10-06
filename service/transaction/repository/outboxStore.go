package repository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// OutboxRecord is one entry of the transactional outbox, written in the same
// unit of work as the aggregate change it describes.
type OutboxRecord struct {
	AggregateType string
	AggregateID   string
	EventType     string
	Payload       []byte
}

type outboxMongoStore struct {
	coll *mongo.Collection
}

// NewOutboxMongoStore returns an outbox writer backed by the outbox_events
// collection.
func NewOutboxMongoStore(coll *mongo.Collection) *outboxMongoStore {
	return &outboxMongoStore{coll: coll}
}

func (s *outboxMongoStore) Insert(ctx context.Context, record OutboxRecord) error {
	now := time.Now()

	id, err := salemongo.NextID(ctx, s.coll.Database(), "outbox_events")
	if err != nil {
		return err
	}

	_, err = s.coll.InsertOne(ctx, bson.M{
		"id":              int64(id),
		"event_key":       uuid.NewString(),
		"topic":           record.EventType,
		"message_key":     record.AggregateID,
		"payload":         record.Payload,
		"status":          "pending",
		"attempts":        int32(0),
		"next_attempt_at": now,
		"last_error":      "",
		"claim_version":   int64(0),
		"created_at":      now,
		"updated_at":      now,
	})
	return err
}
