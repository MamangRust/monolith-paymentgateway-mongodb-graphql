package repository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type resetTokenRepository struct {
	db *mongo.Database
}

func NewResetTokenRepository(db *mongo.Database) *resetTokenRepository {
	return &resetTokenRepository{db: db}
}

func (r *resetTokenRepository) coll() *mongo.Collection {
	return r.db.Collection("reset_tokens")
}

func (r *resetTokenRepository) FindByToken(ctx context.Context, code string) (*salemongo.ResetToken, error) {
	var rt salemongo.ResetToken
	err := r.coll().FindOne(ctx, bson.M{"token": code}).Decode(&rt)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &rt, nil
}

func (r *resetTokenRepository) CreateResetToken(ctx context.Context, req *requests.CreateResetTokenRequest) (*salemongo.ResetToken, error) {
	expiryDate, err := time.Parse("2006-01-02 15:04:05", req.ExpiredAt)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	resetTokenID, err := salemongo.NextID(ctx, r.coll().Database(), "reset_tokens")
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	rt := &salemongo.ResetToken{
		ID:         resetTokenID,
		UserID:     int64(req.UserID),
		Token:      req.ResetToken,
		ExpiryDate: expiryDate,
	}
	_, err = r.coll().InsertOne(ctx, rt)
	if err != nil {
		return nil, sharedErrors.ErrFailed("create reset token").WithInternal(err)
	}
	return rt, nil
}

func (r *resetTokenRepository) DeleteResetToken(ctx context.Context, userID int32) error {
	_, err := r.coll().DeleteMany(ctx, bson.M{"user_id": int64(userID)})
	if err != nil {
		return sharedErrors.ErrFailed("delete reset token").WithInternal(err)
	}
	return nil
}
