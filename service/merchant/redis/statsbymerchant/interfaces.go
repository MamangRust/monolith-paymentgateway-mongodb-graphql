package merchantstatsbymerchant

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type MerchantStatsMethodByMerchantCache interface {
	GetMonthlyPaymentMethodByMerchantsCache(ctx context.Context, req *requests.MonthYearPaymentMethodMerchant) ([]*salemongo.MerchantMonthlyPaymentMethodRow, bool)
	SetMonthlyPaymentMethodByMerchantsCache(ctx context.Context, req *requests.MonthYearPaymentMethodMerchant, data []*salemongo.MerchantMonthlyPaymentMethodRow)

	GetYearlyPaymentMethodByMerchantsCache(ctx context.Context, req *requests.MonthYearPaymentMethodMerchant) ([]*salemongo.MerchantYearlyPaymentMethodRow, bool)
	SetYearlyPaymentMethodByMerchantsCache(ctx context.Context, req *requests.MonthYearPaymentMethodMerchant, data []*salemongo.MerchantYearlyPaymentMethodRow)
}

type MerchantStatsAmountByMerchantCache interface {
	GetMonthlyAmountByMerchantsCache(ctx context.Context, req *requests.MonthYearAmountMerchant) ([]*salemongo.MerchantMonthlyAmountRow, bool)
	SetMonthlyAmountByMerchantsCache(ctx context.Context, req *requests.MonthYearAmountMerchant, data []*salemongo.MerchantMonthlyAmountRow)

	GetYearlyAmountByMerchantsCache(ctx context.Context, req *requests.MonthYearAmountMerchant) ([]*salemongo.MerchantYearlyAmountRow, bool)
	SetYearlyAmountByMerchantsCache(ctx context.Context, req *requests.MonthYearAmountMerchant, data []*salemongo.MerchantYearlyAmountRow)
}

type MerchantStatsTotalAmountByMerchantCache interface {
	GetMonthlyTotalAmountByMerchantsCache(ctx context.Context, req *requests.MonthYearTotalAmountMerchant) ([]*salemongo.MerchantMonthlyTotalAmountRow, bool)
	SetMonthlyTotalAmountByMerchantsCache(ctx context.Context, req *requests.MonthYearTotalAmountMerchant, data []*salemongo.MerchantMonthlyTotalAmountRow)

	GetYearlyTotalAmountByMerchantsCache(ctx context.Context, req *requests.MonthYearTotalAmountMerchant) ([]*salemongo.MerchantYearlyTotalAmountRow, bool)
	SetYearlyTotalAmountByMerchantsCache(ctx context.Context, req *requests.MonthYearTotalAmountMerchant, data []*salemongo.MerchantYearlyTotalAmountRow)
}
