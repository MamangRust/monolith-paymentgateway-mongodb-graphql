package transaction_test

import (
	"context"
	"time"

	card_repo "github.com/MamangRust/monolith-payment-gateway-card/repository"
	merchant_repo "github.com/MamangRust/monolith-payment-gateway-merchant/repository"
	models "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	saldo_repo "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type realMerchantRepo struct {
	repo merchant_repo.MerchantQueryRepository
}

func (r *realMerchantRepo) FindByApiKey(ctx context.Context, api_key string) (*models.MerchantAllFieldsRow, error) {
	merchant, err := r.repo.FindByApiKey(ctx, api_key)
	if err != nil || merchant == nil {
		return nil, err
	}
	createdAt, updatedAt := time.Time{}, time.Time{}
	if merchant.CreatedAt != nil {
		createdAt = *merchant.CreatedAt
	}
	if merchant.UpdatedAt != nil {
		updatedAt = *merchant.UpdatedAt
	}
	return &models.MerchantAllFieldsRow{
		MerchantID: merchant.MerchantID,
		MerchantNo: merchant.MerchantNo,
		Name:       merchant.Name,
		ApiKey:     merchant.ApiKey,
		UserID:     merchant.UserID,
		Status:     merchant.Status,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

type transactionCardRepo struct {
	query   card_repo.CardQueryRepository
	command card_repo.CardCommandRepository
}

func (r *transactionCardRepo) FindCardByUserId(ctx context.Context, user_id int) (*models.CardAllFieldsRow, error) {
	card, err := r.query.FindCardByUserId(ctx, user_id)
	if err != nil {
		return nil, err
	}
	return &models.CardAllFieldsRow{CardID: card.CardID, UserID: card.UserID, CardNumber: card.CardNumber, CardType: card.CardType, ExpireDate: card.ExpireDate, Cvv: card.Cvv, CardProvider: card.CardProvider, Status: card.Status}, nil
}
func (r *transactionCardRepo) FindUserCardByCardNumber(ctx context.Context, card_number string) (*models.CardByEmailRow, error) {
	card, err := r.query.FindUserCardByCardNumber(ctx, card_number)
	if err != nil {
		return nil, err
	}
	return &models.CardByEmailRow{CardID: card.CardID, UserID: card.UserID, CardNumber: card.CardNumber, CardType: card.CardType, ExpireDate: card.ExpireDate, Cvv: card.Cvv, CardProvider: card.CardProvider}, nil
}
func (r *transactionCardRepo) FindCardByCardNumber(ctx context.Context, card_number string) (*models.CardAllFieldsRow, error) {
	card, err := r.query.FindCardByCardNumber(ctx, card_number)
	if err != nil {
		return nil, err
	}
	return &models.CardAllFieldsRow{CardID: card.CardID, UserID: card.UserID, CardNumber: card.CardNumber, CardType: card.CardType, ExpireDate: card.ExpireDate, Cvv: card.Cvv, CardProvider: card.CardProvider, Status: card.Status}, nil
}
func (r *transactionCardRepo) UpdateCard(ctx context.Context, request *requests.UpdateCardRequest) (*models.CardUpdateRow, error) {
	card, err := r.command.UpdateCard(ctx, request)
	if err != nil {
		return nil, err
	}
	return &models.CardUpdateRow{CardID: card.CardID, UserID: card.UserID, CardNumber: card.CardNumber, CardType: card.CardType, ExpireDate: card.ExpireDate, Cvv: card.Cvv, CardProvider: card.CardProvider}, nil
}
func (r *transactionCardRepo) UpdateCardOutstandingBalance(ctx context.Context, cardID int, outstandingBalance int) (*models.UpdateOutstandingBalanceRow, error) {
	return &models.UpdateOutstandingBalanceRow{}, nil
}
func (r *transactionCardRepo) AddRewardPoints(ctx context.Context, cardID int, points int) (*models.AddRewardPointsRow, error) {
	return &models.AddRewardPointsRow{}, nil
}

type realCardRepo = transactionCardRepo

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
