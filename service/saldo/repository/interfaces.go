package repository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

// SaldoResult is the response type for saldo list queries (replaces sqlc row types).
type SaldoResult struct {
	SaldoID        int32
	CardNumber     string
	TotalBalance   int64
	WithdrawAmount *int64
	WithdrawTime   *time.Time
	CreatedAt      *time.Time
	UpdatedAt      *time.Time
	DeletedAt      *time.Time
	TotalCount     int64
}

// ReconciliationQueueResult is the response type for reconciliation queue queries.
type ReconciliationQueueResult struct {
	QueueID               int64
	SaldoID               int32
	CardNumber            string
	CurrentBalance        int64
	LedgerBalance         int64
	Difference            int64
	LedgerEntries         int64
	Status                string
	MismatchCount         int64
	FirstSeenAt           time.Time
	LastSeenAt            time.Time
	ResolvedAt            *time.Time
	ResolutionOperationID *string
	ResolutionNote        *string
}

// LedgerEntryResult is the response type for ledger entry queries.
type LedgerEntryResult struct {
	EntryID       int64
	OperationID   string
	CardNumber    string
	Direction     string
	Amount        int64
	Delta         int64
	BalanceBefore int64
	BalanceAfter  int64
	SourceType    string
	SourceID      *string
	CreatedAt     time.Time
}

// SaldoMutationResult is the response type for debit/credit/adjustment results.
type SaldoMutationResult struct {
	SaldoID        int32
	CardNumber     string
	TotalBalance   int64
	WithdrawAmount *int32
	WithdrawTime   *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// SaldoAdjustmentResult is the response type for saldo adjustment.
type SaldoAdjustmentResult struct {
	SaldoID      int32
	CardNumber   string
	TotalBalance int64
}

type SaldoQueryRepository interface {
	ListReconciliationQueue(ctx context.Context, status string, limit int32) ([]*ReconciliationQueueResult, error)
	ListLedgerEntries(ctx context.Context, cardNumber string, limit int32) ([]*LedgerEntryResult, error)
	FindAllSaldos(ctx context.Context, req *requests.FindAllSaldos) ([]*SaldoResult, error)
	FindByActive(ctx context.Context, req *requests.FindAllSaldos) ([]*SaldoResult, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllSaldos) ([]*SaldoResult, error)
	FindById(ctx context.Context, saldo_id int) (*SaldoResult, error)
	FindByCardNumber(ctx context.Context, card_number string) (*salemongo.Saldo, error)
	CountAllSaldos(ctx context.Context, search string) (int64, error)
	CountActiveSaldos(ctx context.Context, search string) (int64, error)
	CountTrashedSaldos(ctx context.Context, search string) (int64, error)
}

type SaldoCommandRepository interface {
	CreateSaldo(ctx context.Context, request *requests.CreateSaldoRequest) (*SaldoMutationResult, error)
	CreateSaldoIfNotExists(ctx context.Context, request *requests.CreateSaldoRequest) error
	UpdateSaldo(ctx context.Context, request *requests.UpdateSaldoRequest) (*SaldoMutationResult, error)
	UpdateSaldoBalance(ctx context.Context, request *requests.UpdateSaldoBalance) (*SaldoMutationResult, error)
	DebitSaldo(ctx context.Context, request *requests.DebitSaldoRequest) (*SaldoMutationResult, error)
	CreditSaldo(ctx context.Context, request *requests.CreditSaldoRequest) (*SaldoMutationResult, error)
	UpdateSaldoWithdraw(ctx context.Context, request *requests.UpdateSaldoWithdraw) (*SaldoMutationResult, error)
	ApplySaldoAdjustment(ctx context.Context, request *requests.ApplySaldoAdjustmentRequest) (*SaldoAdjustmentResult, error)
	ResolveReconciliation(ctx context.Context, queueID int64, operationID, note string) error
	TrashedSaldo(ctx context.Context, saldoID int) (*salemongo.Saldo, error)
	RestoreSaldo(ctx context.Context, saldoID int) (*salemongo.Saldo, error)
	DeleteSaldoPermanent(ctx context.Context, saldo_id int) (bool, error)
	RestoreAllSaldo(ctx context.Context) (bool, error)
	DeleteAllSaldoPermanent(ctx context.Context) (bool, error)
}

type CardRepository interface {
	FindCardByCardNumber(ctx context.Context, card_number string) (*salemongo.Card, error)
}
