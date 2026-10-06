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

type userQueryRepository struct {
	coll *mongo.Collection
}

func NewUserQueryRepository(coll *mongo.Collection) UserQueryRepository {
	return &userQueryRepository{coll: coll}
}

func searchFilter(search string) bson.M {
	filter := bson.M{"deleted_at": nil}
	if search != "" {
		filter["$or"] = []bson.M{
			{"firstname": bson.M{"$regex": search, "$options": "i"}},
			{"lastname": bson.M{"$regex": search, "$options": "i"}},
			{"email": bson.M{"$regex": search, "$options": "i"}},
		}
	}
	return filter
}

func (r *userQueryRepository) FindAllUsers(ctx context.Context, req *requests.FindAllUsers) ([]*salemongo.User, error) {
	offset := (req.Page - 1) * req.PageSize

	cursor, err := r.coll.Find(ctx, searchFilter(req.Search),
		options.Find().SetSort(bson.D{{Key: "user_id", Value: 1}}).
			SetSkip(int64(offset)).SetLimit(int64(req.PageSize)))
	if err != nil {
		return nil, sharedErrors.ErrFailed("find all users").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var users []*salemongo.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, sharedErrors.ErrFailed("find all users").WithInternal(err)
	}
	return users, nil
}

func (r *userQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllUsers) ([]*salemongo.User, error) {
	return r.FindAllUsers(ctx, req)
}

func (r *userQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllUsers) ([]*salemongo.User, error) {
	offset := (req.Page - 1) * req.PageSize

	filter := bson.M{"deleted_at": bson.M{"$ne": nil}}
	if req.Search != "" {
		filter["$or"] = []bson.M{
			{"firstname": bson.M{"$regex": req.Search, "$options": "i"}},
			{"lastname": bson.M{"$regex": req.Search, "$options": "i"}},
			{"email": bson.M{"$regex": req.Search, "$options": "i"}},
		}
	}

	cursor, err := r.coll.Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "user_id", Value: 1}}).
			SetSkip(int64(offset)).SetLimit(int64(req.PageSize)))
	if err != nil {
		return nil, sharedErrors.ErrFailed("find trashed users").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var users []*salemongo.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, sharedErrors.ErrFailed("find trashed users").WithInternal(err)
	}
	return users, nil
}

func (r *userQueryRepository) FindById(ctx context.Context, userID int) (*salemongo.User, error) {
	var user salemongo.User
	err := r.coll.FindOne(ctx, bson.M{"user_id": userID, "deleted_at": nil}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, sharedErrors.ErrNotFound.WithMessage("user not found")
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &user, nil
}

func (r *userQueryRepository) FindByEmail(ctx context.Context, email string) (*salemongo.User, error) {
	var user salemongo.User
	err := r.coll.FindOne(ctx, bson.M{"email": email, "deleted_at": nil}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, sharedErrors.ErrNotFound.WithMessage("user not found")
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &user, nil
}

func (r *userQueryRepository) FindByVerificationCode(ctx context.Context, code string) (*salemongo.User, error) {
	var user salemongo.User
	err := r.coll.FindOne(ctx, bson.M{"verification_code": code}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, sharedErrors.ErrNotFound.WithMessage("user not found")
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &user, nil
}

func (r *userQueryRepository) CountAllUsers(ctx context.Context, search string) (int64, error) {
	return r.CountActiveUsers(ctx, search)
}

func (r *userQueryRepository) CountActiveUsers(ctx context.Context, search string) (int64, error) {
	return r.coll.CountDocuments(ctx, searchFilter(search))
}

func (r *userQueryRepository) CountTrashedUsers(ctx context.Context, search string) (int64, error) {
	filter := bson.M{"deleted_at": bson.M{"$ne": nil}}
	if search != "" {
		filter["$or"] = []bson.M{
			{"firstname": bson.M{"$regex": search, "$options": "i"}},
			{"lastname": bson.M{"$regex": search, "$options": "i"}},
			{"email": bson.M{"$regex": search, "$options": "i"}},
		}
	}
	return r.coll.CountDocuments(ctx, filter)
}
