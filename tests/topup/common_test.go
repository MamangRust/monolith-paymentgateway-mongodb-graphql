package topup_test

import (
	"context"
	"time"

	card_repo "github.com/MamangRust/monolith-payment-gateway-card/repository"
	models "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	saldo_repo "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type realCardRepo struct {
	query card_repo.CardQueryRepository
	cmd   card_repo.CardCommandRepository
}

func (r *realCardRepo) FindUserCardByCardNumber(ctx context.Context, card_number string) (*models.CardByEmailRow, error) {
	card, err := r.query.FindUserCardByCardNumber(ctx, card_number)
	if err != nil || card == nil {
		return nil, err
	}
	return &models.CardByEmailRow{CardNumber: card.CardNumber, Email: "topup@example.com"}, nil
}
func (r *realCardRepo) FindCardByCardNumber(ctx context.Context, card_number string) (*models.CardAllFieldsRow, error) {
	card, err := r.query.FindCardByCardNumber(ctx, card_number)
	if err != nil || card == nil {
		return nil, err
	}
	return &models.CardAllFieldsRow{CardNumber: card.CardNumber, CardType: card.CardType}, nil
}
func (r *realCardRepo) UpdateCard(ctx context.Context, request *requests.UpdateCardRequest) (*models.CardUpdateRow, error) {
	return &models.CardUpdateRow{CardType: request.CardType}, nil
}

type realSaldoRepo struct {
	repo saldo_repo.Repositories
}

func (r *realSaldoRepo) FindByCardNumber(ctx context.Context, card_number string) (*models.Saldo, error) {
	return r.repo.FindByCardNumber(ctx, card_number)
}
func (r *realSaldoRepo) UpdateSaldoBalance(ctx context.Context, request *requests.UpdateSaldoBalance) (*models.UpdateSaldoBalanceRow, error) {
	result, err := r.repo.UpdateSaldoBalance(ctx, request)
	if err != nil || result == nil {
		return nil, err
	}
	return &models.UpdateSaldoBalanceRow{
		SaldoID:      result.SaldoID,
		CardNumber:   result.CardNumber,
		TotalBalance: int32(result.TotalBalance),
	}, nil
}

type dummyCacheMetrics struct{}

func (d *dummyCacheMetrics) RecordCacheHit(ctx context.Context, key string)                  {}
func (d *dummyCacheMetrics) RecordCacheMiss(ctx context.Context, key string)                 {}
func (d *dummyCacheMetrics) RecordCacheSet(ctx context.Context, key string, success bool)    {}
func (d *dummyCacheMetrics) RecordCacheDelete(ctx context.Context, key string, success bool) {}
func (d *dummyCacheMetrics) RecordCacheOperationLatency(ctx context.Context, operation string, duration time.Duration) {
}
func (d *dummyCacheMetrics) RecordCacheError(ctx context.Context, operation, key string, err error) {}
