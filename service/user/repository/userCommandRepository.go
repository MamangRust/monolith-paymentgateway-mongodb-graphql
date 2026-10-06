package repository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type userCommandRepository struct {
	coll *mongo.Collection
}

func NewUserCommandRepository(coll *mongo.Collection) UserCommandRepository {
	return &userCommandRepository{coll: coll}
}

func (r *userCommandRepository) CreateUser(ctx context.Context, request *requests.CreateUserRequest) (*salemongo.User, error) {
	verified := false
	verifyCode := uuid.New().String()

	userID, err := salemongo.NextID(ctx, r.coll.Database(), "users")
	if err != nil {
		return nil, sharedErrors.ErrFailed("create user").WithInternal(err)
	}

	now := time.Now()
	user := &salemongo.User{
		UserID:           userID,
		Firstname:        request.FirstName,
		Lastname:         request.LastName,
		Email:            request.Email,
		Password:         request.Password,
		VerificationCode: verifyCode,
		IsVerified:       &verified,
		CreatedAt:        &now,
		UpdatedAt:        &now,
	}

	result, err := r.coll.InsertOne(ctx, user)
	if err != nil {
		if salemongo.IsDuplicateKeyError(err) {
			return nil, sharedErrors.NewConflictError("email already exists").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("create user").WithInternal(err)
	}
	user.ID = result.InsertedID.(primitive.ObjectID)
	return user, nil
}

func (r *userCommandRepository) UpdateUser(ctx context.Context, request *requests.UpdateUserRequest) (*salemongo.User, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"user_id": *request.UserID, "deleted_at": nil},
		bson.M{"$set": bson.M{
			"firstname":  request.FirstName,
			"lastname":   request.LastName,
			"email":      request.Email,
			"password":   request.Password,
			"updated_at": now,
		}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var user salemongo.User
	if err := result.Decode(&user); err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "user", "update user")
	}
	return &user, nil
}

func (r *userCommandRepository) UpdateIsVerified(ctx context.Context, userID int, isVerified bool) (*salemongo.User, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"user_id": userID, "deleted_at": nil},
		bson.M{"$set": bson.M{"is_verified": isVerified, "updated_at": now}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var user salemongo.User
	if err := result.Decode(&user); err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "user", "update user")
	}
	return &user, nil
}

func (r *userCommandRepository) UpdatePassword(ctx context.Context, userID int, password string) (*salemongo.User, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"user_id": userID, "deleted_at": nil},
		bson.M{"$set": bson.M{"password": password, "updated_at": now}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var user salemongo.User
	if err := result.Decode(&user); err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "user", "update user")
	}
	return &user, nil
}

func (r *userCommandRepository) TrashedUser(ctx context.Context, id int) (*salemongo.User, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"user_id": id, "deleted_at": nil},
		bson.M{"$set": bson.M{"deleted_at": now}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var user salemongo.User
	if err := result.Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, sharedErrors.ErrNotFoundResponse("user")
		}
		return nil, sharedErrors.ErrFailed("trash user").WithInternal(err)
	}
	return &user, nil
}

func (r *userCommandRepository) RestoreUser(ctx context.Context, id int) (*salemongo.User, error) {
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"user_id": id, "deleted_at": bson.M{"$ne": nil}},
		bson.M{"$unset": bson.M{"deleted_at": ""}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var user salemongo.User
	if err := result.Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, sharedErrors.ErrNotFoundResponse("user")
		}
		return nil, sharedErrors.ErrFailed("restore user").WithInternal(err)
	}
	return &user, nil
}

func (r *userCommandRepository) DeleteUserPermanent(ctx context.Context, id int) (bool, error) {
	result, err := r.coll.DeleteMany(ctx, bson.M{"user_id": id})
	if err != nil {
		return false, sharedErrors.ErrFailed("delete user permanently").WithInternal(err)
	}
	if result.DeletedCount == 0 {
		return false, sharedErrors.ErrNotFoundResponse("user")
	}
	return true, nil
}

func (r *userCommandRepository) RestoreAllUser(ctx context.Context) (bool, error) {
	_, err := r.coll.UpdateMany(ctx, bson.M{"deleted_at": bson.M{"$ne": nil}}, bson.M{"$unset": bson.M{"deleted_at": ""}})
	if err != nil {
		return false, sharedErrors.ErrFailed("restore all users").WithInternal(err)
	}
	return true, nil
}

func (r *userCommandRepository) DeleteAllUserPermanent(ctx context.Context) (bool, error) {
	_, err := r.coll.DeleteMany(ctx, bson.M{"deleted_at": bson.M{"$ne": nil}})
	if err != nil {
		return false, sharedErrors.ErrFailed("delete all users permanently").WithInternal(err)
	}
	return true, nil
}
