package repository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/randomvcc"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type cardCommandRepository struct {
	coll *mongo.Collection
}

func NewCardCommandRepository(coll *mongo.Collection) CardCommandRepository {
	return &cardCommandRepository{coll: coll}
}

func (r *cardCommandRepository) CreateCard(ctx context.Context, request *requests.CreateCardRequest) (*salemongo.Card, error) {
	number, err := randomvcc.RandomCardNumber()
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	cardID, err := salemongo.NextID(ctx, r.coll.Database(), "cards")
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	now := time.Now()
	card := &salemongo.Card{
		CardID:       cardID,
		UserID:       int32(request.UserID),
		CardNumber:   number,
		CardType:     request.CardType,
		ExpireDate:   request.ExpireDate,
		Cvv:          request.CVV,
		CardProvider: request.CardProvider,
		Status:       "active",
		CreditLimit:  50000000, // Default 50M
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}

	result, err := r.coll.InsertOne(ctx, card)
	if err != nil {
		if salemongo.IsDuplicateKeyError(err) {
			return nil, sharedErrors.NewConflictError("card number already exists").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("create card").WithInternal(err)
	}
	card.ID = result.InsertedID.(primitive.ObjectID)
	return card, nil
}

func (r *cardCommandRepository) UpdateCard(ctx context.Context, request *requests.UpdateCardRequest) (*salemongo.Card, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"card_id": request.CardID, "deleted_at": nil},
		bson.M{"$set": bson.M{
			"card_type":     request.CardType,
			"expire_date":   request.ExpireDate,
			"cvv":           request.CVV,
			"card_provider": request.CardProvider,
			"updated_at":    now,
		}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var card salemongo.Card
	if err := result.Decode(&card); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, sharedErrors.ErrNotFoundResponse("card").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("update card").WithInternal(err)
	}
	return &card, nil
}

func (r *cardCommandRepository) TrashedCard(ctx context.Context, cardID int) (*salemongo.Card, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"card_id": cardID, "deleted_at": nil},
		bson.M{"$set": bson.M{"deleted_at": now}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var card salemongo.Card
	if err := result.Decode(&card); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, sharedErrors.ErrNotFoundResponse("card")
		}
		return nil, sharedErrors.ErrFailed("trash card").WithInternal(err)
	}
	return &card, nil
}

func (r *cardCommandRepository) RestoreCard(ctx context.Context, cardID int) (*salemongo.Card, error) {
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"card_id": cardID, "deleted_at": bson.M{"$ne": nil}},
		bson.M{"$unset": bson.M{"deleted_at": ""}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var card salemongo.Card
	if err := result.Decode(&card); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, sharedErrors.ErrNotFoundResponse("card")
		}
		return nil, sharedErrors.ErrFailed("restore card").WithInternal(err)
	}
	return &card, nil
}

func (r *cardCommandRepository) DeleteCardPermanent(ctx context.Context, cardID int) (bool, error) {
	result, err := r.coll.DeleteMany(ctx, bson.M{"card_id": cardID})
	if err != nil {
		return false, sharedErrors.ErrFailed("delete card permanently").WithInternal(err)
	}
	if result.DeletedCount == 0 {
		return false, sharedErrors.ErrNotFoundResponse("card")
	}
	return true, nil
}

func (r *cardCommandRepository) RestoreAllCard(ctx context.Context) (bool, error) {
	_, err := r.coll.UpdateMany(ctx, bson.M{"deleted_at": bson.M{"$ne": nil}}, bson.M{"$unset": bson.M{"deleted_at": ""}})
	if err != nil {
		return false, sharedErrors.ErrFailed("restore all cards").WithInternal(err)
	}
	return true, nil
}

func (r *cardCommandRepository) DeleteAllCardPermanent(ctx context.Context) (bool, error) {
	_, err := r.coll.DeleteMany(ctx, bson.M{"deleted_at": bson.M{"$ne": nil}})
	if err != nil {
		return false, sharedErrors.ErrFailed("delete all cards permanently").WithInternal(err)
	}
	return true, nil
}

func (r *cardCommandRepository) ToggleCardStatus(ctx context.Context, request *requests.ToggleCardStatusRequest) (*salemongo.Card, error) {
	var card salemongo.Card
	err := r.coll.FindOne(ctx, bson.M{"card_id": request.CardID, "deleted_at": nil}).Decode(&card)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, sharedErrors.ErrNotFoundResponse("card").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("toggle card status").WithInternal(err)
	}

	newStatus := "active"
	if card.Status == "active" {
		newStatus = "inactive"
	}

	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"card_id": request.CardID},
		bson.M{"$set": bson.M{"status": newStatus, "updated_at": time.Now()}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	if err := result.Decode(&card); err != nil {
		return nil, sharedErrors.ErrFailed("toggle card status").WithInternal(err)
	}
	return &card, nil
}

func (r *cardCommandRepository) UpdateCreditLimit(ctx context.Context, request *requests.UpdateCreditLimitRequest) (*salemongo.Card, error) {
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"card_id": request.CardID, "deleted_at": nil},
		bson.M{"$set": bson.M{"credit_limit": int32(request.CreditLimit), "updated_at": time.Now()}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var card salemongo.Card
	if err := result.Decode(&card); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, sharedErrors.ErrNotFoundResponse("card").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("update credit limit").WithInternal(err)
	}
	return &card, nil
}

func (r *cardCommandRepository) RedeemPoints(ctx context.Context, request *requests.RedeemPointsRequest) (*salemongo.Card, error) {
	var card salemongo.Card
	err := r.coll.FindOne(ctx, bson.M{"card_id": request.CardID, "deleted_at": nil}).Decode(&card)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, sharedErrors.ErrNotFoundResponse("card").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("redeem reward points").WithInternal(err)
	}

	if card.RewardPoints < int32(request.Points) {
		return nil, sharedErrors.NewBadRequestError("insufficient reward points")
	}

	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"card_id": request.CardID},
		bson.M{"$inc": bson.M{"reward_points": -int32(request.Points)}, "$set": bson.M{"updated_at": time.Now()}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	if err := result.Decode(&card); err != nil {
		return nil, sharedErrors.ErrFailed("redeem reward points").WithInternal(err)
	}
	return &card, nil
}
