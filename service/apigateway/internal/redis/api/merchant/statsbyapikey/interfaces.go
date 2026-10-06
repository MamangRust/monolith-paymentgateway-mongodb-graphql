package merchant_stats_byapikey_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
)

type MerchantStatsMethodByApiKeyCache interface {
	GetMonthlyPaymentMethodByApikeysCache(ctx context.Context, req *model.FindYearMerchantByApikeyInput) (*model.APIResponseMerchantMonthlyPaymentMethod, bool)
	SetMonthlyPaymentMethodByApikeysCache(ctx context.Context, req *model.FindYearMerchantByApikeyInput, data *model.APIResponseMerchantMonthlyPaymentMethod)

	GetYearlyPaymentMethodByApikeysCache(ctx context.Context, req *model.FindYearMerchantByApikeyInput) (*model.APIResponseMerchantYearlyPaymentMethod, bool)
	SetYearlyPaymentMethodByApikeysCache(ctx context.Context, req *model.FindYearMerchantByApikeyInput, data *model.APIResponseMerchantYearlyPaymentMethod)
}

type MerchantStatsAmountByApiKeyCache interface {
	GetMonthlyAmountByApikeysCache(ctx context.Context, req *model.FindYearMerchantByApikeyInput) (*model.APIResponseMerchantMonthlyAmount, bool)
	SetMonthlyAmountByApikeysCache(ctx context.Context, req *model.FindYearMerchantByApikeyInput, data *model.APIResponseMerchantMonthlyAmount)

	GetYearlyAmountByApikeysCache(ctx context.Context, req *model.FindYearMerchantByApikeyInput) (*model.APIResponseMerchantYearlyAmount, bool)
	SetYearlyAmountByApikeysCache(ctx context.Context, req *model.FindYearMerchantByApikeyInput, data *model.APIResponseMerchantYearlyAmount)
}

type MerchantStatsTotalAmountByApiKeyCache interface {
	GetMonthlyTotalAmountByApikeysCache(ctx context.Context, req *model.FindYearMerchantByApikeyInput) (*model.APIResponseMerchantMonthlyTotalAmount, bool)
	SetMonthlyTotalAmountByApikeysCache(ctx context.Context, req *model.FindYearMerchantByApikeyInput, data *model.APIResponseMerchantMonthlyTotalAmount)

	GetYearlyTotalAmountByApikeysCache(ctx context.Context, req *model.FindYearMerchantByApikeyInput) (*model.APIResponseMerchantYearlyTotalAmount, bool)
	SetYearlyTotalAmountByApikeysCache(ctx context.Context, req *model.FindYearMerchantByApikeyInput, data *model.APIResponseMerchantYearlyTotalAmount)
}
