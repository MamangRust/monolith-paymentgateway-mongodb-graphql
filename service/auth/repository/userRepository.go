package repository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type userRepository struct {
	db *mongo.Database
}

func NewUserRepository(db *mongo.Database) *userRepository {
	return &userRepository{db: db}
}

func (r *userRepository) coll() *mongo.Collection {
	return r.db.Collection("users")
}

func (r *userRepository) FindById(ctx context.Context, userID int) (*salemongo.User, error) {
	var user salemongo.User
	err := r.coll().FindOne(ctx, bson.M{"user_id": int32(userID)}).Decode(&user)
	if err != nil {
		return nil, sharedErrors.ErrUserNotFound.WithInternal(err)
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*salemongo.User, error) {
	var user salemongo.User
	err := r.coll().FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, sharedErrors.ErrUserNotFound.WithInternal(err)
	}
	return &user, nil
}

func (r *userRepository) FindByEmailAndVerify(ctx context.Context, email string) (*salemongo.User, error) {
	var user salemongo.User
	err := r.coll().FindOne(ctx, bson.M{
		"email": email,
		"$or": []bson.M{
			{"is_verified": true},
			{"is_verified": bson.M{"$exists": false}},
		},
	}).Decode(&user)
	if err != nil {
		return nil, sharedErrors.ErrUserNotFound.WithInternal(err)
	}
	return &user, nil
}

func (r *userRepository) FindByVerificationCode(ctx context.Context, verificationCode string) (*salemongo.User, error) {
	var user salemongo.User
	err := r.coll().FindOne(ctx, bson.M{"verification_code": verificationCode}).Decode(&user)
	if err != nil {
		return nil, sharedErrors.ErrUserNotFound.WithInternal(err)
	}
	return &user, nil
}

func (r *userRepository) CreateUser(ctx context.Context, request *requests.RegisterRequest) (*salemongo.User, error) {
	// Check if email already exists
	count, err := r.coll().CountDocuments(ctx, bson.M{"email": request.Email})
	if err != nil {
		return nil, sharedErrors.ErrFailed("create user").WithInternal(err)
	}
	if count > 0 {
		return nil, sharedErrors.NewConflictError("email already exists")
	}

	userID, err := salemongo.NextID(ctx, r.coll().Database(), "users")
	if err != nil {
		return nil, sharedErrors.ErrFailed("create user").WithInternal(err)
	}

	now := time.Now()
	user := &salemongo.User{
		UserID:    userID,
		Firstname: request.FirstName,
		Lastname:  request.LastName,
		Email:     request.Email,
		Password:  request.Password,
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	_, err = r.coll().InsertOne(ctx, user)
	if err != nil {
		return nil, sharedErrors.ErrFailed("create user").WithInternal(err)
	}
	return user, nil
}

func (r *userRepository) UpdateUserIsVerified(ctx context.Context, userID int, isVerified bool) (*salemongo.User, error) {
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var user salemongo.User
	err := r.coll().FindOneAndUpdate(
		ctx,
		bson.M{"user_id": int32(userID)},
		bson.M{"$set": bson.M{"is_verified": isVerified}},
		opts,
	).Decode(&user)
	if err != nil {
		return nil, sharedErrors.ErrFailed("update user verification").WithInternal(err)
	}
	return &user, nil
}

func (r *userRepository) UpdateUserPassword(ctx context.Context, userID int, password string) (*salemongo.User, error) {
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var user salemongo.User
	err := r.coll().FindOneAndUpdate(
		ctx,
		bson.M{"user_id": int32(userID)},
		bson.M{"$set": bson.M{"password": password}},
		opts,
	).Decode(&user)
	if err != nil {
		return nil, sharedErrors.ErrFailed("update user password").WithInternal(err)
	}
	return &user, nil
}
