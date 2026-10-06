package repository

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/idempotency"
)

// TopupQueryResult is the paginated topup list response.
type TopupQueryResult struct {
	Topup      *salemongo.Topup
	TotalCount int64
}

// TopupByCardResult is the paginated topup list by card number.
type TopupByCardResult struct {
	Topup      *salemongo.Topup
	TotalCount int64
}

type IdempotencyRepository interface {
	idempotency.Store
}

type OutboxRepository interface {
	Insert(ctx context.Context, record OutboxRecord) error
}

type SaldoRepository interface {
	FindByCardNumber(ctx context.Context, card_number string) (*salemongo.Saldo, error)
	UpdateSaldoBalance(ctx context.Context, request *requests.UpdateSaldoBalance) (*salemongo.UpdateSaldoBalanceRow, error)
}

type TopupQueryRepository interface {
	FindAllTopups(ctx context.Context, req *requests.FindAllTopups) ([]*TopupQueryResult, error)
	FindByActive(ctx context.Context, req *requests.FindAllTopups) ([]*TopupQueryResult, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllTopups) ([]*TopupQueryResult, error)
	FindAllTopupByCardNumber(ctx context.Context, req *requests.FindAllTopupsByCardNumber) ([]*TopupByCardResult, error)

	FindById(ctx context.Context, topup_id int) (*salemongo.Topup, error)
}

type TopupCommandRepository interface {
	CreateTopup(ctx context.Context, request *requests.CreateTopupRequest) (*salemongo.Topup, error)
	UpdateTopup(ctx context.Context, request *requests.UpdateTopupRequest) (*salemongo.Topup, error)

	UpdateTopupAmount(ctx context.Context, request *requests.UpdateTopupAmount) (*salemongo.Topup, error)
	CompareAndSwapTopupAmount(ctx context.Context, topupID, expectedAmount, newAmount int) (bool, error)
	UpdateTopupStatus(ctx context.Context, request *requests.UpdateTopupStatus) (*salemongo.Topup, error)

	TrashedTopup(ctx context.Context, topup_id int) (*salemongo.Topup, error)
	RestoreTopup(ctx context.Context, topup_id int) (*salemongo.Topup, error)
	DeleteTopupPermanent(ctx context.Context, topup_id int) (bool, error)

	RestoreAllTopup(ctx context.Context) (bool, error)
	DeleteAllTopupPermanent(ctx context.Context) (bool, error)
}

type CardRepository interface {
	FindUserCardByCardNumber(ctx context.Context, card_number string) (*salemongo.CardByEmailRow, error)
	FindCardByCardNumber(ctx context.Context, card_number string) (*salemongo.CardAllFieldsRow, error)
	UpdateCard(ctx context.Context, request *requests.UpdateCardRequest) (*salemongo.CardUpdateRow, error)
}
