package withdraw_test

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
}

func (r *realCardRepo) FindUserCardByCardNumber(ctx context.Context, card_number string) (*models.CardByEmailRow, error) {
	card, err := r.query.FindUserCardByCardNumber(ctx, card_number)
	if err != nil || card == nil {
		return nil, err
	}
	return &models.CardByEmailRow{
		CardID:       card.CardID,
		UserID:       card.UserID,
		CardNumber:   card.CardNumber,
		CardType:     card.CardType,
		ExpireDate:   card.ExpireDate,
		Cvv:          card.Cvv,
		CardProvider: card.CardProvider,
	}, nil
}

type realSaldoRepo struct {
	repo saldo_repo.Repositories
}

func (r *realSaldoRepo) FindByCardNumber(ctx context.Context, card_number string) (*models.Saldo, error) {
	return r.repo.FindByCardNumber(ctx, card_number)
}
func (r *realSaldoRepo) UpdateSaldoBalance(ctx context.Context, request *requests.UpdateSaldoBalance) (*models.UpdateSaldoBalanceRow, error) {
	res, err := r.repo.UpdateSaldoBalance(ctx, request)
	if err != nil || res == nil {
		return nil, err
	}
	return &models.UpdateSaldoBalanceRow{
		SaldoID:      res.SaldoID,
		CardNumber:   res.CardNumber,
		TotalBalance: int32(res.TotalBalance),
	}, nil
}
func (r *realSaldoRepo) UpdateSaldoWithdraw(ctx context.Context, request *requests.UpdateSaldoWithdraw) (*models.UpdateSaldoWithdrawRow, error) {
	res, err := r.repo.UpdateSaldoWithdraw(ctx, request)
	if err != nil || res == nil {
		return nil, err
	}
	return &models.UpdateSaldoWithdrawRow{
		SaldoID:        res.SaldoID,
		CardNumber:     res.CardNumber,
		TotalBalance:   int32(res.TotalBalance),
		WithdrawAmount: res.WithdrawAmount,
		WithdrawTime:   res.WithdrawTime,
		CreatedAt:      res.CreatedAt,
		UpdatedAt:      res.UpdatedAt,
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
