package adapter

import (
	"context"

	pbmerchant "github.com/MamangRust/monolith-payment-gateway-pb/merchant"
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/resilience"
)

type MerchantAdapter interface {
	FindByApiKey(ctx context.Context, api_key string) (*salemongo.MerchantAllFieldsRow, error)
	FindByMerchantId(ctx context.Context, merchant_id int) (*salemongo.MerchantAllFieldsRow, error)
}

type merchantGRPCAdapter struct {
	QueryClient pbmerchant.MerchantQueryServiceClient
	guard       *resilience.DependencyGuard
}

func (a *merchantGRPCAdapter) SetGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

func NewMerchantAdapter(queryClient pbmerchant.MerchantQueryServiceClient, opts ...GuardOption) MerchantAdapter {
	a := &merchantGRPCAdapter{
		QueryClient: queryClient,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *merchantGRPCAdapter) FindByApiKey(ctx context.Context, api_key string) (*salemongo.MerchantAllFieldsRow, error) {
	var resp *pbmerchant.ApiResponseMerchant
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.QueryClient.FindByApiKey(callCtx, &pbmerchant.FindByApiKeyRequest{
			ApiKey: api_key,
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return mapMerchantResponse(resp), nil
}

func (a *merchantGRPCAdapter) FindByMerchantId(ctx context.Context, merchant_id int) (*salemongo.MerchantAllFieldsRow, error) {
	var resp *pbmerchant.ApiResponseMerchant
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.QueryClient.FindByIdMerchant(callCtx, &pbmerchant.FindByIdMerchantRequest{
			MerchantId: int32(merchant_id),
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return mapMerchantResponse(resp), nil
}

func mapMerchantResponse(resp *pbmerchant.ApiResponseMerchant) *salemongo.MerchantAllFieldsRow {
	if resp == nil || resp.Data == nil {
		return nil
	}

	return &salemongo.MerchantAllFieldsRow{
		MerchantID: resp.Data.Id,
		Name:       resp.Data.Name,
		ApiKey:     resp.Data.ApiKey,
		UserID:     resp.Data.UserId,
		Status:     resp.Data.Status,
		CreatedAt:  parseDate(resp.Data.CreatedAt),
		UpdatedAt:  parseDate(resp.Data.UpdatedAt),
	}
}
