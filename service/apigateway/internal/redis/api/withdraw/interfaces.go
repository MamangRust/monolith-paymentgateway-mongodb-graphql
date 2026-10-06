package withdraw_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
)

type WithdrawQueryCache interface {
	GetCachedWithdrawsCache(ctx context.Context, req *model.FindAllWithdrawInput) (*model.APIResponsePaginationWithdraw, bool)
	SetCachedWithdrawsCache(ctx context.Context, req *model.FindAllWithdrawInput, data *model.APIResponsePaginationWithdraw)

	GetCachedWithdrawByCardCache(ctx context.Context, req *model.FindAllWithdrawByCardNumberInput) (*model.APIResponsePaginationWithdraw, bool)
	SetCachedWithdrawByCardCache(ctx context.Context, req *model.FindAllWithdrawByCardNumberInput, data *model.APIResponsePaginationWithdraw)

	GetCachedWithdrawActiveCache(ctx context.Context, req *model.FindAllWithdrawInput) (*model.APIResponsePaginationWithdrawDeleteAt, bool)
	SetCachedWithdrawActiveCache(ctx context.Context, req *model.FindAllWithdrawInput, data *model.APIResponsePaginationWithdrawDeleteAt)

	GetCachedWithdrawTrashedCache(ctx context.Context, req *model.FindAllWithdrawInput) (*model.APIResponsePaginationWithdrawDeleteAt, bool)
	SetCachedWithdrawTrashedCache(ctx context.Context, req *model.FindAllWithdrawInput, data *model.APIResponsePaginationWithdrawDeleteAt)

	GetCachedWithdrawCache(ctx context.Context, id int) (*model.APIResponseWithdraw, bool)
	SetCachedWithdrawCache(ctx context.Context, data *model.APIResponseWithdraw)
}

type WithdrawCommandCache interface {
	DeleteCachedWithdrawCache(ctx context.Context, id int)
}
