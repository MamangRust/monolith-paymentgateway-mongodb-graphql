package mencache

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type SaldoQueryCache interface {
	GetCachedSaldos(ctx context.Context, req *requests.FindAllSaldos) ([]*repository.SaldoResult, *int, bool)
	SetCachedSaldos(ctx context.Context, req *requests.FindAllSaldos, data []*repository.SaldoResult, totalRecords *int)
	GetCachedSaldoById(ctx context.Context, saldo_id int) (*repository.SaldoResult, bool)
	SetCachedSaldoById(ctx context.Context, saldo_id int, data *repository.SaldoResult)
	GetCachedSaldoByCardNumber(ctx context.Context, card_number string) (*salemongo.Saldo, bool)
	SetCachedSaldoByCardNumber(ctx context.Context, card_number string, data *salemongo.Saldo)
	GetCachedSaldoByActive(ctx context.Context, req *requests.FindAllSaldos) ([]*repository.SaldoResult, *int, bool)
	SetCachedSaldoByActive(ctx context.Context, req *requests.FindAllSaldos, data []*repository.SaldoResult, totalRecords *int)
	GetCachedSaldoByTrashed(ctx context.Context, req *requests.FindAllSaldos) ([]*repository.SaldoResult, *int, bool)
	SetCachedSaldoByTrashed(ctx context.Context, req *requests.FindAllSaldos, data []*repository.SaldoResult, totalRecords *int)
}

type SaldoCommandCache interface {
	DeleteSaldoCache(ctx context.Context, saldo_id int)
	DeleteSaldoCacheByCardNumber(ctx context.Context, card_number string)
}
