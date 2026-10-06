package saldo_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
)

type SaldoQueryCache interface {
	GetCachedSaldos(ctx context.Context, req *model.FindAllSaldoInput) (*model.APIResponsePaginationSaldo, bool)
	SetCachedSaldos(ctx context.Context, req *model.FindAllSaldoInput, data *model.APIResponsePaginationSaldo)

	GetCachedSaldoById(ctx context.Context, saldo_id int) (*model.APIResponseSaldo, bool)
	SetCachedSaldoById(ctx context.Context, saldo_id int, data *model.APIResponseSaldo)

	GetCachedSaldoByCardNumber(ctx context.Context, card_number string) (*model.APIResponseSaldo, bool)
	SetCachedSaldoByCardNumber(ctx context.Context, card_number string, data *model.APIResponseSaldo)

	GetCachedSaldoByActive(ctx context.Context, req *model.FindAllSaldoInput) (*model.APIResponsePaginationSaldoDeleteAt, bool)
	SetCachedSaldoByActive(ctx context.Context, req *model.FindAllSaldoInput, data *model.APIResponsePaginationSaldoDeleteAt)

	GetCachedSaldoByTrashed(ctx context.Context, req *model.FindAllSaldoInput) (*model.APIResponsePaginationSaldoDeleteAt, bool)
	SetCachedSaldoByTrashed(ctx context.Context, req *model.FindAllSaldoInput, data *model.APIResponsePaginationSaldoDeleteAt)
}

type SaldoCommandCache interface {
	DeleteSaldoCache(ctx context.Context, saldo_id int)
}
