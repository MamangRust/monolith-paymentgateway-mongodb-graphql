package repository

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type roleQueryRepository struct {
	coll *mongo.Collection
}

func NewRoleQueryRepository(coll *mongo.Collection) RoleQueryRepository {
	return &roleQueryRepository{coll: coll}
}

func (r *roleQueryRepository) FindAllRoles(ctx context.Context, req *requests.FindAllRoles) ([]*salemongo.Role, error) {
	offset := (req.Page - 1) * req.PageSize

	filter := bson.M{"deleted_at": nil}
	if req.Search != "" {
		filter["role_name"] = bson.M{"$regex": req.Search, "$options": "i"}
	}

	cursor, err := r.coll.Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "role_id", Value: 1}}).
			SetSkip(int64(offset)).SetLimit(int64(req.PageSize)))
	if err != nil {
		return nil, sharedErrors.ErrFailed("find all roles").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var roles []*salemongo.Role
	if err := cursor.All(ctx, &roles); err != nil {
		return nil, sharedErrors.ErrFailed("find all roles").WithInternal(err)
	}
	return roles, nil
}

func (r *roleQueryRepository) FindByActiveRole(ctx context.Context, req *requests.FindAllRoles) ([]*salemongo.Role, error) {
	return r.FindAllRoles(ctx, req)
}

func (r *roleQueryRepository) FindByTrashedRole(ctx context.Context, req *requests.FindAllRoles) ([]*salemongo.Role, error) {
	offset := (req.Page - 1) * req.PageSize

	filter := bson.M{"deleted_at": bson.M{"$ne": nil}}
	if req.Search != "" {
		filter["role_name"] = bson.M{"$regex": req.Search, "$options": "i"}
	}

	cursor, err := r.coll.Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "role_id", Value: 1}}).
			SetSkip(int64(offset)).SetLimit(int64(req.PageSize)))
	if err != nil {
		return nil, sharedErrors.ErrFailed("find trashed roles").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var roles []*salemongo.Role
	if err := cursor.All(ctx, &roles); err != nil {
		return nil, sharedErrors.ErrFailed("find trashed roles").WithInternal(err)
	}
	return roles, nil
}

func (r *roleQueryRepository) FindById(ctx context.Context, id int) (*salemongo.Role, error) {
	var role salemongo.Role
	err := r.coll.FindOne(ctx, bson.M{"role_id": id}).Decode(&role)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, sharedErrors.ErrRoleNotFound.WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &role, nil
}

func (r *roleQueryRepository) FindByName(ctx context.Context, name string) (*salemongo.Role, error) {
	var role salemongo.Role
	err := r.coll.FindOne(ctx, bson.M{"role_name": name, "deleted_at": nil}).Decode(&role)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, sharedErrors.ErrRoleNotFound.WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &role, nil
}

func (r *roleQueryRepository) FindByUserId(ctx context.Context, userID int) ([]*salemongo.Role, error) {
	// Find user_roles for the user, then fetch roles
	urColl := r.coll.Database().Collection("user_roles")
	cursor, err := urColl.Find(ctx, bson.M{"user_id": int32(userID)})
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	defer cursor.Close(ctx)

	var userRoles []struct {
		RoleID int32 `bson:"role_id"`
	}
	if err := cursor.All(ctx, &userRoles); err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	var roleIDs []int32
	for _, ur := range userRoles {
		roleIDs = append(roleIDs, ur.RoleID)
	}

	if len(roleIDs) == 0 {
		return nil, nil
	}

	roleCursor, err := r.coll.Find(ctx, bson.M{
		"role_id":    bson.M{"$in": roleIDs},
		"deleted_at": nil,
	})
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	defer roleCursor.Close(ctx)

	var roles []*salemongo.Role
	if err := roleCursor.All(ctx, &roles); err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return roles, nil
}

func (r *roleQueryRepository) CountAllRoles(ctx context.Context, search string) (int64, error) {
	return r.CountActiveRoles(ctx, search)
}

func (r *roleQueryRepository) CountActiveRoles(ctx context.Context, search string) (int64, error) {
	filter := bson.M{"deleted_at": nil}
	if search != "" {
		filter["role_name"] = bson.M{"$regex": search, "$options": "i"}
	}
	count, err := r.coll.CountDocuments(ctx, filter)
	return count, err
}

func (r *roleQueryRepository) CountTrashedRoles(ctx context.Context, search string) (int64, error) {
	filter := bson.M{"deleted_at": bson.M{"$ne": nil}}
	if search != "" {
		filter["role_name"] = bson.M{"$regex": search, "$options": "i"}
	}
	count, err := r.coll.CountDocuments(ctx, filter)
	return count, err
}
