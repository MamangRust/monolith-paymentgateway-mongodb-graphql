package repository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type roleCommandRepository struct {
	coll *mongo.Collection
}

func NewRoleCommandRepository(coll *mongo.Collection) RoleCommandRepository {
	return &roleCommandRepository{coll: coll}
}

func (r *roleCommandRepository) CreateRole(ctx context.Context, req *requests.CreateRoleRequest) (*salemongo.Role, error) {
	roleID, err := salemongo.NextID(ctx, r.coll.Database(), "roles")
	if err != nil {
		return nil, sharedErrors.ErrFailed("create role").WithInternal(err)
	}

	now := time.Now()
	role := &salemongo.Role{
		RoleID:    roleID,
		RoleName:  req.Name,
		CreatedAt: &now,
		UpdatedAt: &now,
	}
	result, err := r.coll.InsertOne(ctx, role)
	if err != nil {
		return nil, sharedErrors.ErrFailed("create role").WithInternal(err)
	}
	role.ID = result.InsertedID.(primitive.ObjectID)
	return role, nil
}

func (r *roleCommandRepository) UpdateRole(ctx context.Context, req *requests.UpdateRoleRequest) (*salemongo.Role, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"role_id": *req.ID, "deleted_at": nil},
		bson.M{"$set": bson.M{"role_name": req.Name, "updated_at": now}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var role salemongo.Role
	if err := result.Decode(&role); err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "role", "update role")
	}
	return &role, nil
}

func (r *roleCommandRepository) TrashedRole(ctx context.Context, id int) (*salemongo.Role, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"role_id": id, "deleted_at": nil},
		bson.M{"$set": bson.M{"deleted_at": now}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var role salemongo.Role
	if err := result.Decode(&role); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, sharedErrors.ErrNotFoundResponse("role")
		}
		return nil, sharedErrors.ErrFailed("trash role").WithInternal(err)
	}
	return &role, nil
}

func (r *roleCommandRepository) RestoreRole(ctx context.Context, id int) (*salemongo.Role, error) {
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"role_id": id, "deleted_at": bson.M{"$ne": nil}},
		bson.M{"$unset": bson.M{"deleted_at": ""}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var role salemongo.Role
	if err := result.Decode(&role); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, sharedErrors.ErrNotFoundResponse("role")
		}
		return nil, sharedErrors.ErrFailed("restore role").WithInternal(err)
	}
	return &role, nil
}

func (r *roleCommandRepository) DeleteRolePermanent(ctx context.Context, id int) (bool, error) {
	result, err := r.coll.DeleteMany(ctx, bson.M{"role_id": id})
	if err != nil {
		return false, sharedErrors.ErrFailed("delete role permanently").WithInternal(err)
	}
	if result.DeletedCount == 0 {
		return false, sharedErrors.ErrNotFoundResponse("role")
	}
	return true, nil
}

func (r *roleCommandRepository) RestoreAllRole(ctx context.Context) (bool, error) {
	_, err := r.coll.UpdateMany(ctx, bson.M{"deleted_at": bson.M{"$ne": nil}}, bson.M{"$unset": bson.M{"deleted_at": ""}})
	if err != nil {
		return false, sharedErrors.ErrFailed("restore all roles").WithInternal(err)
	}
	return true, nil
}

func (r *roleCommandRepository) DeleteAllRolePermanent(ctx context.Context) (bool, error) {
	_, err := r.coll.DeleteMany(ctx, bson.M{"deleted_at": bson.M{"$ne": nil}})
	if err != nil {
		return false, sharedErrors.ErrFailed("delete all roles permanently").WithInternal(err)
	}
	return true, nil
}

func (r *roleCommandRepository) roleByID(ctx context.Context, roleID int) (*salemongo.Role, error) {
	var role salemongo.Role
	err := r.coll.FindOne(ctx, bson.M{"role_id": roleID}).Decode(&role)
	if err != nil {
		return nil, sharedErrors.ErrRoleNotFound.WithInternal(err)
	}
	return &role, nil
}
