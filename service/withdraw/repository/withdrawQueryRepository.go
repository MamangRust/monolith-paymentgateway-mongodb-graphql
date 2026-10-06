package repository

import (
	"context"
	"errors"
	"regexp"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type withdrawQueryRepository struct {
	coll *mongo.Collection
}

func NewWithdrawQueryRepository(coll *mongo.Collection) WithdrawQueryRepository {
	return &withdrawQueryRepository{coll: coll}
}

func withdrawSearchFilter(search string) bson.M {
	rx := bson.M{"$regex": regexp.QuoteMeta(search), "$options": "i"}
	return bson.M{"$or": bson.A{
		bson.M{"card_number": rx},
		bson.M{"status": rx},
	}}
}

func (r *withdrawQueryRepository) findPage(ctx context.Context, filter bson.M, page, pageSize int) ([]*salemongo.Withdraw, int64, error) {
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, sharedErrors.ErrFailed("count withdraws").WithInternal(err)
	}

	opts := options.Find().
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize)).
		SetSort(bson.D{{Key: "withdraw_id", Value: 1}})

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, sharedErrors.ErrFailed("find withdraws").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var withdraws []*salemongo.Withdraw
	if err := cursor.All(ctx, &withdraws); err != nil {
		return nil, 0, sharedErrors.ErrFailed("decode withdraws").WithInternal(err)
	}
	return withdraws, total, nil
}

func (r *withdrawQueryRepository) FindAll(ctx context.Context, req *requests.FindAllWithdraws) ([]*WithdrawQueryResult, error) {
	filter := bson.M{}
	if req.Search != "" {
		filter = withdrawSearchFilter(req.Search)
	}
	withdraws, total, err := r.findPage(ctx, filter, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	results := make([]*WithdrawQueryResult, len(withdraws))
	for i, w := range withdraws {
		results[i] = &WithdrawQueryResult{Withdraw: w, TotalCount: total}
	}
	return results, nil
}

func (r *withdrawQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllWithdraws) ([]*WithdrawQueryResult, error) {
	filter := bson.M{"deleted_at": nil}
	if req.Search != "" {
		filter = withdrawSearchFilter(req.Search)
		filter["deleted_at"] = nil
	}
	withdraws, total, err := r.findPage(ctx, filter, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	results := make([]*WithdrawQueryResult, len(withdraws))
	for i, w := range withdraws {
		results[i] = &WithdrawQueryResult{Withdraw: w, TotalCount: total}
	}
	return results, nil
}

func (r *withdrawQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllWithdraws) ([]*WithdrawQueryResult, error) {
	filter := bson.M{"deleted_at": bson.M{"$ne": nil}}
	if req.Search != "" {
		filter = withdrawSearchFilter(req.Search)
		filter["deleted_at"] = bson.M{"$ne": nil}
	}
	withdraws, total, err := r.findPage(ctx, filter, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	results := make([]*WithdrawQueryResult, len(withdraws))
	for i, w := range withdraws {
		results[i] = &WithdrawQueryResult{Withdraw: w, TotalCount: total}
	}
	return results, nil
}

func (r *withdrawQueryRepository) FindAllByCardNumber(ctx context.Context, req *requests.FindAllWithdrawCardNumber) ([]*WithdrawQueryResult, error) {
	filter := bson.M{"card_number": req.CardNumber}
	if req.Search != "" {
		filter["status"] = bson.M{"$regex": regexp.QuoteMeta(req.Search), "$options": "i"}
	}
	withdraws, total, err := r.findPage(ctx, filter, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	results := make([]*WithdrawQueryResult, len(withdraws))
	for i, w := range withdraws {
		results[i] = &WithdrawQueryResult{Withdraw: w, TotalCount: total}
	}
	return results, nil
}

func (r *withdrawQueryRepository) FindById(ctx context.Context, id int) (*salemongo.Withdraw, error) {
	var withdraw salemongo.Withdraw
	err := r.coll.FindOne(ctx, bson.M{"withdraw_id": id}).Decode(&withdraw)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("withdraw").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &withdraw, nil
}

// GetTodayWithdrawSumByCardNumber sums successful withdrawals for the current
// calendar day in Asia/Jakarta time. MongoDB stores withdraw_time as UTC, so
// the daily window is computed as [start-of-day, start-of-next-day) in the
// Jakarta timezone offset (+07:00).
func (r *withdrawQueryRepository) GetTodayWithdrawSumByCardNumber(ctx context.Context, cardNumber string) (int64, error) {
	jkt := time.FixedZone("Asia/Jakarta", 7*60*60)
	now := time.Now().In(jkt)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, jkt)
	endOfDay := startOfDay.AddDate(0, 0, 1)

	pipeline := salemongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"card_number":   cardNumber,
			"status":        "success",
			"withdraw_time": bson.M{"$gte": startOfDay, "$lt": endOfDay},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$withdraw_amount"},
		}}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, sharedErrors.ErrFailed("get today withdraw sum").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var row struct {
		Total int64 `bson:"total"`
	}
	if cursor.Next(ctx) {
		if err := cursor.Decode(&row); err != nil {
			return 0, sharedErrors.ErrFailed("decode withdraw sum").WithInternal(err)
		}
	}
	return row.Total, nil
}
