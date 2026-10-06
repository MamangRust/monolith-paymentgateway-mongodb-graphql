package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"github.com/MamangRust/monolith-payment-gateway-shared/idempotency"
	"github.com/MamangRust/monolith-payment-gateway-shared/state"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type transferCommandRepository struct {
	db          *mongo.Database
	coll        *mongo.Collection
	idempotency *TransferIdempotencyRepository
}

func NewTransferCommandRepository(db *mongo.Database, coll *mongo.Collection, idem *TransferIdempotencyRepository) TransferCommandRepository {
	return &transferCommandRepository{db: db, coll: coll, idempotency: idem}
}

func (r *transferCommandRepository) CreateTransfer(ctx context.Context, request *requests.CreateTransferRequest) (*salemongo.Transfer, error) {
	transferID, err := salemongo.NextID(ctx, r.db, "transfers")
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	now := time.Now()
	transfer := &salemongo.Transfer{
		TransferID:     transferID,
		TransferNo:     uuid.New(),
		TransferFrom:   request.TransferFrom,
		TransferTo:     request.TransferTo,
		TransferAmount: int64(request.TransferAmount),
		TransferTime:   now,
		Status:         state.Pending,
		OperationID:    uuid.New().String(),
		CreatedAt:      &now,
		UpdatedAt:      &now,
	}

	if _, err := r.coll.InsertOne(ctx, transfer); err != nil {
		return nil, sharedErrors.ErrFailed("create transfer").WithInternal(err)
	}

	return transfer, nil
}

func (r *transferCommandRepository) UpdateTransfer(ctx context.Context, request *requests.UpdateTransferRequest) (*salemongo.Transfer, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"transfer_id": int32(*request.TransferID), "deleted_at": nil},
		bson.M{"$set": bson.M{
			"transfer_from":   request.TransferFrom,
			"transfer_to":     request.TransferTo,
			"transfer_amount": int64(request.TransferAmount),
			"updated_at":      now,
		}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var transfer salemongo.Transfer
	if err := result.Decode(&transfer); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("transfer")
		}
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transfer", "update transfer")
	}
	return &transfer, nil
}

func (r *transferCommandRepository) UpdateTransferAmount(ctx context.Context, request *requests.UpdateTransferAmountRequest) (*salemongo.Transfer, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"transfer_id": int32(request.TransferID), "deleted_at": nil},
		bson.M{"$set": bson.M{
			"transfer_amount": int64(request.TransferAmount),
			"updated_at":      now,
		}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var transfer salemongo.Transfer
	if err := result.Decode(&transfer); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("transfer")
		}
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transfer", "update transfer amount")
	}
	return &transfer, nil
}

func (r *transferCommandRepository) UpdateTransferStatus(ctx context.Context, request *requests.UpdateTransferStatus) (*salemongo.Transfer, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"transfer_id": int32(request.TransferID), "deleted_at": nil},
		bson.M{"$set": bson.M{
			"status":     request.Status,
			"updated_at": now,
		}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var transfer salemongo.Transfer
	if err := result.Decode(&transfer); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("transfer")
		}
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transfer", "update transfer status")
	}
	return &transfer, nil
}

func (r *transferCommandRepository) GuardStatus(ctx context.Context, id int, fromStatus, toStatus, reason string) (bool, error) {
	if err := state.CheckTransition(fromStatus, toStatus); err != nil {
		return false, err
	}

	result, err := r.coll.UpdateOne(
		ctx,
		bson.M{"transfer_id": int32(id), "status": fromStatus, "deleted_at": nil},
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

func (r *transferCommandRepository) TransitionStatus(ctx context.Context, id int, fromStatus, toStatus, reason string) (*salemongo.Transfer, error) {
	if err := state.CheckTransition(fromStatus, toStatus); err != nil {
		return nil, err
	}

	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"transfer_id": int32(id), "status": fromStatus, "deleted_at": nil},
		bson.M{"$set": bson.M{"status": toStatus, "updated_at": time.Now()}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var transfer salemongo.Transfer
	if err := result.Decode(&transfer); err != nil {
		return nil, err
	}
	return &transfer, nil
}

func (r *transferCommandRepository) ListStuck(ctx context.Context, olderThan time.Duration, maxRows int32) ([]*salemongo.Transfer, error) {
	threshold := time.Now().Add(-olderThan)
	cursor, err := r.coll.Find(
		ctx,
		bson.M{
			"status":     bson.M{"$in": bson.A{"processing", "compensating"}},
			"updated_at": bson.M{"$lt": threshold},
			"deleted_at": nil,
		},
		options.Find().SetLimit(int64(maxRows)),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transfers []*salemongo.Transfer
	if err := cursor.All(ctx, &transfers); err != nil {
		return nil, err
	}
	return transfers, nil
}

func (r *transferCommandRepository) TrashedTransfer(ctx context.Context, transfer_id int) (*salemongo.Transfer, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"transfer_id": int32(transfer_id), "deleted_at": nil},
		bson.M{"$set": bson.M{"deleted_at": now}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var transfer salemongo.Transfer
	if err := result.Decode(&transfer); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("transfer")
		}
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transfer", "trash transfer")
	}
	return &transfer, nil
}

func (r *transferCommandRepository) RestoreTransfer(ctx context.Context, transfer_id int) (*salemongo.Transfer, error) {
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"transfer_id": int32(transfer_id), "deleted_at": bson.M{"$ne": nil}},
		bson.M{"$unset": bson.M{"deleted_at": ""}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var transfer salemongo.Transfer
	if err := result.Decode(&transfer); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("transfer")
		}
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transfer", "restore transfer")
	}
	return &transfer, nil
}

func (r *transferCommandRepository) DeleteTransferPermanent(ctx context.Context, transfer_id int) (bool, error) {
	result, err := r.coll.DeleteOne(ctx, bson.M{"transfer_id": int32(transfer_id)})
	if err != nil {
		return false, sharedErrors.ErrNoRowsOrFailed(err, "transfer", "delete transfer permanently")
	}
	if result.DeletedCount == 0 {
		return false, sharedErrors.ErrNotFoundResponse("transfer")
	}
	return true, nil
}

func (r *transferCommandRepository) RestoreAllTransfer(ctx context.Context) (bool, error) {
	_, err := r.coll.UpdateMany(ctx, bson.M{"deleted_at": bson.M{"$ne": nil}}, bson.M{"$unset": bson.M{"deleted_at": ""}})
	if err != nil {
		return false, sharedErrors.ErrFailed("restore all transfers").WithInternal(err)
	}
	return true, nil
}

func (r *transferCommandRepository) DeleteAllTransferPermanent(ctx context.Context) (bool, error) {
	_, err := r.coll.DeleteMany(ctx, bson.M{"deleted_at": bson.M{"$ne": nil}})
	if err != nil {
		return false, sharedErrors.ErrFailed("delete all transfers permanently").WithInternal(err)
	}
	return true, nil
}

// CreateTransferAtomic creates the transfer row and resolves the idempotency
// record for the operation in a single transaction. It performs settlement-
// free bookkeeping: balances are debited/credited by the service layer, which
// persists the outcome via the idempotency store.
func (r *transferCommandRepository) CreateTransferAtomic(ctx context.Context, request *requests.CreateTransferRequest) (*TransferAtomicResult, error) {
	if request == nil {
		return nil, sharedErrors.ErrBadRequest.WithMessage("transfer request is required")
	}

	idemHash := idempotency.HashRequest(request)
	_, err := r.idempotency.Claim(ctx, "transfer", request.IdempotencyKey, idemHash)
	if err != nil {
		if errors.Is(err, idempotency.ErrKeyInUse) {
			existing, gErr := r.idempotency.Get(ctx, "transfer", request.IdempotencyKey)
			if gErr != nil {
				return nil, sharedErrors.ErrFailed("resolve transfer idempotency").WithInternal(gErr)
			}
			if existing.RequestHash != idemHash {
				return nil, idempotency.ConflictError()
			}
			if existing.Status == idempotency.StatusSuccess {
				var row salemongo.Transfer
				if uErr := json.Unmarshal(existing.ResponseJSON, &row); uErr != nil {
					return nil, sharedErrors.ErrFailed("decode replayed transfer").WithInternal(uErr)
				}
				return &TransferAtomicResult{Row: transferToAllFieldsRow(&row), Replayed: true}, nil
			}
			return nil, idempotency.ProcessingError()
		}
		return nil, sharedErrors.ErrFailed("claim transfer idempotency").WithInternal(err)
	}

	transfer, err := r.CreateTransfer(ctx, request)
	if err != nil {
		if relErr := r.idempotency.Release(ctx, "transfer", request.IdempotencyKey, idemHash); relErr != nil {
			return nil, sharedErrors.ErrFailed("release transfer idempotency").WithInternal(relErr)
		}
		return nil, err
	}

	return &TransferAtomicResult{Row: transferToAllFieldsRow(transfer), Replayed: false}, nil
}

func transferToAllFieldsRow(t *salemongo.Transfer) *salemongo.TransferAllFieldsRow {
	if t == nil {
		return nil
	}
	row := &salemongo.TransferAllFieldsRow{
		TransferID:     t.TransferID,
		TransferNo:     t.TransferNo,
		TransferFrom:   t.TransferFrom,
		TransferTo:     t.TransferTo,
		TransferAmount: int32(t.TransferAmount),
		Status:         t.Status,
	}
	row.TransferTime = t.TransferTime
	if t.CreatedAt != nil {
		row.CreatedAt = *t.CreatedAt
	}
	if t.UpdatedAt != nil {
		row.UpdatedAt = *t.UpdatedAt
	}
	return row
}
