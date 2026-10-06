package repository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type saldoQueryRepository struct {
	coll *mongo.Collection
}

func NewSaldoQueryRepository(coll *mongo.Collection) SaldoQueryRepository {
	return &saldoQueryRepository{coll: coll}
}

func (r *saldoQueryRepository) ListReconciliationQueue(ctx context.Context, status string, limit int32) ([]*ReconciliationQueueResult, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	rqColl := r.coll.Database().Collection("reconciliation_queue")

	cursor, err := rqColl.Find(ctx, bson.M{"status": status},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, sharedErrors.ErrFailed("list reconciliation queue").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var queueItems []salemongo.ReconciliationQueue
	if err := cursor.All(ctx, &queueItems); err != nil {
		return nil, sharedErrors.ErrFailed("list reconciliation queue").WithInternal(err)
	}

	results := make([]*ReconciliationQueueResult, 0, len(queueItems))
	for _, rq := range queueItems {
		res := &ReconciliationQueueResult{
			QueueID:               rq.QueueID,
			SaldoID:               rq.SaldoID,
			CardNumber:            rq.CardNumber,
			CurrentBalance:        rq.CurrentBalance,
			LedgerBalance:         rq.LedgerBalance,
			Difference:            rq.Difference,
			LedgerEntries:         rq.LedgerEntries,
			Status:                rq.Status,
			MismatchCount:         rq.MismatchCount,
			FirstSeenAt:           derefTime(rq.FirstSeenAt),
			LastSeenAt:            derefTime(rq.LastSeenAt),
			ResolvedAt:            rq.ResolvedAt,
			ResolutionOperationID: rq.ResolutionOperationID,
			ResolutionNote:        rq.ResolutionNote,
		}
		results = append(results, res)
	}
	return results, nil
}

func (r *saldoQueryRepository) ListLedgerEntries(ctx context.Context, cardNumber string, limit int32) ([]*LedgerEntryResult, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	ledgerColl := r.coll.Database().Collection("balance_ledger")

	cursor, err := ledgerColl.Find(ctx, bson.M{"card_number": cardNumber},
		options.Find().SetSort(bson.D{{Key: "entry_id", Value: -1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, sharedErrors.ErrFailed("list ledger entries").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var entries []salemongo.BalanceLedger
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, sharedErrors.ErrFailed("list ledger entries").WithInternal(err)
	}

	results := make([]*LedgerEntryResult, 0, len(entries))
	for _, e := range entries {
		var createdAt time.Time
		if e.CreatedAt != nil {
			createdAt = *e.CreatedAt
		}
		results = append(results, &LedgerEntryResult{
			EntryID:       e.EntryID,
			OperationID:   e.OperationID,
			CardNumber:    e.CardNumber,
			Direction:     e.Direction,
			Amount:        e.Amount,
			Delta:         e.Delta,
			BalanceBefore: e.BalanceBefore,
			BalanceAfter:  e.BalanceAfter,
			SourceType:    e.SourceType,
			SourceID:      e.SourceID,
			CreatedAt:     createdAt,
		})
	}
	return results, nil
}

func (r *saldoQueryRepository) FindAllSaldos(ctx context.Context, req *requests.FindAllSaldos) ([]*SaldoResult, error) {
	offset := (req.Page - 1) * req.PageSize

	filter := bson.M{}
	if req.Search != "" {
		filter["card_number"] = bson.M{"$regex": req.Search, "$options": "i"}
	}

	cursor, err := r.coll.Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "saldo_id", Value: 1}}).
			SetSkip(int64(offset)).SetLimit(int64(req.PageSize)))
	if err != nil {
		return nil, sharedErrors.ErrFailed("find all saldo records").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var saldos []salemongo.Saldo
	if err := cursor.All(ctx, &saldos); err != nil {
		return nil, sharedErrors.ErrFailed("find all saldo records").WithInternal(err)
	}

	results := make([]*SaldoResult, 0, len(saldos))
	for _, s := range saldos {
		results = append(results, saldoToResult(&s))
	}
	return results, nil
}

func (r *saldoQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllSaldos) ([]*SaldoResult, error) {
	offset := (req.Page - 1) * req.PageSize

	filter := bson.M{"deleted_at": nil}
	if req.Search != "" {
		filter["card_number"] = bson.M{"$regex": req.Search, "$options": "i"}
	}

	cursor, err := r.coll.Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "saldo_id", Value: 1}}).
			SetSkip(int64(offset)).SetLimit(int64(req.PageSize)))
	if err != nil {
		return nil, sharedErrors.ErrFailed("find active saldo records").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var saldos []salemongo.Saldo
	if err := cursor.All(ctx, &saldos); err != nil {
		return nil, sharedErrors.ErrFailed("find active saldo records").WithInternal(err)
	}

	results := make([]*SaldoResult, 0, len(saldos))
	for _, s := range saldos {
		results = append(results, saldoToResult(&s))
	}
	return results, nil
}

func (r *saldoQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllSaldos) ([]*SaldoResult, error) {
	offset := (req.Page - 1) * req.PageSize

	filter := bson.M{"deleted_at": bson.M{"$ne": nil}}
	if req.Search != "" {
		filter["card_number"] = bson.M{"$regex": req.Search, "$options": "i"}
	}

	cursor, err := r.coll.Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "saldo_id", Value: 1}}).
			SetSkip(int64(offset)).SetLimit(int64(req.PageSize)))
	if err != nil {
		return nil, sharedErrors.ErrFailed("find trashed saldo records").WithInternal(err)
	}
	defer cursor.Close(ctx)

	var saldos []salemongo.Saldo
	if err := cursor.All(ctx, &saldos); err != nil {
		return nil, sharedErrors.ErrFailed("find trashed saldo records").WithInternal(err)
	}

	results := make([]*SaldoResult, 0, len(saldos))
	for _, s := range saldos {
		results = append(results, saldoToResult(&s))
	}
	return results, nil
}

func (r *saldoQueryRepository) FindById(ctx context.Context, saldoID int) (*SaldoResult, error) {
	var saldo salemongo.Saldo
	err := r.coll.FindOne(ctx, bson.M{"saldo_id": saldoID}).Decode(&saldo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, sharedErrors.ErrNotFoundResponse("saldo record").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return saldoToResult(&saldo), nil
}

func (r *saldoQueryRepository) FindByCardNumber(ctx context.Context, cardNumber string) (*salemongo.Saldo, error) {
	var saldo salemongo.Saldo
	err := r.coll.FindOne(ctx, bson.M{"card_number": cardNumber}).Decode(&saldo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, sharedErrors.ErrNotFoundResponse("saldo record").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &saldo, nil
}

func (r *saldoQueryRepository) CountAllSaldos(ctx context.Context, search string) (int64, error) {
	return r.CountActiveSaldos(ctx, search)
}

func (r *saldoQueryRepository) CountActiveSaldos(ctx context.Context, search string) (int64, error) {
	filter := bson.M{"deleted_at": nil}
	if search != "" {
		filter["card_number"] = bson.M{"$regex": search, "$options": "i"}
	}
	count, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *saldoQueryRepository) CountTrashedSaldos(ctx context.Context, search string) (int64, error) {
	filter := bson.M{"deleted_at": bson.M{"$ne": nil}}
	if search != "" {
		filter["card_number"] = bson.M{"$regex": search, "$options": "i"}
	}
	count, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func saldoToResult(s *salemongo.Saldo) *SaldoResult {
	return &SaldoResult{
		SaldoID:        s.SaldoID,
		CardNumber:     s.CardNumber,
		TotalBalance:   s.TotalBalance,
		WithdrawAmount: s.WithdrawAmount,
		WithdrawTime:   s.WithdrawTime,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
		DeletedAt:      s.DeletedAt,
	}
}
