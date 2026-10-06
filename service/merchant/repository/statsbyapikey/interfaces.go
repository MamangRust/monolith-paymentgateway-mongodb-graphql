package merchantstatsapikeyrepository

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type MerchantStatsMethodByApiKeyRepository interface {
	GetMonthlyPaymentMethodByApikey(ctx context.Context, req *requests.MonthYearPaymentMethodApiKey) ([]*salemongo.MerchantMonthlyPaymentMethodRow, error)
	GetYearlyPaymentMethodByApikey(ctx context.Context, req *requests.MonthYearPaymentMethodApiKey) ([]*salemongo.MerchantYearlyPaymentMethodRow, error)
}

type MerchantStatsAmountByApiKeyRepository interface {
	GetMonthlyAmountByApikey(ctx context.Context, req *requests.MonthYearAmountApiKey) ([]*salemongo.MerchantMonthlyAmountRow, error)
	GetYearlyAmountByApikey(ctx context.Context, req *requests.MonthYearAmountApiKey) ([]*salemongo.MerchantYearlyAmountRow, error)
}

type MerchantStatsTotalAmountByApiKeyRepository interface {
	GetMonthlyTotalAmountByApikey(ctx context.Context, req *requests.MonthYearTotalAmountApiKey) ([]*salemongo.MerchantMonthlyTotalAmountRow, error)
	GetYearlyTotalAmountByApikey(ctx context.Context, req *requests.MonthYearTotalAmountApiKey) ([]*salemongo.MerchantYearlyTotalAmountRow, error)
}
