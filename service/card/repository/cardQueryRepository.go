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

type cardQueryRepository struct {
	coll *mongo.Collection
}

func NewCardQueryRepository(coll *mongo.Collection) CardQueryRepository {
	return &cardQueryRepository{coll: coll}
}

func cardSearchFilter(search string) bson.M {
	filter := bson.M{"deleted_at": nil}
	if search != "" {
		filter["$or"] = []bson.M{
			{"card_number": bson.M{"$regex": search, "$options": "i"}},
			{"card_type": bson.M{"$regex": search, "$options": "i"}},
		}
	}
	return filter
}

func (r *cardQueryRepository) FindAllCards(ctx context.Context, req *requests.FindAllCards) ([]*salemongo.Card, error) {
	offset := (req.Page - 1) * req.PageSize

	cursor, err := r.coll.Find(ctx, cardSearchFilter(req.Search),
		options.Find().SetSort(bson.D{{Key: "card_id", Value: 1}}).
			SetSkip(int64(offset)).SetLimit(int64(req.PageSize)))
	if err != nil {
		return nil, sharedErrors.ErrFailed("find all cards").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var cards []*salemongo.Card
	if err := cursor.All(ctx, &cards); err != nil {
		return nil, sharedErrors.ErrFailed("find all cards").WithInternal(err)
	}
	return cards, nil
}

func (r *cardQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllCards) ([]*salemongo.Card, error) {
	return r.FindAllCards(ctx, req)
}

func (r *cardQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllCards) ([]*salemongo.Card, error) {
	offset := (req.Page - 1) * req.PageSize

	filter := bson.M{"deleted_at": bson.M{"$ne": nil}}
	if req.Search != "" {
		filter["$or"] = []bson.M{
			{"card_number": bson.M{"$regex": req.Search, "$options": "i"}},
			{"card_type": bson.M{"$regex": req.Search, "$options": "i"}},
		}
	}

	cursor, err := r.coll.Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "card_id", Value: 1}}).
			SetSkip(int64(offset)).SetLimit(int64(req.PageSize)))
	if err != nil {
		return nil, sharedErrors.ErrFailed("find trashed cards").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var cards []*salemongo.Card
	if err := cursor.All(ctx, &cards); err != nil {
		return nil, sharedErrors.ErrFailed("find trashed cards").WithInternal(err)
	}
	return cards, nil
}

func (r *cardQueryRepository) FindById(ctx context.Context, cardID int) (*salemongo.Card, error) {
	var card salemongo.Card
	err := r.coll.FindOne(ctx, bson.M{"card_id": cardID, "deleted_at": nil}).Decode(&card)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, sharedErrors.ErrNotFoundResponse("card").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &card, nil
}

func (r *cardQueryRepository) FindCardByUserId(ctx context.Context, userID int) (*salemongo.Card, error) {
	var card salemongo.Card
	err := r.coll.FindOne(ctx, bson.M{"user_id": int32(userID), "deleted_at": nil}).Decode(&card)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, sharedErrors.ErrNotFoundResponse("card").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &card, nil
}

func (r *cardQueryRepository) FindCardByCardNumber(ctx context.Context, cardNumber string) (*salemongo.Card, error) {
	var card salemongo.Card
	err := r.coll.FindOne(ctx, bson.M{"card_number": cardNumber, "deleted_at": nil}).Decode(&card)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, sharedErrors.ErrNotFoundResponse("card").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &card, nil
}

func (r *cardQueryRepository) FindUserCardByCardNumber(ctx context.Context, cardNumber string) (*salemongo.Card, error) {
	return r.FindCardByCardNumber(ctx, cardNumber)
}

func (r *cardQueryRepository) CountAllCards(ctx context.Context, search string) (int64, error) {
	return r.CountActiveCards(ctx, search)
}

func (r *cardQueryRepository) CountActiveCards(ctx context.Context, search string) (int64, error) {
	count, err := r.coll.CountDocuments(ctx, cardSearchFilter(search))
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *cardQueryRepository) CountTrashedCards(ctx context.Context, search string) (int64, error) {
	filter := bson.M{"deleted_at": bson.M{"$ne": nil}}
	if search != "" {
		filter["$or"] = []bson.M{
			{"card_number": bson.M{"$regex": search, "$options": "i"}},
			{"card_type": bson.M{"$regex": search, "$options": "i"}},
		}
	}
	count, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}
