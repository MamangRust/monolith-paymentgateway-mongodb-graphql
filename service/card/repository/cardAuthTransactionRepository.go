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

type cardAuthTransactionRepository struct {
	coll *mongo.Collection
}

func NewCardAuthTransactionRepository(coll *mongo.Collection) CardAuthTransactionRepository {
	return &cardAuthTransactionRepository{coll: coll}
}

func (r *cardAuthTransactionRepository) InsertPending(ctx context.Context, req *requests.AuthorizeCardRequest) (*salemongo.CardAuthTransaction, error) {
	authID, err := salemongo.NextID(ctx, r.coll.Database(), "card_auth_transactions")
	if err != nil {
		return nil, sharedErrors.ErrFailed("insert authorization transaction").WithInternal(err)
	}

	now := time.Now()
	txn := &salemongo.CardAuthTransaction{
		AuthID:         authID,
		TxnID:          req.IdempotencyKey + "-txn",
		CardNumber:     req.CardNumber,
		MerchantID:     int32(req.MerchantID),
		Amount:         req.Amount,
		Currency:       req.Currency,
		Mcc:            req.Mcc,
		PosEntryMode:   req.PosEntryMode,
		IdempotencyKey: req.IdempotencyKey,
		Status:         "pending",
		CreatedAt:      &now,
		UpdatedAt:      &now,
	}
	result, err := r.coll.InsertOne(ctx, txn)
	if err != nil {
		return nil, sharedErrors.ErrFailed("insert authorization transaction").WithInternal(err)
	}
	txn.ID = result.InsertedID.(primitive.ObjectID)
	return txn, nil
}

func (r *cardAuthTransactionRepository) setStatus(ctx context.Context, txnID, status string) (*salemongo.CardAuthTransaction, error) {
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"txn_id": txnID},
		bson.M{"$set": bson.M{"status": status, "updated_at": time.Now()}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var txn salemongo.CardAuthTransaction
	if err := result.Decode(&txn); err != nil {
		return nil, sharedErrors.ErrFailed("update authorization transaction").WithInternal(err)
	}
	return &txn, nil
}

func (r *cardAuthTransactionRepository) Approve(ctx context.Context, txnID string) (*salemongo.CardAuthTransaction, error) {
	return r.setStatus(ctx, txnID, "approved")
}

func (r *cardAuthTransactionRepository) Decline(ctx context.Context, txnID string) (*salemongo.CardAuthTransaction, error) {
	return r.setStatus(ctx, txnID, "declined")
}

func (r *cardAuthTransactionRepository) Reverse(ctx context.Context, txnID string) (*salemongo.CardAuthTransaction, error) {
	return r.setStatus(ctx, txnID, "reversed")
}

func (r *cardAuthTransactionRepository) FindByIdempotencyKey(ctx context.Context, key string) (*salemongo.CardAuthTransaction, error) {
	var txn salemongo.CardAuthTransaction
	err := r.coll.FindOne(ctx, bson.M{"idempotency_key": key}).Decode(&txn)
	if err != nil {
		return nil, sharedErrors.ErrFailed("get authorization transaction").WithInternal(err)
	}
	return &txn, nil
}

func (r *cardAuthTransactionRepository) FindByTxnID(ctx context.Context, txnID string) (*salemongo.CardAuthTransaction, error) {
	var txn salemongo.CardAuthTransaction
	err := r.coll.FindOne(ctx, bson.M{"txn_id": txnID}).Decode(&txn)
	if err != nil {
		return nil, sharedErrors.ErrFailed("get authorization transaction").WithInternal(err)
	}
	return &txn, nil
}

func (r *cardAuthTransactionRepository) FindByCardNumber(ctx context.Context, cardNumber string, page, pageSize int) ([]*salemongo.GetAuthTxnByCardNumberRow, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	cursor, err := r.coll.Find(ctx, bson.M{"card_number": cardNumber},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(int64(offset)).SetLimit(int64(pageSize)))
	if err != nil {
		return nil, sharedErrors.ErrFailed("get authorization transactions").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var txns []*salemongo.CardAuthTransaction
	if err := cursor.All(ctx, &txns); err != nil {
		return nil, sharedErrors.ErrFailed("get authorization transactions").WithInternal(err)
	}

	rows := make([]*salemongo.GetAuthTxnByCardNumberRow, len(txns))
	for i, t := range txns {
		createdAt := t.CreatedAt
		if createdAt == nil {
			now := time.Now()
			createdAt = &now
		}
		updatedAt := t.UpdatedAt
		if updatedAt == nil {
			now := time.Now()
			updatedAt = &now
		}
		rows[i] = &salemongo.GetAuthTxnByCardNumberRow{
			TxnID:          t.TxnID,
			CardNumber:     t.CardNumber,
			MerchantID:     t.MerchantID,
			Amount:         t.Amount,
			Currency:       t.Currency,
			Mcc:            t.Mcc,
			PosEntryMode:   t.PosEntryMode,
			Status:         t.Status,
			IdempotencyKey: t.IdempotencyKey,
			RiskScore:      t.RiskScore,
			CreatedAt:      *createdAt,
			UpdatedAt:      *updatedAt,
		}
	}
	return rows, nil
}

func (r *cardAuthTransactionRepository) CountRecentByCardNumber(ctx context.Context, cardNumber string, since time.Time) (int, error) {
	count, err := r.coll.CountDocuments(ctx, bson.M{
		"card_number": cardNumber,
		"created_at":  bson.M{"$gte": since},
	})
	if err != nil {
		return 0, sharedErrors.ErrFailed("count recent authorization transactions").WithInternal(err)
	}
	return int(count), nil
}

func (r *cardAuthTransactionRepository) UpdateRiskScore(ctx context.Context, txnID string, score int) error {
	_, err := r.coll.UpdateOne(
		ctx,
		bson.M{"txn_id": txnID},
		bson.M{"$set": bson.M{"risk_score": int32(score), "updated_at": time.Now()}},
	)
	if err != nil {
		return sharedErrors.ErrFailed("update authorization risk score").WithInternal(err)
	}
	return nil
}
