package repository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type cardPaymentRepository struct {
	coll *mongo.Collection
}

func NewCardPaymentRepository(coll *mongo.Collection) CardPaymentRepository {
	return &cardPaymentRepository{coll: coll}
}

func (r *cardPaymentRepository) PostPayment(ctx context.Context, req *requests.PostPaymentRequest) (*salemongo.CardPayment, error) {
	var billingID *int32
	if req.BillingID != nil {
		billingID = new(int32)
		*billingID = int32(*req.BillingID)
	}

	paymentID, err := salemongo.NextID(ctx, r.coll.Database(), "card_payments")
	if err != nil {
		return nil, sharedErrors.ErrFailed("create card payment").WithInternal(err)
	}

	now := timeNowPtr()
	payment := &salemongo.CardPayment{
		PaymentID:      paymentID,
		PaymentUuid:    uuid.New().String(),
		CardNumber:     req.CardNumber,
		BillingID:      billingID,
		Amount:         req.Amount,
		PaymentChannel: req.PaymentChannel,
		ReferenceID:    req.ReferenceID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	result, err := r.coll.InsertOne(ctx, payment)
	if err != nil {
		return nil, sharedErrors.ErrFailed("create card payment").WithInternal(err)
	}
	payment.ID = result.InsertedID.(primitive.ObjectID)
	return payment, nil
}

func (r *cardPaymentRepository) PostPaymentIdempotent(ctx context.Context, req *requests.PostPaymentRequest, expectedBillingStatus, targetBillingStatus string) (*CardPaymentResult, error) {
	var billingID *int32
	if req.BillingID != nil {
		billingID = new(int32)
		*billingID = int32(*req.BillingID)
	}

	paymentID, err := salemongo.NextID(ctx, r.coll.Database(), "card_payments")
	if err != nil {
		return nil, sharedErrors.ErrFailed("create card payment idempotent").WithInternal(err)
	}

	now := timeNowPtr()
	payment := &salemongo.CardPayment{
		PaymentID:      paymentID,
		PaymentUuid:    uuid.New().String(),
		CardNumber:     req.CardNumber,
		BillingID:      billingID,
		Amount:         req.Amount,
		PaymentChannel: req.PaymentChannel,
		ReferenceID:    req.ReferenceID,
		Status:         "paid",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// If billing cycle involved, transition it atomically. Cycles are keyed by
	// their numeric billing id, not the ObjectID, so the predicate must match on
	// billing_id.
	if expectedBillingStatus != "" && billingID != nil {
		filter := bson.M{
			"billing_id": *billingID,
			"status":     expectedBillingStatus,
		}
		update := bson.M{"$set": bson.M{"status": targetBillingStatus, "updated_at": now}}
		result, err := r.coll.Database().Collection("billing_cycles").UpdateOne(ctx, filter, update)
		if err != nil {
			return nil, sharedErrors.ErrFailed("update billing cycle status").WithInternal(err)
		}
		if result.MatchedCount == 0 {
			return nil, sharedErrors.ErrConflict.WithMessage("billing cycle status changed concurrently")
		}
	}

	result, err := r.coll.InsertOne(ctx, payment)
	if err != nil {
		return nil, sharedErrors.ErrFailed("create card payment idempotent").WithInternal(err)
	}
	payment.ID = result.InsertedID.(primitive.ObjectID)
	return &CardPaymentResult{Payment: payment, Replayed: false}, nil
}

func (r *cardPaymentRepository) GetPaymentByReferenceID(ctx context.Context, referenceID string) (*salemongo.CardPayment, error) {
	var payment salemongo.CardPayment
	err := r.coll.FindOne(ctx, bson.M{"reference_id": referenceID}).Decode(&payment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, sharedErrors.ErrFailed("get card payment by reference ID").WithInternal(err)
	}
	return &payment, nil
}

func (r *cardPaymentRepository) GetPaymentHistory(ctx context.Context, cardNumber string, page, pageSize int) ([]*salemongo.CardPaymentListRow, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	totalCount, err := r.coll.CountDocuments(ctx, bson.M{"card_number": cardNumber})
	if err != nil {
		return nil, sharedErrors.ErrFailed("count card payments").WithInternal(err)
	}

	cursor, err := r.coll.Find(ctx, bson.M{"card_number": cardNumber},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(int64(offset)).SetLimit(int64(pageSize)))
	if err != nil {
		return nil, sharedErrors.ErrFailed("get card payments").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var payments []*salemongo.CardPayment
	if err := cursor.All(ctx, &payments); err != nil {
		return nil, sharedErrors.ErrFailed("get card payments").WithInternal(err)
	}

	rows := make([]*salemongo.CardPaymentListRow, len(payments))
	for i, p := range payments {
		createdAt := time.Now()
		if p.CreatedAt != nil {
			createdAt = *p.CreatedAt
		}
		updatedAt := time.Now()
		if p.UpdatedAt != nil {
			updatedAt = *p.UpdatedAt
		}
		rows[i] = &salemongo.CardPaymentListRow{
			PaymentID:      p.PaymentID,
			PaymentUuid:    p.PaymentUuid,
			CardNumber:     p.CardNumber,
			BillingID:      p.BillingID,
			Amount:         p.Amount,
			PaymentChannel: p.PaymentChannel,
			ReferenceID:    p.ReferenceID,
			Status:         p.Status,
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
			TotalCount:     totalCount,
		}
	}
	return rows, nil
}

func (r *cardPaymentRepository) CountPayments(ctx context.Context, cardNumber string) (int, error) {
	count, err := r.coll.CountDocuments(ctx, bson.M{"card_number": cardNumber})
	if err != nil {
		return 0, sharedErrors.ErrFailed("count card payments").WithInternal(err)
	}
	return int(count), nil
}

func timeNowPtr() *time.Time {
	t := time.Now()
	return &t
}
