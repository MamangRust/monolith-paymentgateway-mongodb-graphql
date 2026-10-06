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

type merchantTransactionRepository struct {
	db           *mongo.Database
	txColl       *mongo.Collection
	merchantColl *mongo.Collection
}

func NewMerchantTransactionRepository(txColl *mongo.Collection) MerchantTransactionRepository {
	return &merchantTransactionRepository{txColl: txColl, merchantColl: txColl.Database().Collection("merchants")}
}

func txSearchFilter(search string, base bson.M) bson.M {
	filter := bson.M{}
	if base != nil {
		for k, v := range base {
			filter[k] = v
		}
	}
	if search != "" {
		filter["card_number"] = bson.M{"$regex": regexp.QuoteMeta(search), "$options": "i"}
	}
	return filter
}

func (r *merchantTransactionRepository) findPage(ctx context.Context, filter bson.M, search string, page, pageSize int) ([]*salemongo.Transaction, error) {
	if search != "" {
		filter["card_number"] = bson.M{"$regex": regexp.QuoteMeta(search), "$options": "i"}
	}
	opts := options.Find().
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize)).
		SetSort(bson.D{{Key: "transaction_id", Value: 1}})

	cursor, err := r.txColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, sharedErrors.ErrFailed("find merchant transactions").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var transactions []*salemongo.Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		return nil, sharedErrors.ErrFailed("decode merchant transactions").WithInternal(err)
	}
	return transactions, nil
}

func (r *merchantTransactionRepository) FindAllTransactions(ctx context.Context, req *requests.FindAllMerchantTransactions) ([]*salemongo.Transaction, error) {
	return r.findPage(ctx, bson.M{}, req.Search, req.Page, req.PageSize)
}

func (r *merchantTransactionRepository) FindAllTransactionsByMerchant(ctx context.Context, req *requests.FindAllMerchantTransactionsById) ([]*salemongo.Transaction, error) {
	return r.findPage(ctx, bson.M{"merchant_id": req.MerchantID}, req.Search, req.Page, req.PageSize)
}

func (r *merchantTransactionRepository) FindAllTransactionsByApikey(ctx context.Context, req *requests.FindAllMerchantTransactionsByApiKey) ([]*salemongo.Transaction, error) {
	var merchant salemongo.Merchant
	err := r.merchantColl.FindOne(ctx, bson.M{"api_key": req.ApiKey}).Decode(&merchant)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return []*salemongo.Transaction{}, nil
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return r.findPage(ctx, bson.M{"merchant_id": merchant.MerchantID}, req.Search, req.Page, req.PageSize)
}

func (r *merchantTransactionRepository) count(ctx context.Context, filter bson.M, search string) (int64, error) {
	if search != "" {
		filter["card_number"] = bson.M{"$regex": regexp.QuoteMeta(search), "$options": "i"}
	}
	count, err := r.txColl.CountDocuments(ctx, filter)
	if err != nil {
		return 0, sharedErrors.ErrFailed("count merchant transactions").WithInternal(err)
	}
	return count, nil
}

func (r *merchantTransactionRepository) CountAllTransactions(ctx context.Context, search string) (int64, error) {
	return r.count(ctx, bson.M{}, search)
}

func (r *merchantTransactionRepository) CountTransactionsByMerchant(ctx context.Context, merchantID int, search string) (int64, error) {
	return r.count(ctx, bson.M{"merchant_id": merchantID}, search)
}

func (r *merchantTransactionRepository) CountTransactionsByApikey(ctx context.Context, apiKey, search string) (int64, error) {
	var merchant salemongo.Merchant
	err := r.merchantColl.FindOne(ctx, bson.M{"api_key": apiKey}).Decode(&merchant)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return 0, nil
		}
		return 0, sharedErrors.ErrInternal.WithInternal(err)
	}

	return r.count(ctx, bson.M{"merchant_id": merchant.MerchantID}, search)
}
