package outbox

import (
	"context"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const (
	OutboxMaxAttempts    = 5
	OutboxBackoff        = 30 * time.Second
	OutboxRelayInterval  = 5 * time.Second
	OutboxRelayBatchSize = 100
	OutboxClaimLease     = 1 * time.Minute
	OutboxRetention      = 7 * 24 * time.Hour
)

type OutboxPublisher interface {
	SendMessage(ctx context.Context, topic string, key string, value []byte) error
}

type OutboxEvent struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	OutboxID      int64              `bson:"outbox_id"`
	Topic         string             `bson:"topic"`
	EventKey      string             `bson:"event_key"`
	Payload       []byte             `bson:"payload"`
	Status        string             `bson:"status"`
	Attempts      int32              `bson:"attempts"`
	NextAttemptAt time.Time          `bson:"next_attempt_at"`
	CreatedAt     time.Time          `bson:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at"`
}

func (OutboxEvent) CollectionName() string { return "outbox_events" }

type OutboxService struct {
	coll      *mongo.Collection
	publisher OutboxPublisher
	logger    logger.LoggerInterface
}

func NewOutboxService(db *mongo.Database, publisher OutboxPublisher, log logger.LoggerInterface) *OutboxService {
	return &OutboxService{
		coll:      db.Collection("outbox_events"),
		publisher: publisher,
		logger:    log,
	}
}

func (s *OutboxService) Enqueue(ctx context.Context, topic, key string, payload []byte) error {
	if s.coll == nil {
		return nil
	}
	now := time.Now()
	doc := OutboxEvent{
		Topic:         topic,
		EventKey:      key,
		Payload:       payload,
		Status:        "pending",
		NextAttemptAt: now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	_, err := s.coll.InsertOne(ctx, doc)
	if err != nil {
		return err
	}
	s.logger.Info("outbox event enqueued", zap.String("topic", topic), zap.String("key", key))
	return nil
}

func (s *OutboxService) EnqueueInTx(ctx context.Context, db *mongo.Database, topic, key string, payload []byte) error {
	return s.Enqueue(ctx, topic, key, payload)
}

func (s *OutboxService) PublishPending(ctx context.Context, limit int) (int, error) {
	if s.coll == nil || s.publisher == nil {
		return 0, nil
	}

	now := time.Now()
	claimDeadline := now.Add(OutboxClaimLease)

	// Claim pending events using FindOneAndUpdate with a skip pattern
	filter := bson.M{
		"status":          "pending",
		"next_attempt_at": bson.M{"$lte": now},
	}
	update := bson.M{
		"$set": bson.M{
			"next_attempt_at": claimDeadline,
		},
	}
	opts := options.FindOneAndUpdate().
		SetSort(bson.D{{Key: "outbox_id", Value: 1}}).
		SetReturnDocument(options.After)

	var claimed []OutboxEvent
	for i := 0; i < limit; i++ {
		var event OutboxEvent
		err := s.coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&event)
		if err != nil {
			break // No more events to claim
		}
		claimed = append(claimed, event)
	}

	delivered := 0
	for _, event := range claimed {
		if err := s.publisher.SendMessage(ctx, event.Topic, event.EventKey, event.Payload); err != nil {
			s.logger.Error("failed to publish outbox event, scheduling retry",
				zap.Error(err), zap.String("event_key", event.EventKey))

			if int(event.Attempts)+1 >= OutboxMaxAttempts {
				_, _ = s.coll.UpdateOne(ctx, bson.M{"_id": event.ID}, bson.M{
					"$set": bson.M{"status": "dead", "updated_at": time.Now()},
				})
				continue
			}
			nextAttempt := time.Now().Add(OutboxBackoff * time.Duration(event.Attempts+1))
			_, _ = s.coll.UpdateOne(ctx, bson.M{"_id": event.ID}, bson.M{
				"$set": bson.M{"attempts": event.Attempts + 1, "next_attempt_at": nextAttempt, "updated_at": time.Now()},
			})
			continue
		}
		_, _ = s.coll.UpdateOne(ctx, bson.M{"_id": event.ID}, bson.M{
			"$set": bson.M{"status": "delivered", "updated_at": time.Now()},
		})
		delivered++
	}
	return delivered, nil
}

func (s *OutboxService) Start(ctx context.Context, interval time.Duration, limit int) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("outbox relay stopped")
			return
		case <-ticker.C:
			if _, err := s.PublishPending(ctx, limit); err != nil {
				s.logger.Error("outbox relay cycle failed", zap.Error(err))
			}
		}
	}
}

// PurgeRetention removes old delivered/dead events
func (s *OutboxService) PurgeRetention(ctx context.Context) error {
	if s.coll == nil {
		return nil
	}
	cutoff := time.Now().Add(-OutboxRetention)
	result, err := s.coll.DeleteMany(ctx, bson.M{
		"status":     bson.M{"$in": []string{"delivered", "dead"}},
		"updated_at": bson.M{"$lt": cutoff},
	})
	if err != nil {
		return err
	}
	if result.DeletedCount > 0 {
		s.logger.Info("outbox retention cleanup", zap.Int64("removed", result.DeletedCount))
	}
	return nil
}
