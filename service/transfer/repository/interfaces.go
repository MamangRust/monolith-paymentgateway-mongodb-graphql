package repository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/idempotency"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IdempotencyRepository interface {
	idempotency.Store
}

type OutboxRepository interface {
	Insert(ctx context.Context, record OutboxRecord) error
}

type SaldoRepository interface {
	FindByCardNumber(ctx context.Context, card_number string) (*salemongo.Saldo, error)
	UpdateSaldoBalance(ctx context.Context, request *requests.UpdateSaldoBalance) (*salemongo.UpdateSaldoBalanceRow, error)
}

type CardRepository interface {
	FindUserCardByCardNumber(ctx context.Context, card_number string) (*salemongo.CardByEmailRow, error)
	FindCardByCardNumber(ctx context.Context, card_number string) (*salemongo.CardAllFieldsRow, error)
}

type TransferQueryResult struct {
	Transfer   *salemongo.Transfer
	TotalCount int64
}

type TransferQueryRepository interface {
	FindAll(ctx context.Context, req *requests.FindAllTransfers) ([]*TransferQueryResult, error)
	FindByActive(ctx context.Context, req *requests.FindAllTransfers) ([]*TransferQueryResult, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllTransfers) ([]*TransferQueryResult, error)
	FindById(ctx context.Context, id int) (*salemongo.Transfer, error)
	FindTransferByTransferFrom(ctx context.Context, transferFrom string) ([]*salemongo.Transfer, error)
	FindTransferByTransferTo(ctx context.Context, transferTo string) ([]*salemongo.Transfer, error)
}

type TransferCommandRepository interface {
	CreateTransfer(ctx context.Context, request *requests.CreateTransferRequest) (*salemongo.Transfer, error)
	UpdateTransfer(ctx context.Context, request *requests.UpdateTransferRequest) (*salemongo.Transfer, error)
	UpdateTransferAmount(ctx context.Context, request *requests.UpdateTransferAmountRequest) (*salemongo.Transfer, error)
	UpdateTransferStatus(ctx context.Context, request *requests.UpdateTransferStatus) (*salemongo.Transfer, error)
	TransitionStatus(ctx context.Context, id int, fromStatus, toStatus, reason string) (*salemongo.Transfer, error)
	GuardStatus(ctx context.Context, id int, fromStatus, toStatus, reason string) (bool, error)
	ListStuck(ctx context.Context, olderThan time.Duration, maxRows int32) ([]*salemongo.Transfer, error)

	TrashedTransfer(ctx context.Context, transferID int) (*salemongo.Transfer, error)
	RestoreTransfer(ctx context.Context, transferID int) (*salemongo.Transfer, error)
	DeleteTransferPermanent(ctx context.Context, transferID int) (bool, error)

	RestoreAllTransfer(ctx context.Context) (bool, error)
	DeleteAllTransferPermanent(ctx context.Context) (bool, error)

	CreateTransferAtomic(ctx context.Context, request *requests.CreateTransferRequest) (*TransferAtomicResult, error)
}

type transferIdempotencyRecord struct {
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

type TransferIdempotencyRepository struct {
	coll *mongo.Collection
}

func NewTransferIdempotencyRepository(coll *mongo.Collection) *TransferIdempotencyRepository {
	return &TransferIdempotencyRepository{coll: coll}
}

func (r *TransferIdempotencyRepository) Claim(ctx context.Context, scope, key, requestHash string) (*idempotency.Record, error) {
	operationID := uuid.New().String()
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	rec := &transferIdempotencyRecord{
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
		if salemongo.IsDuplicateKeyError(err) {
			return nil, idempotency.ErrKeyInUse
		}
		return nil, err
	}

	return mapTransferIdempotencyRecord(rec), nil
}

func (r *TransferIdempotencyRepository) Get(ctx context.Context, scope, key string) (*idempotency.Record, error) {
	var rec transferIdempotencyRecord
	err := r.coll.FindOne(ctx, bson.M{"scope": scope, "idempotency_key": key}).Decode(&rec)
	if err != nil {
		return nil, err
	}
	return mapTransferIdempotencyRecord(&rec), nil
}

func (r *TransferIdempotencyRepository) Complete(ctx context.Context, scope, key, requestHash string, outcome idempotency.Outcome) error {
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

func (r *TransferIdempotencyRepository) Release(ctx context.Context, scope, key, requestHash string) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"scope": scope, "idempotency_key": key, "request_hash": requestHash})
	return err
}

func mapTransferIdempotencyRecord(rec *transferIdempotencyRecord) *idempotency.Record {
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
