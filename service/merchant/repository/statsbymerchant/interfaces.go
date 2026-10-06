package merchantstatsmerchantrepository

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type MerchantStatsMethodByMerchantRepository interface {
	GetMonthlyPaymentMethodByMerchants(ctx context.Context, req *requests.MonthYearPaymentMethodMerchant) ([]*salemongo.MerchantMonthlyPaymentMethodRow, error)
	GetYearlyPaymentMethodByMerchants(ctx context.Context, req *requests.MonthYearPaymentMethodMerchant) ([]*salemongo.MerchantYearlyPaymentMethodRow, error)
}

type MerchantStatsAmountByMerchantRepository interface {
	GetMonthlyAmountByMerchants(ctx context.Context, req *requests.MonthYearAmountMerchant) ([]*salemongo.MerchantMonthlyAmountRow, error)
	GetYearlyAmountByMerchants(ctx context.Context, req *requests.MonthYearAmountMerchant) ([]*salemongo.MerchantYearlyAmountRow, error)
}

type MerchantStatsTotalAmountByMerchantRepository interface {
	GetMonthlyTotalAmountByMerchants(ctx context.Context, req *requests.MonthYearTotalAmountMerchant) ([]*salemongo.MerchantMonthlyTotalAmountRow, error)
	GetYearlyTotalAmountByMerchants(ctx context.Context, req *requests.MonthYearTotalAmountMerchant) ([]*salemongo.MerchantYearlyTotalAmountRow, error)
}
