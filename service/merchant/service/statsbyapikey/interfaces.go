package merchantstatsbyapikeyservice

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type MerchantStatsByApiKeyAmountService interface {
	FindMonthlyAmountByApikey(ctx context.Context, req *requests.MonthYearAmountApiKey) ([]*salemongo.MerchantMonthlyAmountRow, error)
	FindYearlyAmountByApikey(ctx context.Context, req *requests.MonthYearAmountApiKey) ([]*salemongo.MerchantYearlyAmountRow, error)
}

type MerchantStatsByApiKeyMethodService interface {
	FindMonthlyPaymentMethodByApikey(ctx context.Context, req *requests.MonthYearPaymentMethodApiKey) ([]*salemongo.MerchantMonthlyPaymentMethodRow, error)
	FindYearlyPaymentMethodByApikey(ctx context.Context, req *requests.MonthYearPaymentMethodApiKey) ([]*salemongo.MerchantYearlyPaymentMethodRow, error)
}

type MerchantStatsByApiKeyTotalAmountService interface {
	FindMonthlyTotalAmountByApikey(ctx context.Context, req *requests.MonthYearTotalAmountApiKey) ([]*salemongo.MerchantMonthlyTotalAmountRow, error)
	FindYearlyTotalAmountByApikey(ctx context.Context, req *requests.MonthYearTotalAmountApiKey) ([]*salemongo.MerchantYearlyTotalAmountRow, error)
}
