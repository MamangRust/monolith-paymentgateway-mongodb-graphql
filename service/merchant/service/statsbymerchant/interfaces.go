package merchantstatsbymerchantservice

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type MerchantStatsByMerchantAmountService interface {
	FindMonthlyAmountByMerchants(ctx context.Context, req *requests.MonthYearAmountMerchant) ([]*salemongo.MerchantMonthlyAmountRow, error)
	FindYearlyAmountByMerchants(ctx context.Context, req *requests.MonthYearAmountMerchant) ([]*salemongo.MerchantYearlyAmountRow, error)
}

type MerchantStatsByMerchantMethodService interface {
	FindMonthlyPaymentMethodByMerchants(ctx context.Context, req *requests.MonthYearPaymentMethodMerchant) ([]*salemongo.MerchantMonthlyPaymentMethodRow, error)
	FindYearlyPaymentMethodByMerchants(ctx context.Context, req *requests.MonthYearPaymentMethodMerchant) ([]*salemongo.MerchantYearlyPaymentMethodRow, error)
}

type MerchantStatsByMerchantTotalAmountService interface {
	FindMonthlyTotalAmountByMerchants(ctx context.Context, req *requests.MonthYearTotalAmountMerchant) ([]*salemongo.MerchantMonthlyTotalAmountRow, error)
	FindYearlyTotalAmountByMerchants(ctx context.Context, req *requests.MonthYearTotalAmountMerchant) ([]*salemongo.MerchantYearlyTotalAmountRow, error)
}
