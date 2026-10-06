package repository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRoleCommandRepository interface {
	AssignRoleToUser(ctx context.Context, userID, roleID int) (*salemongo.UserRole, error)
	RemoveRoleFromUser(ctx context.Context, userID, roleID int) (bool, error)
}

type userRoleCommandRepository struct {
	coll *mongo.Collection
}

func NewUserRoleCommandRepository(coll *mongo.Collection) UserRoleCommandRepository {
	return &userRoleCommandRepository{coll: coll}
}

func (r *userRoleCommandRepository) AssignRoleToUser(ctx context.Context, userID, roleID int) (*salemongo.UserRole, error) {
	userRoleID, err := salemongo.NextID(ctx, r.coll.Database(), "user_roles")
	if err != nil {
		return nil, sharedErrors.ErrFailed("create user role").WithInternal(err)
	}

	now := time.Now()
	userRole := &salemongo.UserRole{
		UserRoleID: userRoleID,
		UserID:     int32(userID),
		RoleID:     int32(roleID),
		CreatedAt:  &now,
		UpdatedAt:  &now,
	}

	if _, err := r.coll.Database().Collection("user_roles").InsertOne(ctx, userRole); err != nil {
		return nil, sharedErrors.ErrFailed("create user role").WithInternal(err)
	}
	return userRole, nil
}

func (r *userRoleCommandRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID int) (bool, error) {
	result, err := r.coll.Database().Collection("user_roles").DeleteOne(ctx, bson.M{
		"user_id": int32(userID),
		"role_id": int32(roleID),
	})
	if err != nil {
		return false, sharedErrors.ErrFailed("delete user role").WithInternal(err)
	}
	return result.DeletedCount > 0, nil
}
