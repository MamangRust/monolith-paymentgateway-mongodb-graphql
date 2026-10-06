package outbox

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrInvalidInboxKey = errors.New("invalid consumer inbox key")

const (
	inboxStatusProcessing = "processing"
	inboxStatusProcessed  = "processed"
	inboxStatusPending    = "pending"
)

// ConsumerInbox is the durable deduplication contract used by Kafka handlers.
type ConsumerInbox interface {
	Reserve(ctx context.Context, consumerName, eventKey, topic string, partition int32, offset int64) (bool, bool, int64, error)
	MarkProcessed(ctx context.Context, consumerName, eventKey string, reservationVersion int64) error
	Release(ctx context.Context, consumerName, eventKey string, reservationVersion int64, processingErr error) error
}

type consumerInboxRow struct {
	ConsumerName       string     `bson:"consumer_name"`
	EventKey           string     `bson:"event_key"`
	Topic              string     `bson:"topic"`
	PartitionID        int32      `bson:"partition_id"`
	MessageOffset      int64      `bson:"message_offset"`
	Status             string     `bson:"status"`
	Attempts           int32      `bson:"attempts"`
	LeaseUntil         time.Time  `bson:"lease_until"`
	LastError          string     `bson:"last_error"`
	ProcessedAt        *time.Time `bson:"processed_at"`
	ReservationVersion int64      `bson:"reservation_version"`
}

// Reserve claims an event for a consumer using FindOneAndUpdate (MongoDB idempotent upsert).
func Reserve(ctx context.Context, db *mongo.Database, consumerName, eventKey, topic string, partition int32, offset int64) (bool, bool, int64, error) {
	if db == nil || consumerName == "" || eventKey == "" {
		return false, false, 0, ErrInvalidInboxKey
	}

	coll := db.Collection("consumer_inbox")
	filter := bson.M{"consumer_name": consumerName, "event_key": eventKey}

	// Try to insert a new document
	now := time.Now()
	doc := bson.M{
		"consumer_name":       consumerName,
		"event_key":           eventKey,
		"topic":               topic,
		"partition_id":        partition,
		"message_offset":      offset,
		"status":              inboxStatusProcessing,
		"attempts":            int32(1),
		"reservation_version": int64(1),
		"lease_until":         now.Add(1 * time.Minute),
		"last_error":          "",
		"processed_at":        nil,
	}

	opts := options.Update().SetUpsert(true)
	update := bson.M{
		"$setOnInsert": doc,
		"$set": bson.M{
			"status":         inboxStatusProcessing,
			"last_error":     "",
			"lease_until":    now.Add(1 * time.Minute),
			"topic":          topic,
			"partition_id":   partition,
			"message_offset": offset,
		},
		"$inc": bson.M{
			"attempts":            int32(1),
			"reservation_version": int64(1),
		},
	}

	result, err := coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return false, false, 0, err
	}

	if result.MatchedCount > 0 && result.UpsertedCount == 0 {
		// Document existed — check if it's processed
		var row consumerInboxRow
		err = coll.FindOne(ctx, bson.M{
			"consumer_name": consumerName,
			"event_key":     eventKey,
		}).Decode(&row)
		if err != nil {
			return false, false, 0, err
		}
		return false, row.Status == inboxStatusProcessed, row.ReservationVersion, nil
	}

	// Fresh insert (upserted)
	return true, false, 1, nil
}

// MarkProcessed completes only the active reservation.
func MarkProcessed(ctx context.Context, db *mongo.Database, consumerName, eventKey string, reservationVersion int64) error {
	if db == nil || consumerName == "" || eventKey == "" {
		return ErrInvalidInboxKey
	}
	coll := db.Collection("consumer_inbox")
	_, err := coll.UpdateOne(ctx, bson.M{
		"consumer_name":       consumerName,
		"event_key":           eventKey,
		"status":              inboxStatusProcessing,
		"reservation_version": reservationVersion,
	}, bson.M{
		"$set": bson.M{
			"status":       inboxStatusProcessed,
			"processed_at": time.Now(),
			"lease_until":  time.Now(),
			"last_error":   "",
		},
	})
	return err
}

// Release releases only the active reservation so the retry path can re-claim the event.
func Release(ctx context.Context, db *mongo.Database, consumerName, eventKey string, reservationVersion int64, processingErr error) error {
	if db == nil || consumerName == "" || eventKey == "" {
		return ErrInvalidInboxKey
	}
	lastError := "consumer processing failed"
	if processingErr != nil {
		lastError = processingErr.Error()
	}
	coll := db.Collection("consumer_inbox")
	_, err := coll.UpdateOne(ctx, bson.M{
		"consumer_name":       consumerName,
		"event_key":           eventKey,
		"status":              inboxStatusProcessing,
		"reservation_version": reservationVersion,
	}, bson.M{
		"$set": bson.M{
			"status":      inboxStatusPending,
			"lease_until": time.Now(),
			"last_error":  lastError,
		},
	})
	return err
}

// MongoInbox adapts a MongoDB connection to ConsumerInbox.
type MongoInbox struct {
	db *mongo.Database
}

func NewMongoInbox(db *mongo.Database) (*MongoInbox, error) {
	if db == nil {
		return nil, errors.New("inbox db is nil")
	}
	return &MongoInbox{db: db}, nil
}

func (i *MongoInbox) Reserve(ctx context.Context, consumerName, eventKey, topic string, partition int32, offset int64) (bool, bool, int64, error) {
	return Reserve(ctx, i.db, consumerName, eventKey, topic, partition, offset)
}

func (i *MongoInbox) MarkProcessed(ctx context.Context, consumerName, eventKey string, reservationVersion int64) error {
	return MarkProcessed(ctx, i.db, consumerName, eventKey, reservationVersion)
}

func (i *MongoInbox) Release(ctx context.Context, consumerName, eventKey string, reservationVersion int64, processingErr error) error {
	return Release(ctx, i.db, consumerName, eventKey, reservationVersion, processingErr)
}
