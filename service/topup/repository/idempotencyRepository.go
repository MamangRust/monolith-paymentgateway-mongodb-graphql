package repository

import (
	"context"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-shared/idempotency"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type topupIdempotencyRecord struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	Scope           string             `bson:"scope"`
	IdempotencyKey  string             `bson:"idempotency_key"`
	RequestHash     string             `bson:"request_hash"`
	OperationID     *string            `bson:"operation_id"`
	Status          string             `bson:"status"`
	ResponsePayload []byte             `bson:"response_payload,omitempty"`
	ResourceID      *int32             `bson:"resource_id,omitempty"`
	CreatedAt       time.Time          `bson:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at"`
	ExpiresAt       time.Time          `bson:"expires_at"`
}

type TopupIdempotencyRepository struct {
	coll *mongo.Collection
}

func NewTopupIdempotencyRepository(coll *mongo.Collection) *TopupIdempotencyRepository {
	return &TopupIdempotencyRepository{coll: coll}
}

func (r *TopupIdempotencyRepository) Claim(ctx context.Context, scope, key, requestHash string) (*idempotency.Record, error) {
	operationID := uuid.New().String()
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	rec := &topupIdempotencyRecord{
		Scope:          scope,
		IdempotencyKey: key,
		RequestHash:    requestHash,
		OperationID:    &operationID,
		Status:         "processing",
		CreatedAt:      now,
		UpdatedAt:      now,
		ExpiresAt:      expiresAt,
	}

	_, err := r.coll.InsertOne(ctx, rec)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			// Key already in use
			return nil, idempotency.ErrKeyInUse
		}
		return nil, err
	}

	return mapTopupIdempotencyRecord(rec), nil
}

func (r *TopupIdempotencyRepository) Get(ctx context.Context, scope, key string) (*idempotency.Record, error) {
	var rec topupIdempotencyRecord
	err := r.coll.FindOne(ctx, bson.M{"scope": scope, "idempotency_key": key}).Decode(&rec)
	if err != nil {
		return nil, err
	}
	return mapTopupIdempotencyRecord(&rec), nil
}

func (r *TopupIdempotencyRepository) Complete(ctx context.Context, scope, key, requestHash string, outcome idempotency.Outcome) error {
	_, err := r.coll.UpdateOne(
		ctx,
		bson.M{"scope": scope, "idempotency_key": key, "request_hash": requestHash},
		bson.M{"$set": bson.M{
			"status":           outcome.Status,
			"response_payload": outcome.ResponseJSON,
			"resource_id":      outcome.ResourceID,
			"updated_at":       time.Now(),
		}},
	)
	return err
}

func (r *TopupIdempotencyRepository) Release(ctx context.Context, scope, key, requestHash string) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"scope": scope, "idempotency_key": key, "request_hash": requestHash})
	return err
}

func mapTopupIdempotencyRecord(rec *topupIdempotencyRecord) *idempotency.Record {
	if rec == nil {
		return nil
	}
	var opID string
	if rec.OperationID != nil {
		opID = *rec.OperationID
	}
	return &idempotency.Record{
		Key:          rec.IdempotencyKey,
		RequestHash:  rec.RequestHash,
		OperationID:  opID,
		Status:       rec.Status,
		ResponseJSON: rec.ResponsePayload,
		ResourceID:   rec.ResourceID,
		CreatedAt:    rec.CreatedAt,
		UpdatedAt:    rec.UpdatedAt,
		ExpiresAt:    rec.ExpiresAt,
	}
}
