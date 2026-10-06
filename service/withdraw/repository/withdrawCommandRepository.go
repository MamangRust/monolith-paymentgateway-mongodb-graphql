package repository

import (
	"context"
	"errors"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type withdrawCommandRepository struct {
	db   *mongo.Database
	coll *mongo.Collection
}

func NewWithdrawCommandRepository(db *mongo.Database, coll *mongo.Collection) WithdrawCommandRepository {
	return &withdrawCommandRepository{db: db, coll: coll}
}

func (r *withdrawCommandRepository) CreateWithdraw(ctx context.Context, request *requests.CreateWithdrawRequest) (*salemongo.Withdraw, error) {
	withdrawID, err := salemongo.NextID(ctx, r.db, "withdraws")
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	now := time.Now()
	withdraw := &salemongo.Withdraw{
		WithdrawID:     withdrawID,
		WithdrawNo:     uuid.New(),
		CardNumber:     request.CardNumber,
		WithdrawAmount: int64(request.WithdrawAmount),
		WithdrawTime:   request.WithdrawTime,
		Status:         "pending",
		OperationID:    uuid.New().String(),
		CreatedAt:      &now,
		UpdatedAt:      &now,
	}

	if _, err := r.coll.InsertOne(ctx, withdraw); err != nil {
		return nil, sharedErrors.ErrFailed("create withdraw").WithInternal(err)
	}

	return withdraw, nil
}

func (r *withdrawCommandRepository) UpdateWithdraw(ctx context.Context, request *requests.UpdateWithdrawRequest) (*salemongo.Withdraw, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"withdraw_id": int32(*request.WithdrawID), "deleted_at": nil},
		bson.M{"$set": bson.M{
			"card_number":     request.CardNumber,
			"withdraw_amount": int64(request.WithdrawAmount),
			"withdraw_time":   request.WithdrawTime,
			"updated_at":      now,
		}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var withdraw salemongo.Withdraw
	if err := result.Decode(&withdraw); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("withdraw")
		}
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "withdraw", "update withdraw")
	}
	return &withdraw, nil
}

func (r *withdrawCommandRepository) UpdateWithdrawStatus(ctx context.Context, request *requests.UpdateWithdrawStatus) (*salemongo.Withdraw, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"withdraw_id": int32(request.WithdrawID), "deleted_at": nil},
		bson.M{"$set": bson.M{
			"status":     request.Status,
			"updated_at": now,
		}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var withdraw salemongo.Withdraw
	if err := result.Decode(&withdraw); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("withdraw")
		}
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "withdraw", "update withdraw status")
	}
	return &withdraw, nil
}

func (r *withdrawCommandRepository) TrashedWithdraw(ctx context.Context, withdraw_id int) (*salemongo.Withdraw, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"withdraw_id": int32(withdraw_id), "deleted_at": nil},
		bson.M{"$set": bson.M{"deleted_at": now}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var withdraw salemongo.Withdraw
	if err := result.Decode(&withdraw); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("withdraw")
		}
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "withdraw", "trash withdraw")
	}
	return &withdraw, nil
}

func (r *withdrawCommandRepository) RestoreWithdraw(ctx context.Context, withdraw_id int) (*salemongo.Withdraw, error) {
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"withdraw_id": int32(withdraw_id), "deleted_at": bson.M{"$ne": nil}},
		bson.M{"$unset": bson.M{"deleted_at": ""}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var withdraw salemongo.Withdraw
	if err := result.Decode(&withdraw); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("withdraw")
		}
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "withdraw", "restore withdraw")
	}
	return &withdraw, nil
}

func (r *withdrawCommandRepository) DeleteWithdrawPermanent(ctx context.Context, withdraw_id int) (bool, error) {
	result, err := r.coll.DeleteOne(ctx, bson.M{"withdraw_id": int32(withdraw_id)})
	if err != nil {
		return false, sharedErrors.ErrNoRowsOrFailed(err, "withdraw", "delete withdraw permanently")
	}
	if result.DeletedCount == 0 {
		return false, sharedErrors.ErrNotFoundResponse("withdraw")
	}
	return true, nil
}

func (r *withdrawCommandRepository) RestoreAllWithdraw(ctx context.Context) (bool, error) {
	_, err := r.coll.UpdateMany(ctx, bson.M{"deleted_at": bson.M{"$ne": nil}}, bson.M{"$unset": bson.M{"deleted_at": ""}})
	if err != nil {
		return false, sharedErrors.ErrFailed("restore all withdraws").WithInternal(err)
	}
	return true, nil
}

func (r *withdrawCommandRepository) DeleteAllWithdrawPermanent(ctx context.Context) (bool, error) {
	_, err := r.coll.DeleteMany(ctx, bson.M{"deleted_at": bson.M{"$ne": nil}})
	if err != nil {
		return false, sharedErrors.ErrFailed("delete all withdraws permanently").WithInternal(err)
	}
	return true, nil
}
