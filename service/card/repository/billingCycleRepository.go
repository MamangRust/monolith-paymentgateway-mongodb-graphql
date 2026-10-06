package repository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	statemachine "github.com/MamangRust/monolith-payment-gateway-shared/domain/status"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type billingCycleRepository struct {
	coll *mongo.Collection
}

func NewBillingCycleRepository(coll *mongo.Collection) BillingCycleRepository {
	return &billingCycleRepository{coll: coll}
}

func (r *billingCycleRepository) GetBillingCycleByID(ctx context.Context, billingID int) (*salemongo.BillingCycle, error) {
	var cycle salemongo.BillingCycle
	err := r.coll.FindOne(ctx, bson.M{"billing_id": billingID}).Decode(&cycle)
	if err != nil {
		return nil, sharedErrors.ErrFailed("get billing cycle by id").WithInternal(err)
	}
	return &cycle, nil
}

func (r *billingCycleRepository) GetBillingCyclesByCardNumber(ctx context.Context, cardNumber string) ([]*salemongo.BillingCycle, error) {
	cursor, err := r.coll.Find(ctx, bson.M{"card_number": cardNumber},
		options.Find().SetSort(bson.D{{Key: "cycle_start", Value: -1}}))
	if err != nil {
		return nil, sharedErrors.ErrFailed("get billing cycles").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var cycles []*salemongo.BillingCycle
	if err := cursor.All(ctx, &cycles); err != nil {
		return nil, sharedErrors.ErrFailed("get billing cycles").WithInternal(err)
	}
	return cycles, nil
}

func (r *billingCycleRepository) CreateBillingCycles(ctx context.Context, cycleStart, cycleEnd, dueDate time.Time) (int, error) {
	// Find all active cards
	cardColl := r.coll.Database().Collection("cards")
	cursor, err := cardColl.Find(ctx, bson.M{"deleted_at": nil})
	if err != nil {
		return 0, sharedErrors.ErrFailed("find active cards for billing cycles").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var activeCards []salemongo.Card
	if err := cursor.All(ctx, &activeCards); err != nil {
		return 0, sharedErrors.ErrFailed("find active cards for billing cycles").WithInternal(err)
	}

	var created int
	now := time.Now()
	for _, card := range activeCards {
		billingID, err := salemongo.NextID(ctx, r.coll.Database(), "billing_cycles")
		if err != nil {
			return created, sharedErrors.ErrFailed("create billing cycle").WithInternal(err)
		}
		cycle := &salemongo.BillingCycle{
			BillingID:  billingID,
			CardNumber: card.CardNumber,
			CycleStart: cycleStart,
			CycleEnd:   cycleEnd,
			DueDate:    dueDate,
			Status:     statemachine.Unpaid,
			CreatedAt:  &now,
			UpdatedAt:  &now,
		}
		result, err := r.coll.InsertOne(ctx, cycle)
		if err != nil {
			if salemongo.IsDuplicateKeyError(err) {
				continue
			}
			return created, sharedErrors.ErrFailed("create billing cycle").WithInternal(err)
		}
		cycle.ID = result.InsertedID.(primitive.ObjectID)
		created++
	}
	return created, nil
}
