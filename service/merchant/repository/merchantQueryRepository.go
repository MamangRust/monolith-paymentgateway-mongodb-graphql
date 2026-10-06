package repository

import (
	"context"
	"errors"
	"regexp"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type merchantQueryRepository struct {
	coll *mongo.Collection
}

func NewMerchantQueryRepository(coll *mongo.Collection) MerchantQueryRepository {
	return &merchantQueryRepository{coll: coll}
}

func merchantSearchFilter(search string, extra bson.M) bson.M {
	filter := bson.M{}
	if extra != nil {
		for k, v := range extra {
			filter[k] = v
		}
	}
	if search != "" {
		filter["name"] = bson.M{"$regex": regexp.QuoteMeta(search), "$options": "i"}
	}
	return filter
}

func (r *merchantQueryRepository) FindAllMerchants(ctx context.Context, req *requests.FindAllMerchants) ([]*salemongo.Merchant, error) {
	return r.findPage(ctx, bson.M{"deleted_at": nil}, req.Search, req.Page, req.PageSize)
}

func (r *merchantQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllMerchants) ([]*salemongo.Merchant, error) {
	return r.findPage(ctx, bson.M{"deleted_at": nil}, req.Search, req.Page, req.PageSize)
}

func (r *merchantQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllMerchants) ([]*salemongo.Merchant, error) {
	return r.findPage(ctx, bson.M{"deleted_at": bson.M{"$ne": nil}}, req.Search, req.Page, req.PageSize)
}

func (r *merchantQueryRepository) findPage(ctx context.Context, filter bson.M, search string, page, pageSize int) ([]*salemongo.Merchant, error) {
	if search != "" {
		filter["name"] = bson.M{"$regex": regexp.QuoteMeta(search), "$options": "i"}
	}
	opts := options.Find().
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize)).
		SetSort(bson.D{{Key: "merchant_id", Value: 1}})

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, sharedErrors.ErrFailed("find merchants").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var merchants []*salemongo.Merchant
	if err := cursor.All(ctx, &merchants); err != nil {
		return nil, sharedErrors.ErrFailed("decode merchants").WithInternal(err)
	}
	return merchants, nil
}

func (r *merchantQueryRepository) FindByMerchantId(ctx context.Context, id int) (*salemongo.Merchant, error) {
	var merchant salemongo.Merchant
	err := r.coll.FindOne(ctx, bson.M{"merchant_id": id}).Decode(&merchant)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("merchant").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &merchant, nil
}

func (r *merchantQueryRepository) FindByApiKey(ctx context.Context, apiKey string) (*salemongo.Merchant, error) {
	var merchant salemongo.Merchant
	err := r.coll.FindOne(ctx, bson.M{"api_key": apiKey}).Decode(&merchant)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("merchant").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &merchant, nil
}

func (r *merchantQueryRepository) FindByName(ctx context.Context, name string) (*salemongo.Merchant, error) {
	var merchant salemongo.Merchant
	err := r.coll.FindOne(ctx, bson.M{"name": name}).Decode(&merchant)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("merchant").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &merchant, nil
}

func (r *merchantQueryRepository) FindByMerchantUserId(ctx context.Context, userID int) ([]*salemongo.Merchant, error) {
	cursor, err := r.coll.Find(ctx, bson.M{"user_id": userID, "deleted_at": nil})
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	defer cursor.Close(ctx)

	var merchants []*salemongo.Merchant
	if err := cursor.All(ctx, &merchants); err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return merchants, nil
}

func (r *merchantQueryRepository) CountAllMerchants(ctx context.Context, search string) (int64, error) {
	return r.count(ctx, bson.M{"deleted_at": nil}, search)
}

func (r *merchantQueryRepository) CountActiveMerchants(ctx context.Context, search string) (int64, error) {
	return r.count(ctx, bson.M{"deleted_at": nil}, search)
}

func (r *merchantQueryRepository) CountTrashedMerchants(ctx context.Context, search string) (int64, error) {
	return r.count(ctx, bson.M{"deleted_at": bson.M{"$ne": nil}}, search)
}

func (r *merchantQueryRepository) count(ctx context.Context, filter bson.M, search string) (int64, error) {
	if search != "" {
		filter["name"] = bson.M{"$regex": regexp.QuoteMeta(search), "$options": "i"}
	}
	count, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, sharedErrors.ErrFailed("count merchants").WithInternal(err)
	}
	return count, nil
}
