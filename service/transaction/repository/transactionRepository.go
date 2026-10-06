package repository

import (
	"context"
	"errors"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"github.com/MamangRust/monolith-payment-gateway-shared/state"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type transactionCommandRepository struct {
	db   *mongo.Database
	coll *mongo.Collection
}

func NewTransactionCommandRepository(db *mongo.Database, coll *mongo.Collection) TransactionCommandRepository {
	return &transactionCommandRepository{db: db, coll: coll}
}

func (r *transactionCommandRepository) CreateTransaction(ctx context.Context, request *requests.CreateTransactionRequest) (*salemongo.Transaction, error) {
	transactionID, err := salemongo.NextID(ctx, r.db, "transactions")
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	now := time.Now()
	txn := &salemongo.Transaction{
		TransactionID:   transactionID,
		TransactionNo:   uuid.New(),
		CardNumber:      request.CardNumber,
		Amount:          int64(request.Amount),
		PaymentMethod:   request.PaymentMethod,
		MerchantID:      int32(*request.MerchantID),
		TransactionTime: request.TransactionTime,
		Status:          "pending",
		OperationID:     uuid.New().String(),
		CreatedAt:       &now,
		UpdatedAt:       &now,
	}

	if _, err := r.coll.InsertOne(ctx, txn); err != nil {
		return nil, sharedErrors.ErrFailed("create transaction").WithInternal(err)
	}

	return txn, nil
}

func (r *transactionCommandRepository) UpdateTransaction(ctx context.Context, request *requests.UpdateTransactionRequest) (*salemongo.Transaction, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"transaction_id": int32(*request.TransactionID), "deleted_at": nil},
		bson.M{"$set": bson.M{
			"card_number":      request.CardNumber,
			"amount":           int64(request.Amount),
			"payment_method":   request.PaymentMethod,
			"merchant_id":      int32(*request.MerchantID),
			"transaction_time": request.TransactionTime,
			"updated_at":       now,
		}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var txn salemongo.Transaction
	if err := result.Decode(&txn); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("transaction")
		}
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transaction", "update transaction")
	}
	return &txn, nil
}

func (r *transactionCommandRepository) UpdateTransactionStatus(ctx context.Context, request *requests.UpdateTransactionStatus) (*salemongo.Transaction, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"transaction_id": int32(request.TransactionID), "deleted_at": nil},
		bson.M{"$set": bson.M{
			"status":     request.Status,
			"updated_at": now,
		}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var txn salemongo.Transaction
	if err := result.Decode(&txn); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("transaction")
		}
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transaction", "update transaction status")
	}
	return &txn, nil
}

func (r *transactionCommandRepository) GuardStatus(ctx context.Context, id int, fromStatus, toStatus, reason string) (bool, error) {
	if err := state.CheckTransition(fromStatus, toStatus); err != nil {
		return false, err
	}

	result, err := r.coll.UpdateOne(
		ctx,
		bson.M{"transaction_id": int32(id), "status": fromStatus},
		bson.M{"$set": bson.M{"status": toStatus, "updated_at": time.Now()}},
	)
	if err != nil {
		return false, err
	}
	if result.ModifiedCount == 0 {
		return false, nil
	}
	return true, nil
}

func (r *transactionCommandRepository) TransitionStatus(ctx context.Context, id int, fromStatus, toStatus, reason string) (*salemongo.Transaction, error) {
	if err := state.CheckTransition(fromStatus, toStatus); err != nil {
		return nil, err
	}

	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"transaction_id": int32(id), "status": fromStatus},
		bson.M{"$set": bson.M{"status": toStatus, "updated_at": time.Now()}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var txn salemongo.Transaction
	if err := result.Decode(&txn); err != nil {
		return nil, err
	}
	return &txn, nil
}

func (r *transactionCommandRepository) ListStuck(ctx context.Context, olderThan time.Duration, maxRows int32) ([]*salemongo.Transaction, error) {
	threshold := time.Now().Add(-olderThan)
	cursor, err := r.coll.Find(
		ctx,
		bson.M{
			"status":     bson.M{"$in": bson.A{"processing", "compensating"}},
			"updated_at": bson.M{"$lt": threshold},
		},
		options.Find().SetLimit(int64(maxRows)),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var txns []*salemongo.Transaction
	if err := cursor.All(ctx, &txns); err != nil {
		return nil, err
	}
	return txns, nil
}

func (r *transactionCommandRepository) TrashedTransaction(ctx context.Context, transaction_id int) (*salemongo.Transaction, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"transaction_id": int32(transaction_id), "deleted_at": nil},
		bson.M{"$set": bson.M{"deleted_at": now}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var txn salemongo.Transaction
	if err := result.Decode(&txn); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("transaction")
		}
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transaction", "trash transaction")
	}
	return &txn, nil
}

func (r *transactionCommandRepository) RestoreTransaction(ctx context.Context, transaction_id int) (*salemongo.Transaction, error) {
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"transaction_id": int32(transaction_id), "deleted_at": bson.M{"$ne": nil}},
		bson.M{"$unset": bson.M{"deleted_at": ""}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var txn salemongo.Transaction
	if err := result.Decode(&txn); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("transaction")
		}
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transaction", "restore transaction")
	}
	return &txn, nil
}

func (r *transactionCommandRepository) DeleteTransactionPermanent(ctx context.Context, transaction_id int) (bool, error) {
	result, err := r.coll.DeleteOne(ctx, bson.M{"transaction_id": int32(transaction_id)})
	if err != nil {
		return false, sharedErrors.ErrNoRowsOrFailed(err, "transaction", "delete transaction permanently")
	}
	if result.DeletedCount == 0 {
		return false, sharedErrors.ErrNotFoundResponse("transaction")
	}
	return true, nil
}

func (r *transactionCommandRepository) RestoreAllTransaction(ctx context.Context) (bool, error) {
	_, err := r.coll.UpdateMany(ctx, bson.M{"deleted_at": bson.M{"$ne": nil}}, bson.M{"$unset": bson.M{"deleted_at": ""}})
	if err != nil {
		return false, sharedErrors.ErrFailed("restore all transactions").WithInternal(err)
	}
	return true, nil
}

func (r *transactionCommandRepository) DeleteAllTransactionPermanent(ctx context.Context) (bool, error) {
	_, err := r.coll.DeleteMany(ctx, bson.M{"deleted_at": bson.M{"$ne": nil}})
	if err != nil {
		return false, sharedErrors.ErrFailed("delete all transactions permanently").WithInternal(err)
	}
	return true, nil
}
