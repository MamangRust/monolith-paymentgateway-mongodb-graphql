package merchantstatsservice

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
)

type MerchantStatsAmountService interface {
	FindMonthlyAmountMerchant(ctx context.Context, year int) ([]*salemongo.MerchantMonthlyAmountRow, error)
	FindYearlyAmountMerchant(ctx context.Context, year int) ([]*salemongo.MerchantYearlyAmountRow, error)
}

type MerchantStatsMethodService interface {
	FindMonthlyPaymentMethodsMerchant(ctx context.Context, year int) ([]*salemongo.MerchantMonthlyPaymentMethodRow, error)
	FindYearlyPaymentMethodMerchant(ctx context.Context, year int) ([]*salemongo.MerchantYearlyPaymentMethodRow, error)
}

type MerchantStatsTotalAmountService interface {
	FindMonthlyTotalAmountMerchant(ctx context.Context, year int) ([]*salemongo.MerchantMonthlyTotalAmountRow, error)
	FindYearlyTotalAmountMerchant(ctx context.Context, year int) ([]*salemongo.MerchantYearlyTotalAmountRow, error)
}
