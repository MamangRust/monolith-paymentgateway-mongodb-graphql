package mencache

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-withdraw/repository"
)

type WithdrawQueryCache interface {
	GetCachedWithdrawsCache(ctx context.Context, req *requests.FindAllWithdraws) ([]*repository.WithdrawQueryResult, *int, bool)
	SetCachedWithdrawsCache(ctx context.Context, req *requests.FindAllWithdraws, data []*repository.WithdrawQueryResult, total *int)

	GetCachedWithdrawByCardCache(ctx context.Context, req *requests.FindAllWithdrawCardNumber) ([]*repository.WithdrawQueryResult, *int, bool)
	SetCachedWithdrawByCardCache(ctx context.Context, req *requests.FindAllWithdrawCardNumber, data []*repository.WithdrawQueryResult, total *int)

	GetCachedWithdrawActiveCache(ctx context.Context, req *requests.FindAllWithdraws) ([]*repository.WithdrawQueryResult, *int, bool)
	SetCachedWithdrawActiveCache(ctx context.Context, req *requests.FindAllWithdraws, data []*repository.WithdrawQueryResult, total *int)

	GetCachedWithdrawTrashedCache(ctx context.Context, req *requests.FindAllWithdraws) ([]*repository.WithdrawQueryResult, *int, bool)
	SetCachedWithdrawTrashedCache(ctx context.Context, req *requests.FindAllWithdraws, data []*repository.WithdrawQueryResult, total *int)

	GetCachedWithdrawCache(ctx context.Context, id int) (*salemongo.Withdraw, bool)
	SetCachedWithdrawCache(ctx context.Context, data *salemongo.Withdraw)
}

type WithdrawCommandCache interface {
	DeleteCachedWithdrawCache(ctx context.Context, id int)
}
