package merchantstatscache

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
)

type MerchantStatsMethodCache interface {
	GetMonthlyPaymentMethodsMerchantCache(ctx context.Context, year int) ([]*salemongo.MerchantMonthlyPaymentMethodRow, bool)
	SetMonthlyPaymentMethodsMerchantCache(ctx context.Context, year int, data []*salemongo.MerchantMonthlyPaymentMethodRow)

	GetYearlyPaymentMethodMerchantCache(ctx context.Context, year int) ([]*salemongo.MerchantYearlyPaymentMethodRow, bool)
	SetYearlyPaymentMethodMerchantCache(ctx context.Context, year int, data []*salemongo.MerchantYearlyPaymentMethodRow)
}

type MerchantStatsAmountCache interface {
	GetMonthlyAmountMerchantCache(ctx context.Context, year int) ([]*salemongo.MerchantMonthlyAmountRow, bool)
	SetMonthlyAmountMerchantCache(ctx context.Context, year int, data []*salemongo.MerchantMonthlyAmountRow)

	GetYearlyAmountMerchantCache(ctx context.Context, year int) ([]*salemongo.MerchantYearlyAmountRow, bool)
	SetYearlyAmountMerchantCache(ctx context.Context, year int, data []*salemongo.MerchantYearlyAmountRow)
}

type MerchantStatsTotalAmountCache interface {
	GetMonthlyTotalAmountMerchantCache(ctx context.Context, year int) ([]*salemongo.MerchantMonthlyTotalAmountRow, bool)
	SetMonthlyTotalAmountMerchantCache(ctx context.Context, year int, data []*salemongo.MerchantMonthlyTotalAmountRow)

	GetYearlyTotalAmountMerchantCache(ctx context.Context, year int) ([]*salemongo.MerchantYearlyTotalAmountRow, bool)
	SetYearlyTotalAmountMerchantCache(ctx context.Context, year int, data []*salemongo.MerchantYearlyTotalAmountRow)
}
