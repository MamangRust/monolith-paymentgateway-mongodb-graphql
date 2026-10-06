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

type topupCommandRepository struct {
	db   *mongo.Database
	coll *mongo.Collection
}

func NewTopupCommandRepository(db *mongo.Database, coll *mongo.Collection) TopupCommandRepository {
	return &topupCommandRepository{db: db, coll: coll}
}

func (r *topupCommandRepository) CreateTopup(ctx context.Context, request *requests.CreateTopupRequest) (*salemongo.Topup, error) {
	topupID, err := salemongo.NextID(ctx, r.db, "topups")
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	now := time.Now()
	topup := &salemongo.Topup{
		TopupID:        topupID,
		TopupNo:        uuid.New(),
		CardNumber:     request.CardNumber,
		TopupAmount:    int64(request.TopupAmount),
		TopupMethod:    request.TopupMethod,
		TopupTime:      now,
		Status:         "pending",
		IdempotencyKey: request.IdempotencyKey,
		OperationID:    uuid.New().String(),
		CreatedAt:      &now,
		UpdatedAt:      &now,
	}

	if _, err := r.coll.InsertOne(ctx, topup); err != nil {
		return nil, sharedErrors.ErrFailed("create topup").WithInternal(err)
	}

	return topup, nil
}

func (r *topupCommandRepository) UpdateTopup(ctx context.Context, request *requests.UpdateTopupRequest) (*salemongo.Topup, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"topup_id": int32(*request.TopupID), "deleted_at": nil},
		bson.M{"$set": bson.M{
			"card_number":  request.CardNumber,
			"topup_amount": int64(request.TopupAmount),
			"topup_method": request.TopupMethod,
			"updated_at":   now,
		}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var topup salemongo.Topup
	if err := result.Decode(&topup); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("topup")
		}
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "topup", "update topup")
	}
	return &topup, nil
}

func (r *topupCommandRepository) UpdateTopupAmount(ctx context.Context, request *requests.UpdateTopupAmount) (*salemongo.Topup, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"topup_id": int32(request.TopupID), "deleted_at": nil},
		bson.M{"$set": bson.M{
			"topup_amount": int64(request.TopupAmount),
			"updated_at":   now,
		}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var topup salemongo.Topup
	if err := result.Decode(&topup); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("topup")
		}
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "topup", "update topup amount")
	}
	return &topup, nil
}

// CompareAndSwapTopupAmount atomically moves a topup's amount from expected to
// newAmount. It reports false when another writer already changed the amount, so
// callers can re-read and retry against the fresh value.
func (r *topupCommandRepository) CompareAndSwapTopupAmount(ctx context.Context, topupID, expectedAmount, newAmount int) (bool, error) {
	result, err := r.coll.UpdateOne(
		ctx,
		bson.M{
			"topup_id":     int32(topupID),
			"topup_amount": int64(expectedAmount),
			"deleted_at":   nil,
		},
		bson.M{"$set": bson.M{
			"topup_amount": int64(newAmount),
			"updated_at":   time.Now(),
		}},
	)
	if err != nil {
		return false, sharedErrors.ErrFailed("update topup amount").WithInternal(err)
	}

	return result.MatchedCount == 1, nil
}

func (r *topupCommandRepository) UpdateTopupStatus(ctx context.Context, request *requests.UpdateTopupStatus) (*salemongo.Topup, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"topup_id": int32(request.TopupID), "deleted_at": nil},
		bson.M{"$set": bson.M{
			"status":     request.Status,
			"updated_at": now,
		}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var topup salemongo.Topup
	if err := result.Decode(&topup); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("topup")
		}
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "topup", "update topup status")
	}
	return &topup, nil
}

func (r *topupCommandRepository) TrashedTopup(ctx context.Context, topup_id int) (*salemongo.Topup, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"topup_id": int32(topup_id), "deleted_at": nil},
		bson.M{"$set": bson.M{"deleted_at": now}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var topup salemongo.Topup
	if err := result.Decode(&topup); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("topup")
		}
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "topup", "trash topup")
	}
	return &topup, nil
}

func (r *topupCommandRepository) RestoreTopup(ctx context.Context, topup_id int) (*salemongo.Topup, error) {
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"topup_id": int32(topup_id), "deleted_at": bson.M{"$ne": nil}},
		bson.M{"$unset": bson.M{"deleted_at": ""}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var topup salemongo.Topup
	if err := result.Decode(&topup); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("topup")
		}
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "topup", "restore topup")
	}
	return &topup, nil
}

func (r *topupCommandRepository) DeleteTopupPermanent(ctx context.Context, topup_id int) (bool, error) {
	result, err := r.coll.DeleteOne(ctx, bson.M{"topup_id": int32(topup_id)})
	if err != nil {
		return false, sharedErrors.ErrNoRowsOrFailed(err, "topup", "delete topup permanently")
	}
	if result.DeletedCount == 0 {
		return false, sharedErrors.ErrNotFoundResponse("topup")
	}
	return true, nil
}

func (r *topupCommandRepository) RestoreAllTopup(ctx context.Context) (bool, error) {
	_, err := r.coll.UpdateMany(ctx, bson.M{"deleted_at": bson.M{"$ne": nil}}, bson.M{"$unset": bson.M{"deleted_at": ""}})
	if err != nil {
		return false, sharedErrors.ErrFailed("restore all topups").WithInternal(err)
	}
	return true, nil
}

func (r *topupCommandRepository) DeleteAllTopupPermanent(ctx context.Context) (bool, error) {
	_, err := r.coll.DeleteMany(ctx, bson.M{"deleted_at": bson.M{"$ne": nil}})
	if err != nil {
		return false, sharedErrors.ErrFailed("delete all topups permanently").WithInternal(err)
	}
	return true, nil
}
