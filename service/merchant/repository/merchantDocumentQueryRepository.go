package repository

import (
	"context"
	"regexp"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type merchantDocumentQueryRepository struct {
	coll *mongo.Collection
}

func NewMerchantDocumentQueryRepository(coll *mongo.Collection) MerchantDocumentQueryRepository {
	return &merchantDocumentQueryRepository{coll: coll}
}

func docSearchFilter(search string, extra bson.M) bson.M {
	filter := bson.M{}
	if extra != nil {
		for k, v := range extra {
			filter[k] = v
		}
	}
	if search != "" {
		filter["document_type"] = bson.M{"$regex": regexp.QuoteMeta(search), "$options": "i"}
	}
	return filter
}

func (r *merchantDocumentQueryRepository) findPage(ctx context.Context, filter bson.M, search string, page, pageSize int) ([]*salemongo.MerchantDocument, error) {
	if search != "" {
		filter["document_type"] = bson.M{"$regex": regexp.QuoteMeta(search), "$options": "i"}
	}
	opts := options.Find().
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize)).
		SetSort(bson.D{{Key: "document_id", Value: 1}})

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, sharedErrors.ErrFailed("find merchant documents").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var docs []*salemongo.MerchantDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, sharedErrors.ErrFailed("decode merchant documents").WithInternal(err)
	}
	return docs, nil
}

func (r *merchantDocumentQueryRepository) FindAllDocuments(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*salemongo.MerchantDocument, error) {
	return r.findPage(ctx, bson.M{"deleted_at": nil}, req.Search, req.Page, req.PageSize)
}

func (r *merchantDocumentQueryRepository) FindByActiveDocuments(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*salemongo.MerchantDocument, error) {
	return r.findPage(ctx, bson.M{"deleted_at": nil}, req.Search, req.Page, req.PageSize)
}

func (r *merchantDocumentQueryRepository) FindByTrashedDocuments(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*salemongo.MerchantDocument, error) {
	return r.findPage(ctx, bson.M{"deleted_at": bson.M{"$ne": nil}}, req.Search, req.Page, req.PageSize)
}

func (r *merchantDocumentQueryRepository) FindByIdDocument(ctx context.Context, id int) (*salemongo.MerchantDocument, error) {
	var doc salemongo.MerchantDocument
	err := r.coll.FindOne(ctx, bson.M{"document_id": id}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, sharedErrors.ErrNotFoundResponse("merchant document").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &doc, nil
}

func (r *merchantDocumentQueryRepository) count(ctx context.Context, filter bson.M, search string) (int64, error) {
	if search != "" {
		filter["document_type"] = bson.M{"$regex": regexp.QuoteMeta(search), "$options": "i"}
	}
	count, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, sharedErrors.ErrFailed("count merchant documents").WithInternal(err)
	}
	return count, nil
}

func (r *merchantDocumentQueryRepository) CountAllDocuments(ctx context.Context, search string) (int64, error) {
	return r.count(ctx, bson.M{"deleted_at": nil}, search)
}

func (r *merchantDocumentQueryRepository) CountActiveDocuments(ctx context.Context, search string) (int64, error) {
	return r.count(ctx, bson.M{"deleted_at": nil}, search)
}

func (r *merchantDocumentQueryRepository) CountTrashedDocuments(ctx context.Context, search string) (int64, error) {
	return r.count(ctx, bson.M{"deleted_at": bson.M{"$ne": nil}}, search)
}
