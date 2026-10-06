package merchant_stats_bymerchant_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
)

type MerchantStatsMethodByMerchantCache interface {
	GetMonthlyPaymentMethodByMerchantsCache(ctx context.Context, req *model.FindYearMerchantByIDInput) (*model.APIResponseMerchantMonthlyPaymentMethod, bool)
	SetMonthlyPaymentMethodByMerchantsCache(ctx context.Context, req *model.FindYearMerchantByIDInput, data *model.APIResponseMerchantMonthlyPaymentMethod)

	GetYearlyPaymentMethodByMerchantsCache(ctx context.Context, req *model.FindYearMerchantByIDInput) (*model.APIResponseMerchantYearlyPaymentMethod, bool)
	SetYearlyPaymentMethodByMerchantsCache(ctx context.Context, req *model.FindYearMerchantByIDInput, data *model.APIResponseMerchantYearlyPaymentMethod)
}

type MerchantStatsAmountByMerchantCache interface {
	GetMonthlyAmountByMerchantsCache(ctx context.Context, req *model.FindYearMerchantByIDInput) (*model.APIResponseMerchantMonthlyAmount, bool)
	SetMonthlyAmountByMerchantsCache(ctx context.Context, req *model.FindYearMerchantByIDInput, data *model.APIResponseMerchantMonthlyAmount)

	GetYearlyAmountByMerchantsCache(ctx context.Context, req *model.FindYearMerchantByIDInput) (*model.APIResponseMerchantYearlyAmount, bool)
	SetYearlyAmountByMerchantsCache(ctx context.Context, req *model.FindYearMerchantByIDInput, data *model.APIResponseMerchantYearlyAmount)
}

type MerchantStatsTotalAmountByMerchantCache interface {
	GetMonthlyTotalAmountByMerchantsCache(ctx context.Context, req *model.FindYearMerchantByIDInput) (*model.APIResponseMerchantMonthlyTotalAmount, bool)
	SetMonthlyTotalAmountByMerchantsCache(ctx context.Context, req *model.FindYearMerchantByIDInput, data *model.APIResponseMerchantMonthlyTotalAmount)

	GetYearlyTotalAmountByMerchantsCache(ctx context.Context, req *model.FindYearMerchantByIDInput) (*model.APIResponseMerchantYearlyTotalAmount, bool)
	SetYearlyTotalAmountByMerchantsCache(ctx context.Context, req *model.FindYearMerchantByIDInput, data *model.APIResponseMerchantYearlyTotalAmount)
}
