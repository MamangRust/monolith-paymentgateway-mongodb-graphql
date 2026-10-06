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

type transactionQueryRepository struct {
	coll *mongo.Collection
}

func NewTransactionQueryRepository(coll *mongo.Collection) TransactionQueryRepository {
	return &transactionQueryRepository{coll: coll}
}

func transactionSearchFilter(search string) bson.M {
	rx := bson.M{"$regex": regexp.QuoteMeta(search), "$options": "i"}
	return bson.M{"$or": bson.A{
		bson.M{"card_number": rx},
		bson.M{"payment_method": rx},
		bson.M{"status": rx},
	}}
}

func transactionCardSearchFilter(search string) bson.M {
	rx := bson.M{"$regex": regexp.QuoteMeta(search), "$options": "i"}
	return bson.M{"$or": bson.A{
		bson.M{"payment_method": rx},
		bson.M{"status": rx},
	}}
}

func (r *transactionQueryRepository) findPage(ctx context.Context, filter bson.M, page, pageSize int) ([]*salemongo.Transaction, int64, error) {
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, sharedErrors.ErrFailed("count transactions").WithInternal(err)
	}

	opts := options.Find().
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize)).
		SetSort(bson.D{{Key: "transaction_time", Value: -1}})

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, sharedErrors.ErrFailed("find transactions").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var txns []*salemongo.Transaction
	if err := cursor.All(ctx, &txns); err != nil {
		return nil, 0, sharedErrors.ErrFailed("decode transactions").WithInternal(err)
	}
	return txns, total, nil
}

func (r *transactionQueryRepository) FindAllTransactions(ctx context.Context, req *requests.FindAllTransactions) ([]*TransactionQueryResult, error) {
	filter := bson.M{}
	if req.Search != "" {
		filter = transactionSearchFilter(req.Search)
	}
	txns, total, err := r.findPage(ctx, filter, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	results := make([]*TransactionQueryResult, len(txns))
	for i, t := range txns {
		results[i] = &TransactionQueryResult{Transaction: t, TotalCount: total}
	}
	return results, nil
}

func (r *transactionQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllTransactions) ([]*TransactionQueryResult, error) {
	filter := bson.M{"deleted_at": nil}
	if req.Search != "" {
		filter = transactionSearchFilter(req.Search)
		filter["deleted_at"] = nil
	}
	txns, total, err := r.findPage(ctx, filter, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	results := make([]*TransactionQueryResult, len(txns))
	for i, t := range txns {
		results[i] = &TransactionQueryResult{Transaction: t, TotalCount: total}
	}
	return results, nil
}

func (r *transactionQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllTransactions) ([]*TransactionQueryResult, error) {
	filter := bson.M{"deleted_at": bson.M{"$ne": nil}}
	if req.Search != "" {
		filter = transactionSearchFilter(req.Search)
		filter["deleted_at"] = bson.M{"$ne": nil}
	}
	txns, total, err := r.findPage(ctx, filter, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	results := make([]*TransactionQueryResult, len(txns))
	for i, t := range txns {
		results[i] = &TransactionQueryResult{Transaction: t, TotalCount: total}
	}
	return results, nil
}

func (r *transactionQueryRepository) FindAllTransactionByCardNumber(ctx context.Context, req *requests.FindAllTransactionCardNumber) ([]*TransactionByCardResult, error) {
	filter := bson.M{"card_number": req.CardNumber}
	if req.Search != "" {
		filter = transactionCardSearchFilter(req.Search)
		filter["card_number"] = req.CardNumber
	}
	txns, total, err := r.findPage(ctx, filter, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	results := make([]*TransactionByCardResult, len(txns))
	for i, t := range txns {
		results[i] = &TransactionByCardResult{Transaction: t, TotalCount: total}
	}
	return results, nil
}

func (r *transactionQueryRepository) FindById(ctx context.Context, transaction_id int) (*salemongo.Transaction, error) {
	var txn salemongo.Transaction
	err := r.coll.FindOne(ctx, bson.M{"transaction_id": int32(transaction_id)}).Decode(&txn)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("transaction").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &txn, nil
}

func (r *transactionQueryRepository) FindTransactionByMerchantId(ctx context.Context, merchant_id int) ([]*salemongo.Transaction, error) {
	cursor, err := r.coll.Find(
		ctx,
		bson.M{"merchant_id": int32(merchant_id)},
		options.Find().SetSort(bson.D{{Key: "transaction_time", Value: -1}}),
	)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	defer cursor.Close(ctx)

	var txns []*salemongo.Transaction
	if err := cursor.All(ctx, &txns); err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return txns, nil
}
