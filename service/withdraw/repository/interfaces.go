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
	UpdateSaldoWithdraw(ctx context.Context, request *requests.UpdateSaldoWithdraw) (*salemongo.UpdateSaldoWithdrawRow, error)
}

type WithdrawQueryResult struct {
	Withdraw   *salemongo.Withdraw
	TotalCount int64
}

type WithdrawQueryRepository interface {
	FindAll(ctx context.Context, req *requests.FindAllWithdraws) ([]*WithdrawQueryResult, error)
	FindByActive(ctx context.Context, req *requests.FindAllWithdraws) ([]*WithdrawQueryResult, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllWithdraws) ([]*WithdrawQueryResult, error)
	FindAllByCardNumber(ctx context.Context, req *requests.FindAllWithdrawCardNumber) ([]*WithdrawQueryResult, error)
	FindById(ctx context.Context, id int) (*salemongo.Withdraw, error)
	GetTodayWithdrawSumByCardNumber(ctx context.Context, cardNumber string) (int64, error)
}

type WithdrawCommandRepository interface {
	CreateWithdraw(ctx context.Context, request *requests.CreateWithdrawRequest) (*salemongo.Withdraw, error)
	UpdateWithdraw(ctx context.Context, request *requests.UpdateWithdrawRequest) (*salemongo.Withdraw, error)
	UpdateWithdrawStatus(ctx context.Context, request *requests.UpdateWithdrawStatus) (*salemongo.Withdraw, error)

	TrashedWithdraw(ctx context.Context, withdrawID int) (*salemongo.Withdraw, error)
	RestoreWithdraw(ctx context.Context, withdrawID int) (*salemongo.Withdraw, error)
	DeleteWithdrawPermanent(ctx context.Context, withdrawID int) (bool, error)

	RestoreAllWithdraw(ctx context.Context) (bool, error)
	DeleteAllWithdrawPermanent(ctx context.Context) (bool, error)
}

type CardRepository interface {
	FindUserCardByCardNumber(ctx context.Context, card_number string) (*salemongo.CardByEmailRow, error)
}

type withdrawIdempotencyRecord struct {
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

type WithdrawIdempotencyRepository struct {
	coll *mongo.Collection
}

func NewWithdrawIdempotencyRepository(coll *mongo.Collection) *WithdrawIdempotencyRepository {
	return &WithdrawIdempotencyRepository{coll: coll}
}

func (r *WithdrawIdempotencyRepository) Claim(ctx context.Context, scope, key, requestHash string) (*idempotency.Record, error) {
	operationID := uuid.New().String()
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	rec := &withdrawIdempotencyRecord{
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

	return mapWithdrawIdempotencyRecord(rec), nil
}

func (r *WithdrawIdempotencyRepository) Get(ctx context.Context, scope, key string) (*idempotency.Record, error) {
	var rec withdrawIdempotencyRecord
	err := r.coll.FindOne(ctx, bson.M{"scope": scope, "idempotency_key": key}).Decode(&rec)
	if err != nil {
		return nil, err
	}
	return mapWithdrawIdempotencyRecord(&rec), nil
}

func (r *WithdrawIdempotencyRepository) Complete(ctx context.Context, scope, key, requestHash string, outcome idempotency.Outcome) error {
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

func (r *WithdrawIdempotencyRepository) Release(ctx context.Context, scope, key, requestHash string) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"scope": scope, "idempotency_key": key, "request_hash": requestHash})
	return err
}

func mapWithdrawIdempotencyRecord(rec *withdrawIdempotencyRecord) *idempotency.Record {
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
