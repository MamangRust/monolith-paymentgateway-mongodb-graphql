package repository

import (
	"context"
	"fmt"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type saldoCommandRepository struct {
	coll *mongo.Collection
}

func NewSaldoCommandRepository(coll *mongo.Collection) SaldoCommandRepository {
	return &saldoCommandRepository{coll: coll}
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// insertLedgerEntry assigns a monotonic entry ID and appends the balance ledger
// entry. The ledger collection enforces a unique index on entry_id, so the ID
// must be allocated before every insert.
func (r *saldoCommandRepository) insertLedgerEntry(ctx context.Context, ledger *salemongo.BalanceLedger) error {
	entryID, err := salemongo.NextID(ctx, r.coll.Database(), "balance_ledger")
	if err != nil {
		return err
	}

	ledger.EntryID = int64(entryID)

	_, err = r.coll.Database().Collection("balance_ledger").InsertOne(ctx, ledger)
	return err
} // CreateSaldo upserts the saldo for a card: an existing record keeps its ID but
// adopts the requested balance, mirroring the ON CONFLICT DO UPDATE semantics of
// the original schema.
func (r *saldoCommandRepository) CreateSaldo(ctx context.Context, request *requests.CreateSaldoRequest) (*SaldoMutationResult, error) {
	balance := amountToInt64(request.TotalBalance)

	saldoID, err := salemongo.NextID(ctx, r.coll.Database(), "saldos")
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"card_number": request.CardNumber},
		bson.M{
			"$set": bson.M{"total_balance": balance, "updated_at": now},
			"$setOnInsert": bson.M{
				"saldo_id":   saldoID,
				"created_at": now,
			},
		},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	)

	var saldo salemongo.Saldo
	if err := result.Decode(&saldo); err != nil {
		return nil, sharedErrors.ErrFailed("create saldo record").WithInternal(err)
	}

	return &SaldoMutationResult{
		SaldoID: saldo.SaldoID, CardNumber: saldo.CardNumber, TotalBalance: saldo.TotalBalance,
		WithdrawAmount: int32Ptr(saldo.WithdrawAmount), WithdrawTime: saldo.WithdrawTime,
		CreatedAt: derefTime(saldo.CreatedAt), UpdatedAt: derefTime(saldo.UpdatedAt),
	}, nil
}

func (r *saldoCommandRepository) CreateSaldoIfNotExists(ctx context.Context, request *requests.CreateSaldoRequest) error {
	balance := amountToInt64(request.TotalBalance)

	saldoID, err := salemongo.NextID(ctx, r.coll.Database(), "saldos")
	if err != nil {
		return sharedErrors.ErrFailed("create saldo record").WithInternal(err)
	}

	now := time.Now()
	_, err = r.coll.UpdateOne(
		ctx,
		bson.M{"card_number": request.CardNumber},
		bson.M{"$setOnInsert": bson.M{
			"saldo_id":      saldoID,
			"card_number":   request.CardNumber,
			"total_balance": balance,
			"created_at":    now,
			"updated_at":    now,
		}},
		options.Update().SetUpsert(true),
	)
	return err
}

func (r *saldoCommandRepository) UpdateSaldo(ctx context.Context, request *requests.UpdateSaldoRequest) (*SaldoMutationResult, error) {
	if request.SaldoID == nil {
		return nil, sharedErrors.NewBadRequestError("saldo ID is required")
	}

	var saldo salemongo.Saldo
	err := r.coll.FindOne(ctx, bson.M{"saldo_id": *request.SaldoID}).Decode(&saldo)
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "saldo record", "update saldo record")
	}

	if saldo.CardNumber != request.CardNumber {
		return nil, sharedErrors.NewConflictError("cannot change card number of an existing saldo")
	}

	target := amountToInt64(request.TotalBalance)
	if target == saldo.TotalBalance {
		return &SaldoMutationResult{
			SaldoID: saldo.SaldoID, CardNumber: saldo.CardNumber, TotalBalance: saldo.TotalBalance,
			WithdrawAmount: int32Ptr(saldo.WithdrawAmount), WithdrawTime: saldo.WithdrawTime,
			CreatedAt: derefTime(saldo.CreatedAt), UpdatedAt: derefTime(saldo.UpdatedAt),
		}, nil
	}

	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"saldo_id": *request.SaldoID},
		bson.M{"$set": bson.M{"total_balance": target, "updated_at": now}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	if err := result.Decode(&saldo); err != nil {
		return nil, sharedErrors.ErrFailed("update saldo record").WithInternal(err)
	}

	return &SaldoMutationResult{
		SaldoID: saldo.SaldoID, CardNumber: saldo.CardNumber, TotalBalance: saldo.TotalBalance,
		WithdrawAmount: int32Ptr(saldo.WithdrawAmount), WithdrawTime: saldo.WithdrawTime,
		CreatedAt: derefTime(saldo.CreatedAt), UpdatedAt: derefTime(saldo.UpdatedAt),
	}, nil
}

func (r *saldoCommandRepository) UpdateSaldoBalance(ctx context.Context, request *requests.UpdateSaldoBalance) (*SaldoMutationResult, error) {
	balance := amountToInt64(request.TotalBalance)
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"card_number": request.CardNumber},
		bson.M{"$set": bson.M{"total_balance": balance, "updated_at": now}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var saldo salemongo.Saldo
	if err := result.Decode(&saldo); err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "saldo record", "update saldo balance")
	}
	return &SaldoMutationResult{
		SaldoID: saldo.SaldoID, CardNumber: saldo.CardNumber, TotalBalance: saldo.TotalBalance,
		WithdrawAmount: int32Ptr(saldo.WithdrawAmount), WithdrawTime: saldo.WithdrawTime,
		CreatedAt: derefTime(saldo.CreatedAt), UpdatedAt: derefTime(saldo.UpdatedAt),
	}, nil
}

func (r *saldoCommandRepository) DebitSaldo(ctx context.Context, request *requests.DebitSaldoRequest) (*SaldoMutationResult, error) {
	amount := amountToInt64(request.Amount)

	operationID := request.OperationID
	if operationID == "" {
		operationID = uuid.NewString()
	}
	sourceType := request.SourceType
	if sourceType == "" {
		sourceType = "unknown"
	}

	// Atomic debit using FindOneAndUpdate with balance check in filter
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"card_number": request.CardNumber, "total_balance": bson.M{"$gte": amount}},
		bson.M{
			"$inc": bson.M{"total_balance": -amount},
			"$set": bson.M{"updated_at": time.Now()},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var saldo salemongo.Saldo
	err := result.Decode(&saldo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Either not found or insufficient balance - check which
			var check salemongo.Saldo
			if findErr := r.coll.FindOne(ctx, bson.M{"card_number": request.CardNumber}).Decode(&check); findErr != nil {
				return nil, sharedErrors.NewConflictError("saldo not found")
			}
			return nil, sharedErrors.NewConflictError("insufficient balance")
		}
		return nil, sharedErrors.ErrFailed("debit saldo").WithInternal(err)
	}

	// Determine balance before from the ledger
	balanceAfter := saldo.TotalBalance
	balanceBefore := balanceAfter + amount

	// Record ledger entry
	ledger := salemongo.BalanceLedger{
		OperationID:   operationID,
		CardNumber:    request.CardNumber,
		Direction:     "debit",
		Amount:        amount,
		Delta:         -amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		SourceType:    sourceType,
		SourceID:      strPtr(request.SourceID),
		CreatedAt:     func() *time.Time { t := time.Now(); return &t }(),
	}
	if err := r.insertLedgerEntry(ctx, &ledger); err != nil {
		return nil, sharedErrors.ErrFailed("debit saldo ledger").WithInternal(err)
	}

	return &SaldoMutationResult{
		SaldoID: saldo.SaldoID, CardNumber: saldo.CardNumber, TotalBalance: saldo.TotalBalance,
		WithdrawAmount: int32Ptr(saldo.WithdrawAmount), WithdrawTime: saldo.WithdrawTime,
		CreatedAt: derefTime(saldo.CreatedAt), UpdatedAt: derefTime(saldo.UpdatedAt),
	}, nil
}

func (r *saldoCommandRepository) CreditSaldo(ctx context.Context, request *requests.CreditSaldoRequest) (*SaldoMutationResult, error) {
	amount := amountToInt64(request.Amount)

	operationID := request.OperationID
	if operationID == "" {
		operationID = uuid.NewString()
	}
	sourceType := request.SourceType
	if sourceType == "" {
		sourceType = "unknown"
	}

	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"card_number": request.CardNumber},
		bson.M{
			"$inc": bson.M{"total_balance": amount},
			"$set": bson.M{"updated_at": time.Now()},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var saldo salemongo.Saldo
	err := result.Decode(&saldo)
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "saldo record", "credit saldo")
	}

	balanceAfter := saldo.TotalBalance
	balanceBefore := balanceAfter - amount

	ledger := salemongo.BalanceLedger{
		OperationID:   operationID,
		CardNumber:    request.CardNumber,
		Direction:     "credit",
		Amount:        amount,
		Delta:         amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		SourceType:    sourceType,
		SourceID:      strPtr(request.SourceID),
		CreatedAt:     func() *time.Time { t := time.Now(); return &t }(),
	}
	if err := r.insertLedgerEntry(ctx, &ledger); err != nil {
		return nil, sharedErrors.ErrFailed("credit saldo ledger").WithInternal(err)
	}

	return &SaldoMutationResult{
		SaldoID: saldo.SaldoID, CardNumber: saldo.CardNumber, TotalBalance: saldo.TotalBalance,
		WithdrawAmount: int32Ptr(saldo.WithdrawAmount), WithdrawTime: saldo.WithdrawTime,
		CreatedAt: derefTime(saldo.CreatedAt), UpdatedAt: derefTime(saldo.UpdatedAt),
	}, nil
}

func (r *saldoCommandRepository) ApplySaldoAdjustment(ctx context.Context, request *requests.ApplySaldoAdjustmentRequest) (*SaldoAdjustmentResult, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"card_number": request.CardNumber},
		bson.M{
			"$inc": bson.M{"total_balance": request.Delta},
			"$set": bson.M{"updated_at": time.Now()},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var saldo salemongo.Saldo
	err := result.Decode(&saldo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, sharedErrors.NewConflictError("saldo adjustment rejected")
		}
		return nil, sharedErrors.ErrFailed("apply saldo adjustment").WithInternal(err)
	}

	balanceAfter := saldo.TotalBalance
	balanceBefore := balanceAfter - request.Delta

	direction := "credit"
	amount := request.Delta
	if amount < 0 {
		direction = "debit"
		amount = -amount
	}

	ledger := salemongo.BalanceLedger{
		OperationID:   request.OperationID,
		CardNumber:    request.CardNumber,
		Direction:     direction,
		Amount:        amount,
		Delta:         request.Delta,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		SourceType:    request.SourceType,
		SourceID:      strPtr(request.SourceID),
		CreatedAt:     func() *time.Time { t := time.Now(); return &t }(),
	}
	if err := r.insertLedgerEntry(ctx, &ledger); err != nil {
		return nil, sharedErrors.ErrFailed("apply saldo adjustment ledger").WithInternal(err)
	}

	return &SaldoAdjustmentResult{
		SaldoID: saldo.SaldoID, CardNumber: saldo.CardNumber, TotalBalance: saldo.TotalBalance,
	}, nil
}

func (r *saldoCommandRepository) ResolveReconciliation(ctx context.Context, queueID int64, operationID, note string) error {
	if queueID <= 0 || operationID == "" {
		return sharedErrors.NewBadRequestError("queue ID and operation ID are required")
	}
	now := time.Now()
	result, err := r.coll.Database().Collection("reconciliation_queue").UpdateOne(
		ctx,
		bson.M{"queue_id": queueID},
		bson.M{"$set": bson.M{
			"status":                  "resolved",
			"resolved_at":             now,
			"resolution_operation_id": operationID,
			"resolution_note":         note,
		}},
	)
	if err != nil {
		return sharedErrors.ErrNoRowsOrFailed(err, "reconciliation queue item", "resolve reconciliation")
	}
	if result.MatchedCount == 0 {
		return sharedErrors.ErrNoRowsOrFailed(mongo.ErrNoDocuments, "reconciliation queue item", "resolve reconciliation")
	}
	return nil
}

func (r *saldoCommandRepository) UpdateSaldoWithdraw(ctx context.Context, request *requests.UpdateSaldoWithdraw) (*SaldoMutationResult, error) {
	setFields := bson.M{"updated_at": time.Now()}
	if request.WithdrawAmount != nil && request.WithdrawTime != nil {
		setFields["withdraw_amount"] = int64(*request.WithdrawAmount)
		setFields["withdraw_time"] = *request.WithdrawTime
	}

	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"card_number": request.CardNumber},
		bson.M{"$set": setFields},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var saldo salemongo.Saldo
	if err := result.Decode(&saldo); err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "saldo record", "update saldo withdrawal metadata")
	}
	return &SaldoMutationResult{
		SaldoID: saldo.SaldoID, CardNumber: saldo.CardNumber, TotalBalance: saldo.TotalBalance,
		WithdrawAmount: int32Ptr(saldo.WithdrawAmount), WithdrawTime: saldo.WithdrawTime,
		CreatedAt: derefTime(saldo.CreatedAt), UpdatedAt: derefTime(saldo.UpdatedAt),
	}, nil
}

func (r *saldoCommandRepository) TrashedSaldo(ctx context.Context, saldoID int) (*salemongo.Saldo, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"saldo_id": saldoID, "deleted_at": nil},
		bson.M{"$set": bson.M{"deleted_at": now}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var saldo salemongo.Saldo
	if err := result.Decode(&saldo); err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "saldo record", "trash saldo record")
	}
	return &saldo, nil
}

func (r *saldoCommandRepository) RestoreSaldo(ctx context.Context, saldoID int) (*salemongo.Saldo, error) {
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"saldo_id": saldoID, "deleted_at": bson.M{"$ne": nil}},
		bson.M{"$unset": bson.M{"deleted_at": ""}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var saldo salemongo.Saldo
	if err := result.Decode(&saldo); err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "saldo record", "restore saldo record")
	}
	return &saldo, nil
}

func (r *saldoCommandRepository) DeleteSaldoPermanent(ctx context.Context, saldoID int) (bool, error) {
	result, err := r.coll.DeleteMany(ctx, bson.M{"saldo_id": saldoID})
	if err != nil {
		return false, sharedErrors.ErrNoRowsOrFailed(err, "saldo record", "delete saldo record permanently")
	}
	if result.DeletedCount == 0 {
		return false, sharedErrors.ErrNotFoundResponse("saldo record")
	}
	return true, nil
}

func (r *saldoCommandRepository) RestoreAllSaldo(ctx context.Context) (bool, error) {
	_, err := r.coll.UpdateMany(ctx, bson.M{"deleted_at": bson.M{"$ne": nil}}, bson.M{"$unset": bson.M{"deleted_at": ""}})
	if err != nil {
		return false, sharedErrors.ErrFailed("restore all saldo records").WithInternal(err)
	}
	return true, nil
}

func (r *saldoCommandRepository) DeleteAllSaldoPermanent(ctx context.Context) (bool, error) {
	_, err := r.coll.DeleteMany(ctx, bson.M{"deleted_at": bson.M{"$ne": nil}})
	if err != nil {
		return false, sharedErrors.ErrFailed(fmt.Sprintf("delete all saldo records permanently")).WithInternal(err)
	}
	return true, nil
}
