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

type transferQueryRepository struct {
	coll *mongo.Collection
}

func NewTransferQueryRepository(coll *mongo.Collection) TransferQueryRepository {
	return &transferQueryRepository{coll: coll}
}

func transferSearchFilter(search string) bson.M {
	rx := bson.M{"$regex": regexp.QuoteMeta(search), "$options": "i"}
	return bson.M{"$or": bson.A{
		bson.M{"transfer_from": rx},
		bson.M{"transfer_to": rx},
		bson.M{"status": rx},
	}}
}

func (r *transferQueryRepository) findPage(ctx context.Context, filter bson.M, page, pageSize int) ([]*salemongo.Transfer, int64, error) {
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, sharedErrors.ErrFailed("count transfers").WithInternal(err)
	}

	opts := options.Find().
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize)).
		SetSort(bson.D{{Key: "transfer_id", Value: 1}})

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, sharedErrors.ErrFailed("find transfers").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var transfers []*salemongo.Transfer
	if err := cursor.All(ctx, &transfers); err != nil {
		return nil, 0, sharedErrors.ErrFailed("decode transfers").WithInternal(err)
	}
	return transfers, total, nil
}

func (r *transferQueryRepository) FindAll(ctx context.Context, req *requests.FindAllTransfers) ([]*TransferQueryResult, error) {
	filter := bson.M{}
	if req.Search != "" {
		filter = transferSearchFilter(req.Search)
	}
	transfers, total, err := r.findPage(ctx, filter, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	results := make([]*TransferQueryResult, len(transfers))
	for i, t := range transfers {
		results[i] = &TransferQueryResult{Transfer: t, TotalCount: total}
	}
	return results, nil
}

func (r *transferQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllTransfers) ([]*TransferQueryResult, error) {
	filter := bson.M{"deleted_at": nil}
	if req.Search != "" {
		filter = transferSearchFilter(req.Search)
		filter["deleted_at"] = nil
	}
	transfers, total, err := r.findPage(ctx, filter, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	results := make([]*TransferQueryResult, len(transfers))
	for i, t := range transfers {
		results[i] = &TransferQueryResult{Transfer: t, TotalCount: total}
	}
	return results, nil
}

func (r *transferQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllTransfers) ([]*TransferQueryResult, error) {
	filter := bson.M{"deleted_at": bson.M{"$ne": nil}}
	if req.Search != "" {
		filter = transferSearchFilter(req.Search)
		filter["deleted_at"] = bson.M{"$ne": nil}
	}
	transfers, total, err := r.findPage(ctx, filter, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	results := make([]*TransferQueryResult, len(transfers))
	for i, t := range transfers {
		results[i] = &TransferQueryResult{Transfer: t, TotalCount: total}
	}
	return results, nil
}

func (r *transferQueryRepository) FindById(ctx context.Context, id int) (*salemongo.Transfer, error) {
	var transfer salemongo.Transfer
	err := r.coll.FindOne(ctx, bson.M{"transfer_id": id}).Decode(&transfer)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("transfer").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &transfer, nil
}

func (r *transferQueryRepository) FindTransferByTransferFrom(ctx context.Context, transfer_from string) ([]*salemongo.Transfer, error) {
	cursor, err := r.coll.Find(ctx, bson.M{"transfer_from": transfer_from})
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	defer cursor.Close(ctx)

	var transfers []*salemongo.Transfer
	if err := cursor.All(ctx, &transfers); err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return transfers, nil
}

func (r *transferQueryRepository) FindTransferByTransferTo(ctx context.Context, transfer_to string) ([]*salemongo.Transfer, error) {
	cursor, err := r.coll.Find(ctx, bson.M{"transfer_to": transfer_to})
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	defer cursor.Close(ctx)

	var transfers []*salemongo.Transfer
	if err := cursor.All(ctx, &transfers); err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return transfers, nil
}
