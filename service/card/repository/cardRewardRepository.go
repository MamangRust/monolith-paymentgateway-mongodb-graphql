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

type cardRewardRepository struct {
	coll *mongo.Collection
}

func NewCardRewardRepository(coll *mongo.Collection) CardRewardRepository {
	return &cardRewardRepository{coll: coll}
}

func earnPoints(amount int64, mcc string) int32 {
	points := int32(amount / 10000)
	switch mcc {
	case "5812", "4511", "5541":
		points *= 2
	}
	if points < 1 {
		points = 1
	}
	return points
}

func (r *cardRewardRepository) EarnRewards(ctx context.Context, req *requests.EarnRewardsRequest) (*salemongo.CardReward, error) {
	rewardID, err := salemongo.NextID(ctx, r.coll.Database(), "card_rewards")
	if err != nil {
		return nil, sharedErrors.ErrFailed("earn card rewards").WithInternal(err)
	}

	now := time.Now()
	reward := &salemongo.CardReward{
		RewardID:     rewardID,
		CardNumber:   req.CardNumber,
		TxnID:        req.TxnID,
		Amount:       req.Amount,
		Mcc:          req.Mcc,
		PointsEarned: earnPoints(req.Amount, req.Mcc),
		ExpiresAt:    now.AddDate(1, 0, 0),
		CreatedAt:    &now,
	}
	result, err := r.coll.InsertOne(ctx, reward)
	if err != nil {
		return nil, sharedErrors.ErrFailed("earn card rewards").WithInternal(err)
	}
	reward.ID = result.InsertedID.(primitive.ObjectID)
	return reward, nil
}

func (r *cardRewardRepository) GetBalance(ctx context.Context, cardNumber string) (int64, error) {
	// Aggregate sum of points_earned where not redeemed and not expired
	pipeline := salemongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"card_number": cardNumber,
			"redeemed":    false,
			"expires_at":  bson.M{"$gt": time.Now()},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$points_earned"},
		}}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, sharedErrors.ErrFailed("get reward balance").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var results []struct {
		Total int64 `bson:"total"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return 0, sharedErrors.ErrFailed("get reward balance").WithInternal(err)
	}
	if len(results) == 0 {
		return 0, nil
	}
	return results[0].Total, nil
}

func (r *cardRewardRepository) GetHistory(ctx context.Context, cardNumber string) ([]*salemongo.CardReward, error) {
	cursor, err := r.coll.Find(ctx, bson.M{"card_number": cardNumber},
		defaultSortByCreatedAtDesc())
	if err != nil {
		return nil, sharedErrors.ErrFailed("get reward history").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var rewards []*salemongo.CardReward
	if err := cursor.All(ctx, &rewards); err != nil {
		return nil, sharedErrors.ErrFailed("get reward history").WithInternal(err)
	}
	return rewards, nil
}

func (r *cardRewardRepository) RedeemRewards(ctx context.Context, cardNumber string, points int64) (int64, error) {
	cursor, err := r.coll.Find(ctx, bson.M{
		"card_number":   cardNumber,
		"redeemed":      false,
		"expires_at":    bson.M{"$gt": time.Now()},
		"points_earned": bson.M{"$gt": 0},
	}, defaultSortByCreatedAtAsc())
	if err != nil {
		return 0, sharedErrors.ErrFailed("find redeemable rewards").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var rewards []salemongo.CardReward
	if err := cursor.All(ctx, &rewards); err != nil {
		return 0, sharedErrors.ErrFailed("find redeemable rewards").WithInternal(err)
	}

	var ids []primitive.ObjectID
	var total int64
	for _, row := range rewards {
		if total >= points {
			break
		}
		ids = append(ids, row.ID)
		total += int64(row.PointsEarned)
	}
	if total < points {
		return 0, sharedErrors.ErrBadRequest.WithMessage("Insufficient reward points")
	}

	_, err = r.coll.UpdateMany(ctx, bson.M{"_id": bson.M{"$in": ids}}, bson.M{"$set": bson.M{"redeemed": true}})
	if err != nil {
		return 0, sharedErrors.ErrFailed("redeem rewards").WithInternal(err)
	}

	// Whole ledger entries are consumed, so the actually redeemed total is the
	// sum of the consumed entries and may exceed the requested points. Callers
	// rely on it to derive the remaining balance.
	return total, nil
}

func defaultSortByCreatedAtDesc() *options.FindOptions {
	return options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
}

func defaultSortByCreatedAtAsc() *options.FindOptions {
	return options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})
}
