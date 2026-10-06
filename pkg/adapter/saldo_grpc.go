package adapter

import (
	"context"
	"time"

	pbcard "github.com/MamangRust/monolith-payment-gateway-pb/card"
	pbsaldo "github.com/MamangRust/monolith-payment-gateway-pb/saldo"
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/resilience"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
)

type SaldoAdapter interface {
	FindByCardNumber(ctx context.Context, card_number string) (*salemongo.Saldo, error)
	UpdateSaldoBalance(ctx context.Context, request *requests.UpdateSaldoBalance) (*salemongo.UpdateSaldoBalanceRow, error)
	DebitSaldo(ctx context.Context, request *requests.DebitSaldoRequest) (*salemongo.UpdateSaldoBalanceRow, error)
	CreditSaldo(ctx context.Context, request *requests.CreditSaldoRequest) (*salemongo.UpdateSaldoBalanceRow, error)
	UpdateSaldoWithdraw(ctx context.Context, request *requests.UpdateSaldoWithdraw) (*salemongo.UpdateSaldoWithdrawRow, error)
}

type saldoGRPCAdapter struct {
	QueryClient   pbsaldo.SaldoQueryServiceClient
	CommandClient pbsaldo.SaldoCommandServiceClient
	guard         *resilience.DependencyGuard
}

func (a *saldoGRPCAdapter) SetGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

func NewSaldoAdapter(queryClient pbsaldo.SaldoQueryServiceClient, commandClient pbsaldo.SaldoCommandServiceClient, opts ...GuardOption) SaldoAdapter {
	a := &saldoGRPCAdapter{
		QueryClient:   queryClient,
		CommandClient: commandClient,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *saldoGRPCAdapter) FindByCardNumber(ctx context.Context, card_number string) (*salemongo.Saldo, error) {
	var resp *pbsaldo.ApiResponseSaldo
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.QueryClient.FindByCardNumber(callCtx, &pbcard.FindByCardNumberRequest{
			CardNumber: card_number,
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Data == nil {
		return nil, nil
	}

	return mapSaldoResponse(resp.Data)
}

func mapSaldoResponse(s *pbsaldo.SaldoResponse) (*salemongo.Saldo, error) {
	if s == nil {
		return nil, sharedErrors.NewBadRequestError("saldo response is required")
	}
	saldo := &salemongo.Saldo{
		SaldoID:      s.SaldoId,
		CardNumber:   s.CardNumber,
		TotalBalance: int64(s.TotalBalance),
	}
	if s.WithdrawAmount != 0 {
		amount := int64(s.WithdrawAmount)
		saldo.WithdrawAmount = &amount
	}
	saldo.WithdrawTime = parseTime(s.WithdrawTime)
	saldo.CreatedAt = parseTime(s.CreatedAt)
	saldo.UpdatedAt = parseTime(s.UpdatedAt)
	return saldo, nil
}

func mapMutationBalance(resp *pbsaldo.SaldoResponse) *salemongo.UpdateSaldoBalanceRow {
	if resp == nil {
		return nil
	}
	row := &salemongo.UpdateSaldoBalanceRow{
		SaldoID:      resp.SaldoId,
		CardNumber:   resp.CardNumber,
		TotalBalance: resp.TotalBalance,
		CreatedAt:    parseDate(resp.CreatedAt),
		UpdatedAt:    parseDate(resp.UpdatedAt),
	}
	if resp.WithdrawAmount != 0 {
		amount := resp.WithdrawAmount
		row.WithdrawAmount = &amount
	}
	row.WithdrawTime = parseTime(resp.WithdrawTime)
	return row
}

func mapMutationWithdraw(resp *pbsaldo.SaldoResponse) *salemongo.UpdateSaldoWithdrawRow {
	if resp == nil {
		return nil
	}
	row := &salemongo.UpdateSaldoWithdrawRow{
		SaldoID:      resp.SaldoId,
		CardNumber:   resp.CardNumber,
		TotalBalance: resp.TotalBalance,
		CreatedAt:    parseDate(resp.CreatedAt),
		UpdatedAt:    parseDate(resp.UpdatedAt),
	}
	if resp.WithdrawAmount != 0 {
		amount := resp.WithdrawAmount
		row.WithdrawAmount = &amount
	}
	row.WithdrawTime = parseTime(resp.WithdrawTime)
	return row
}

func (a *saldoGRPCAdapter) UpdateSaldoBalance(ctx context.Context, request *requests.UpdateSaldoBalance) (*salemongo.UpdateSaldoBalanceRow, error) {
	var resp *pbsaldo.ApiResponseSaldo
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.CommandClient.UpdateSaldo(callCtx, &pbsaldo.UpdateSaldoRequest{
			CardNumber:   request.CardNumber,
			TotalBalance: int32(request.TotalBalance),
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, sharedErrors.NewBadRequestError("saldo response is required")
	}
	return mapMutationBalance(resp.Data), nil
}

func (a *saldoGRPCAdapter) DebitSaldo(ctx context.Context, request *requests.DebitSaldoRequest) (*salemongo.UpdateSaldoBalanceRow, error) {
	var resp *pbsaldo.ApiResponseSaldo
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.CommandClient.DebitSaldo(callCtx, &pbsaldo.DebitSaldoRequest{
			CardNumber:  request.CardNumber,
			Amount:      int32(request.Amount),
			OperationId: request.OperationID,
			SourceType:  request.SourceType,
			SourceId:    request.SourceID,
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, sharedErrors.NewBadRequestError("saldo response is required")
	}
	return mapMutationBalance(resp.Data), nil
}

func (a *saldoGRPCAdapter) CreditSaldo(ctx context.Context, request *requests.CreditSaldoRequest) (*salemongo.UpdateSaldoBalanceRow, error) {
	var resp *pbsaldo.ApiResponseSaldo
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.CommandClient.CreditSaldo(callCtx, &pbsaldo.CreditSaldoRequest{
			CardNumber:  request.CardNumber,
			Amount:      int32(request.Amount),
			OperationId: request.OperationID,
			SourceType:  request.SourceType,
			SourceId:    request.SourceID,
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, sharedErrors.NewBadRequestError("saldo response is required")
	}
	return mapMutationBalance(resp.Data), nil
}

func (a *saldoGRPCAdapter) UpdateSaldoWithdraw(ctx context.Context, request *requests.UpdateSaldoWithdraw) (*salemongo.UpdateSaldoWithdrawRow, error) {
	var resp *pbsaldo.ApiResponseSaldo
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.CommandClient.UpdateSaldoWithdraw(callCtx, &pbsaldo.UpdateSaldoWithdrawRequest{
			CardNumber:     request.CardNumber,
			TotalBalance:   int32(request.TotalBalance),
			WithdrawAmount: int32(*request.WithdrawAmount),
			WithdrawTime:   request.WithdrawTime.Format(time.RFC3339),
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, sharedErrors.NewBadRequestError("saldo response is required")
	}
	return mapMutationWithdraw(resp.Data), nil
}
