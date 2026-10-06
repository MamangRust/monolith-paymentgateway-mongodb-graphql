package service

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-withdraw/repository"
)

type WithdrawQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllWithdraws) ([]*repository.WithdrawQueryResult, *int, error)
	FindAllByCardNumber(ctx context.Context, req *requests.FindAllWithdrawCardNumber) ([]*repository.WithdrawQueryResult, *int, error)
	FindById(ctx context.Context, withdrawID int) (*salemongo.Withdraw, error)
	FindByActive(ctx context.Context, req *requests.FindAllWithdraws) ([]*repository.WithdrawQueryResult, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllWithdraws) ([]*repository.WithdrawQueryResult, *int, error)
}

type WithdrawCommandService interface {
	Create(ctx context.Context, request *requests.CreateWithdrawRequest) (*salemongo.Withdraw, error)
	Update(ctx context.Context, request *requests.UpdateWithdrawRequest) (*salemongo.Withdraw, error)
	TrashedWithdraw(ctx context.Context, withdraw_id int) (*salemongo.Withdraw, error)
	RestoreWithdraw(ctx context.Context, withdraw_id int) (*salemongo.Withdraw, error)
	DeleteWithdrawPermanent(ctx context.Context, withdraw_id int) (bool, error)

	RestoreAllWithdraw(ctx context.Context) (bool, error)
	DeleteAllWithdrawPermanent(ctx context.Context) (bool, error)
}
