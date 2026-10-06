package adapter

import (
	"context"
	"fmt"

	pbcard "github.com/MamangRust/monolith-payment-gateway-pb/card"
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/resilience"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CardAdapter interface {
	FindCardByUserId(ctx context.Context, user_id int) (*salemongo.CardAllFieldsRow, error)
	FindUserCardByCardNumber(ctx context.Context, card_number string) (*salemongo.CardByEmailRow, error)
	FindCardByCardNumber(ctx context.Context, card_number string) (*salemongo.CardAllFieldsRow, error)
	UpdateCard(ctx context.Context, request *requests.UpdateCardRequest) (*salemongo.CardUpdateRow, error)
}

type cardGRPCAdapter struct {
	QueryClient   pbcard.CardQueryServiceClient
	CommandClient pbcard.CardCommandServiceClient
	guard         *resilience.DependencyGuard
}

func (a *cardGRPCAdapter) SetGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

func NewCardAdapter(queryClient pbcard.CardQueryServiceClient, commandClient pbcard.CardCommandServiceClient, opts ...GuardOption) CardAdapter {
	a := &cardGRPCAdapter{
		QueryClient:   queryClient,
		CommandClient: commandClient,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *cardGRPCAdapter) FindCardByUserId(ctx context.Context, user_id int) (*salemongo.CardAllFieldsRow, error) {
	var resp *pbcard.ApiResponseCard
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.QueryClient.FindByUserIdCard(callCtx, &pbcard.FindByUserIdCardRequest{
			UserId: int32(user_id),
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Data == nil {
		return nil, nil
	}

	return mapCardAllFields(resp.Data), nil
}

func (a *cardGRPCAdapter) FindUserCardByCardNumber(ctx context.Context, card_number string) (*salemongo.CardByEmailRow, error) {
	var resp *pbcard.CardWithEmailResponse
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.QueryClient.FindUserCardByCardNumber(callCtx, &pbcard.FindByCardNumberRequest{
			CardNumber: card_number,
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("card lookup returned an empty response for card_number=%q", card_number)
	}

	return &salemongo.CardByEmailRow{
		CardID:       resp.Id,
		Email:        resp.Email,
		UserID:       resp.UserId,
		CardNumber:   resp.CardNumber,
		CardType:     resp.CardType,
		ExpireDate:   parseDate(resp.ExpireDate),
		Cvv:          resp.Cvv,
		CardProvider: resp.CardProvider,
		CreatedAt:    parseDate(resp.CreatedAt),
		UpdatedAt:    parseDate(resp.UpdatedAt),
	}, nil
}

func (a *cardGRPCAdapter) FindCardByCardNumber(ctx context.Context, card_number string) (*salemongo.CardAllFieldsRow, error) {
	var resp *pbcard.ApiResponseCard
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

	return mapCardAllFields(resp.Data), nil
}

func (a *cardGRPCAdapter) UpdateCard(ctx context.Context, request *requests.UpdateCardRequest) (*salemongo.CardUpdateRow, error) {
	var resp *pbcard.ApiResponseCard
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.CommandClient.UpdateCard(callCtx, &pbcard.UpdateCardRequest{
			CardId:       int32(request.CardID),
			UserId:       int32(request.UserID),
			CardType:     request.CardType,
			ExpireDate:   timestamppb.New(request.ExpireDate),
			Cvv:          request.CVV,
			CardProvider: request.CardProvider,
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Data == nil {
		return nil, nil
	}

	c := resp.Data
	return &salemongo.CardUpdateRow{
		CardID:       c.Id,
		UserID:       c.UserId,
		CardNumber:   c.CardNumber,
		CardType:     c.CardType,
		ExpireDate:   parseDate(c.ExpireDate),
		Cvv:          c.Cvv,
		CardProvider: c.CardProvider,
		CreatedAt:    parseDate(c.CreatedAt),
		UpdatedAt:    parseDate(c.UpdatedAt),
	}, nil
}

func mapCardAllFields(c *pbcard.CardResponse) *salemongo.CardAllFieldsRow {
	if c == nil {
		return nil
	}

	return &salemongo.CardAllFieldsRow{
		CardID:             c.Id,
		UserID:             c.UserId,
		CardNumber:         c.CardNumber,
		CardType:           c.CardType,
		ExpireDate:         parseDate(c.ExpireDate),
		Cvv:                c.Cvv,
		CardProvider:       c.CardProvider,
		Status:             c.Status,
		CreditLimit:        c.CreditLimit,
		OutstandingBalance: c.OutstandingBalance,
		RewardPoints:       c.RewardPoints,
		CreatedAt:          parseDate(c.CreatedAt),
		UpdatedAt:          parseDate(c.UpdatedAt),
	}
}
