package repository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	refreshtoken_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/refresh_token_errors/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type refreshTokenRepository struct {
	db *mongo.Database
}

func NewRefreshTokenRepository(db *mongo.Database) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) coll() *mongo.Collection {
	return r.db.Collection("refresh_tokens")
}

func (r *refreshTokenRepository) FindByToken(ctx context.Context, token string) (*salemongo.RefreshToken, error) {
	var res salemongo.RefreshToken
	err := r.coll().FindOne(ctx, bson.M{"token": token}).Decode(&res)
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "Refresh Token", "find refresh token by token")
	}
	return &res, nil
}

func (r *refreshTokenRepository) FindByUserId(ctx context.Context, user_id int) (*salemongo.RefreshToken, error) {
	var res salemongo.RefreshToken
	err := r.coll().FindOne(ctx, bson.M{"user_id": int32(user_id)}).Decode(&res)
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "Refresh Token", "find refresh token by user ID")
	}
	return &res, nil
}

func (r *refreshTokenRepository) CreateRefreshToken(ctx context.Context, req *requests.CreateRefreshToken) (*salemongo.RefreshToken, error) {
	layout := "2006-01-02 15:04:05"
	expirationTime, err := time.Parse(layout, req.ExpiresAt)
	if err != nil {
		return nil, refreshtoken_errors.ErrParseDate.WithInternal(err)
	}

	refreshTokenID, err := salemongo.NextID(ctx, r.coll().Database(), "refresh_tokens")
	if err != nil {
		return nil, refreshtoken_errors.ErrCreateRefreshToken.WithInternal(err)
	}

	now := time.Now()
	res := &salemongo.RefreshToken{
		RefreshTokenID: refreshTokenID,
		UserID:         int32(req.UserId),
		Token:          req.Token,
		Expiration:     expirationTime,
		CreatedAt:      &now,
		UpdatedAt:      &now,
	}

	_, err = r.coll().InsertOne(ctx, res)
	if err != nil {
		return nil, refreshtoken_errors.ErrCreateRefreshToken.WithInternal(err)
	}
	return res, nil
}

func (r *refreshTokenRepository) UpdateRefreshToken(ctx context.Context, req *requests.UpdateRefreshToken) (*salemongo.RefreshToken, error) {
	layout := "2006-01-02 15:04:05"
	expirationTime, err := time.Parse(layout, req.ExpiresAt)
	if err != nil {
		return nil, refreshtoken_errors.ErrParseDate.WithInternal(err)
	}

	now := time.Now()
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var res salemongo.RefreshToken
	err = r.coll().FindOneAndUpdate(
		ctx,
		bson.M{"user_id": int32(req.UserId)},
		bson.M{"$set": bson.M{
			"token":      req.Token,
			"expiration": expirationTime,
			"updated_at": now,
		}},
	).Decode(&res)
	if err != nil {
		return nil, refreshtoken_errors.ErrUpdateRefreshToken.WithInternal(err)
	}
	_ = opts
	return &res, nil
}

func (r *refreshTokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	_, err := r.coll().DeleteOne(ctx, bson.M{"token": token})
	if err != nil {
		return refreshtoken_errors.ErrDeleteRefreshToken.WithInternal(err)
	}
	return nil
}

func (r *refreshTokenRepository) DeleteRefreshTokenByUserId(ctx context.Context, user_id int) error {
	_, err := r.coll().DeleteMany(ctx, bson.M{"user_id": int32(user_id)})
	if err != nil {
		return refreshtoken_errors.ErrDeleteByUserID.WithInternal(err)
	}
	return nil
}
