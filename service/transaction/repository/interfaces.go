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

// TransactionQueryResult is the paginated transaction list response.
type TransactionQueryResult struct {
	Transaction *salemongo.Transaction
	TotalCount  int64
}

// TransactionByCardResult is the paginated transaction list by card number.
type TransactionByCardResult struct {
	Transaction *salemongo.Transaction
	TotalCount  int64
}

type IdempotencyRepository interface {
	idempotency.Store
}

type OutboxRepository interface {
	Insert(ctx context.Context, record OutboxRecord) error
}

type MerchantRepository interface {
	FindByApiKey(ctx context.Context, api_key string) (*salemongo.MerchantAllFieldsRow, error)
}

type SaldoRepository interface {
	FindByCardNumber(ctx context.Context, card_number string) (*salemongo.Saldo, error)
	UpdateSaldoBalance(ctx context.Context, request *requests.UpdateSaldoBalance) (*salemongo.UpdateSaldoBalanceRow, error)
}

type CardRepository interface {
	FindCardByUserId(ctx context.Context, user_id int) (*salemongo.CardAllFieldsRow, error)
	FindUserCardByCardNumber(ctx context.Context, card_number string) (*salemongo.CardByEmailRow, error)
	FindCardByCardNumber(ctx context.Context, card_number string) (*salemongo.CardAllFieldsRow, error)
	UpdateCard(ctx context.Context, request *requests.UpdateCardRequest) (*salemongo.CardUpdateRow, error)
}

type TransactionQueryRepository interface {
	FindAllTransactions(ctx context.Context, req *requests.FindAllTransactions) ([]*TransactionQueryResult, error)
	FindByActive(ctx context.Context, req *requests.FindAllTransactions) ([]*TransactionQueryResult, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllTransactions) ([]*TransactionQueryResult, error)
	FindAllTransactionByCardNumber(ctx context.Context, req *requests.FindAllTransactionCardNumber) ([]*TransactionByCardResult, error)
	FindById(ctx context.Context, transaction_id int) (*salemongo.Transaction, error)
	FindTransactionByMerchantId(ctx context.Context, merchant_id int) ([]*salemongo.Transaction, error)
}

type TransactionCommandRepository interface {
	CreateTransaction(ctx context.Context, request *requests.CreateTransactionRequest) (*salemongo.Transaction, error)
	UpdateTransaction(ctx context.Context, request *requests.UpdateTransactionRequest) (*salemongo.Transaction, error)
	UpdateTransactionStatus(ctx context.Context, request *requests.UpdateTransactionStatus) (*salemongo.Transaction, error)
	TransitionStatus(ctx context.Context, id int, fromStatus, toStatus, reason string) (*salemongo.Transaction, error)
	GuardStatus(ctx context.Context, id int, fromStatus, toStatus, reason string) (bool, error)
	ListStuck(ctx context.Context, olderThan time.Duration, maxRows int32) ([]*salemongo.Transaction, error)
	TrashedTransaction(ctx context.Context, transaction_id int) (*salemongo.Transaction, error)
	RestoreTransaction(ctx context.Context, topup_id int) (*salemongo.Transaction, error)
	DeleteTransactionPermanent(ctx context.Context, topup_id int) (bool, error)

	RestoreAllTransaction(ctx context.Context) (bool, error)
	DeleteAllTransactionPermanent(ctx context.Context) (bool, error)
}

type transactionIdempotencyRecord struct {
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

type TransactionIdempotencyRepository struct {
	coll *mongo.Collection
}

func NewTransactionIdempotencyRepository(coll *mongo.Collection) *TransactionIdempotencyRepository {
	return &TransactionIdempotencyRepository{coll: coll}
}

func (r *TransactionIdempotencyRepository) Claim(ctx context.Context, scope, key, requestHash string) (*idempotency.Record, error) {
	operationID := uuid.New().String()
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	rec := &transactionIdempotencyRecord{
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

	return mapTransactionIdempotencyRecord(rec), nil
}

func (r *TransactionIdempotencyRepository) Get(ctx context.Context, scope, key string) (*idempotency.Record, error) {
	var rec transactionIdempotencyRecord
	err := r.coll.FindOne(ctx, bson.M{"scope": scope, "idempotency_key": key}).Decode(&rec)
	if err != nil {
		return nil, err
	}
	return mapTransactionIdempotencyRecord(&rec), nil
}

func (r *TransactionIdempotencyRepository) Complete(ctx context.Context, scope, key, requestHash string, outcome idempotency.Outcome) error {
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

func (r *TransactionIdempotencyRepository) Release(ctx context.Context, scope, key, requestHash string) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"scope": scope, "idempotency_key": key, "request_hash": requestHash})
	return err
}

func mapTransactionIdempotencyRecord(rec *transactionIdempotencyRecord) *idempotency.Record {
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
