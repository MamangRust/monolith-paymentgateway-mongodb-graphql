package repository

import (
	"context"
	"time"

	apikey "github.com/MamangRust/monolith-payment-gateway-pkg/api-key"
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type merchantCommandRepository struct {
	db   *mongo.Database
	coll *mongo.Collection
}

func NewMerchantCommandRepository(db *mongo.Database, coll *mongo.Collection) MerchantCommandRepository {
	return &merchantCommandRepository{db: db, coll: coll}
}

func (r *merchantCommandRepository) CreateMerchant(ctx context.Context, request *requests.CreateMerchantRequest) (*salemongo.Merchant, error) {
	apiKey, err := apikey.GenerateApiKey()
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	now := time.Now()
	merchantID, err := salemongo.NextID(ctx, r.db, "merchants")
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	merchant := &salemongo.Merchant{
		MerchantID: merchantID,
		MerchantNo: uuid.New(),
		Name:       request.Name,
		ApiKey:     apiKey,
		UserID:     int32(request.UserID),
		Status:     "inactive",
		CreatedAt:  &now,
		UpdatedAt:  &now,
	}

	result, err := r.coll.InsertOne(ctx, merchant)
	if err != nil {
		if salemongo.IsDuplicateKeyError(err) {
			return nil, sharedErrors.NewConflictError("merchant api key already exists").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("create merchant").WithInternal(err)
	}
	merchant.ID = result.InsertedID.(primitive.ObjectID)
	return merchant, nil
}

func (r *merchantCommandRepository) UpdateMerchant(ctx context.Context, request *requests.UpdateMerchantRequest) (*salemongo.Merchant, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"merchant_id": *request.MerchantID, "deleted_at": nil},
		bson.M{"$set": bson.M{
			"name":       request.Name,
			"user_id":    int32(request.UserID),
			"status":     request.Status,
			"updated_at": now,
		}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var merchant salemongo.Merchant
	if err := result.Decode(&merchant); err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "merchant", "update merchant")
	}
	return &merchant, nil
}

func (r *merchantCommandRepository) UpdateMerchantStatus(ctx context.Context, request *requests.UpdateMerchantStatusRequest) (*salemongo.Merchant, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"merchant_id": *request.MerchantID, "deleted_at": nil},
		bson.M{"$set": bson.M{"status": request.Status, "updated_at": now}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var merchant salemongo.Merchant
	if err := result.Decode(&merchant); err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "merchant", "update merchant status")
	}
	return &merchant, nil
}

func (r *merchantCommandRepository) TrashedMerchant(ctx context.Context, merchantID int) (*salemongo.Merchant, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"merchant_id": merchantID, "deleted_at": nil},
		bson.M{"$set": bson.M{"deleted_at": now}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var merchant salemongo.Merchant
	if err := result.Decode(&merchant); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, sharedErrors.ErrNotFoundResponse("merchant")
		}
		return nil, sharedErrors.ErrFailed("trash merchant").WithInternal(err)
	}
	return &merchant, nil
}

func (r *merchantCommandRepository) RestoreMerchant(ctx context.Context, merchantID int) (*salemongo.Merchant, error) {
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"merchant_id": merchantID, "deleted_at": bson.M{"$ne": nil}},
		bson.M{"$unset": bson.M{"deleted_at": ""}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var merchant salemongo.Merchant
	if err := result.Decode(&merchant); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, sharedErrors.ErrNotFoundResponse("merchant")
		}
		return nil, sharedErrors.ErrFailed("restore merchant").WithInternal(err)
	}
	return &merchant, nil
}

func (r *merchantCommandRepository) DeleteMerchantPermanent(ctx context.Context, merchantID int) (bool, error) {
	result, err := r.coll.DeleteMany(ctx, bson.M{"merchant_id": merchantID})
	if err != nil {
		return false, sharedErrors.ErrNoRowsOrFailed(err, "merchant", "delete merchant permanently")
	}
	if result.DeletedCount == 0 {
		return false, sharedErrors.ErrNotFoundResponse("merchant")
	}
	return true, nil
}

func (r *merchantCommandRepository) RestoreAllMerchant(ctx context.Context) (bool, error) {
	_, err := r.coll.UpdateMany(ctx, bson.M{"deleted_at": bson.M{"$ne": nil}}, bson.M{"$unset": bson.M{"deleted_at": ""}})
	if err != nil {
		return false, sharedErrors.ErrFailed("restore all merchants").WithInternal(err)
	}
	return true, nil
}

func (r *merchantCommandRepository) DeleteAllMerchantPermanent(ctx context.Context) (bool, error) {
	_, err := r.coll.DeleteMany(ctx, bson.M{"deleted_at": bson.M{"$ne": nil}})
	if err != nil {
		return false, sharedErrors.ErrFailed("delete all merchants permanently").WithInternal(err)
	}
	return true, nil
}
