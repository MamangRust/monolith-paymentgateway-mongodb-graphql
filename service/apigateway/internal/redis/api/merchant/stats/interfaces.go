package merchant_stats_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
)

type MerchantStatsMethodCache interface {
	GetMonthlyPaymentMethodsMerchantCache(ctx context.Context, year int) (*model.APIResponseMerchantMonthlyPaymentMethod, bool)
	SetMonthlyPaymentMethodsMerchantCache(ctx context.Context, year int, data *model.APIResponseMerchantMonthlyPaymentMethod)

	GetYearlyPaymentMethodMerchantCache(ctx context.Context, year int) (*model.APIResponseMerchantYearlyPaymentMethod, bool)
	SetYearlyPaymentMethodMerchantCache(ctx context.Context, year int, data *model.APIResponseMerchantYearlyPaymentMethod)
}

type MerchantStatsAmountCache interface {
	GetMonthlyAmountMerchantCache(ctx context.Context, year int) (*model.APIResponseMerchantMonthlyAmount, bool)
	SetMonthlyAmountMerchantCache(ctx context.Context, year int, data *model.APIResponseMerchantMonthlyAmount)

	GetYearlyAmountMerchantCache(ctx context.Context, year int) (*model.APIResponseMerchantYearlyAmount, bool)
	SetYearlyAmountMerchantCache(ctx context.Context, year int, data *model.APIResponseMerchantYearlyAmount)
}

type MerchantStatsTotalAmountCache interface {
	GetMonthlyTotalAmountMerchantCache(ctx context.Context, year int) (*model.APIResponseMerchantMonthlyTotalAmount, bool)
	SetMonthlyTotalAmountMerchantCache(ctx context.Context, year int, data *model.APIResponseMerchantMonthlyTotalAmount)

	GetYearlyTotalAmountMerchantCache(ctx context.Context, year int) (*model.APIResponseMerchantYearlyTotalAmount, bool)
	SetYearlyTotalAmountMerchantCache(ctx context.Context, year int, data *model.APIResponseMerchantYearlyTotalAmount)
}
