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

type topupQueryRepository struct {
	coll *mongo.Collection
}

func NewTopupQueryRepository(coll *mongo.Collection) TopupQueryRepository {
	return &topupQueryRepository{coll: coll}
}

// topupSearchFilter builds the OR-search across card_number/topup_method/status.
func topupSearchFilter(search string) bson.M {
	rx := bson.M{"$regex": regexp.QuoteMeta(search), "$options": "i"}
	return bson.M{"$or": bson.A{
		bson.M{"card_number": rx},
		bson.M{"topup_method": rx},
		bson.M{"status": rx},
	}}
}

func topupCardSearchFilter(search string) bson.M {
	rx := bson.M{"$regex": regexp.QuoteMeta(search), "$options": "i"}
	return bson.M{"$or": bson.A{
		bson.M{"topup_method": rx},
		bson.M{"status": rx},
	}}
}

func (r *topupQueryRepository) findPage(ctx context.Context, filter bson.M, page, pageSize int) ([]*salemongo.Topup, int64, error) {
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, sharedErrors.ErrFailed("count topups").WithInternal(err)
	}

	opts := options.Find().
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize)).
		SetSort(bson.D{{Key: "topup_time", Value: -1}})

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, sharedErrors.ErrFailed("find topups").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var topups []*salemongo.Topup
	if err := cursor.All(ctx, &topups); err != nil {
		return nil, 0, sharedErrors.ErrFailed("decode topups").WithInternal(err)
	}
	return topups, total, nil
}

func (r *topupQueryRepository) FindAllTopups(ctx context.Context, req *requests.FindAllTopups) ([]*TopupQueryResult, error) {
	filter := bson.M{}
	if req.Search != "" {
		filter = topupSearchFilter(req.Search)
	}
	topups, total, err := r.findPage(ctx, filter, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	results := make([]*TopupQueryResult, len(topups))
	for i, t := range topups {
		results[i] = &TopupQueryResult{Topup: t, TotalCount: total}
	}
	return results, nil
}

func (r *topupQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllTopups) ([]*TopupQueryResult, error) {
	filter := bson.M{"deleted_at": nil}
	if req.Search != "" {
		filter = topupSearchFilter(req.Search)
		filter["deleted_at"] = nil
	}
	topups, total, err := r.findPage(ctx, filter, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	results := make([]*TopupQueryResult, len(topups))
	for i, t := range topups {
		results[i] = &TopupQueryResult{Topup: t, TotalCount: total}
	}
	return results, nil
}

func (r *topupQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllTopups) ([]*TopupQueryResult, error) {
	filter := bson.M{"deleted_at": bson.M{"$ne": nil}}
	if req.Search != "" {
		filter = topupSearchFilter(req.Search)
		filter["deleted_at"] = bson.M{"$ne": nil}
	}
	topups, total, err := r.findPage(ctx, filter, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	results := make([]*TopupQueryResult, len(topups))
	for i, t := range topups {
		results[i] = &TopupQueryResult{Topup: t, TotalCount: total}
	}
	return results, nil
}

func (r *topupQueryRepository) FindAllTopupByCardNumber(ctx context.Context, req *requests.FindAllTopupsByCardNumber) ([]*TopupByCardResult, error) {
	filter := bson.M{"card_number": req.CardNumber}
	if req.Search != "" {
		filter = topupCardSearchFilter(req.Search)
		filter["card_number"] = req.CardNumber
	}
	topups, total, err := r.findPage(ctx, filter, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	results := make([]*TopupByCardResult, len(topups))
	for i, t := range topups {
		results[i] = &TopupByCardResult{Topup: t, TotalCount: total}
	}
	return results, nil
}

func (r *topupQueryRepository) FindById(ctx context.Context, topup_id int) (*salemongo.Topup, error) {
	var topup salemongo.Topup
	err := r.coll.FindOne(ctx, bson.M{"topup_id": int32(topup_id)}).Decode(&topup)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("topup").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &topup, nil
}
