package merchantstatsrepository

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
)

type MerchantStatsMethodRepository interface {
	GetMonthlyPaymentMethodsMerchant(ctx context.Context, year int) ([]*salemongo.MerchantMonthlyPaymentMethodRow, error)
	GetYearlyPaymentMethodMerchant(ctx context.Context, year int) ([]*salemongo.MerchantYearlyPaymentMethodRow, error)
}

type MerchantStatsAmountRepository interface {
	GetMonthlyAmountMerchant(ctx context.Context, year int) ([]*salemongo.MerchantMonthlyAmountRow, error)
	GetYearlyAmountMerchant(ctx context.Context, year int) ([]*salemongo.MerchantYearlyAmountRow, error)
}

type MerchantStatsTotalAmountRepository interface {
	GetMonthlyTotalAmountMerchant(ctx context.Context, year int) ([]*salemongo.MerchantMonthlyTotalAmountRow, error)
	GetYearlyTotalAmountMerchant(ctx context.Context, year int) ([]*salemongo.MerchantYearlyTotalAmountRow, error)
}
