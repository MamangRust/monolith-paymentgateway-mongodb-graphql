package merchantstatsapikey

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type MerchantStatsMethodByApiKeyCache interface {
	GetMonthlyPaymentMethodByApikeysCache(ctx context.Context, req *requests.MonthYearPaymentMethodApiKey) ([]*salemongo.MerchantMonthlyPaymentMethodRow, bool)
	SetMonthlyPaymentMethodByApikeysCache(ctx context.Context, req *requests.MonthYearPaymentMethodApiKey, data []*salemongo.MerchantMonthlyPaymentMethodRow)

	GetYearlyPaymentMethodByApikeysCache(ctx context.Context, req *requests.MonthYearPaymentMethodApiKey) ([]*salemongo.MerchantYearlyPaymentMethodRow, bool)
	SetYearlyPaymentMethodByApikeysCache(ctx context.Context, req *requests.MonthYearPaymentMethodApiKey, data []*salemongo.MerchantYearlyPaymentMethodRow)
}

type MerchantStatsAmountByApiKeyCache interface {
	GetMonthlyAmountByApikeysCache(ctx context.Context, req *requests.MonthYearAmountApiKey) ([]*salemongo.MerchantMonthlyAmountRow, bool)
	SetMonthlyAmountByApikeysCache(ctx context.Context, req *requests.MonthYearAmountApiKey, data []*salemongo.MerchantMonthlyAmountRow)

	GetYearlyAmountByApikeysCache(ctx context.Context, req *requests.MonthYearAmountApiKey) ([]*salemongo.MerchantYearlyAmountRow, bool)
	SetYearlyAmountByApikeysCache(ctx context.Context, req *requests.MonthYearAmountApiKey, data []*salemongo.MerchantYearlyAmountRow)
}

type MerchantStatsTotalAmountByApiKeyCache interface {
	GetMonthlyTotalAmountByApikeysCache(ctx context.Context, req *requests.MonthYearTotalAmountApiKey) ([]*salemongo.MerchantMonthlyTotalAmountRow, bool)
	SetMonthlyTotalAmountByApikeysCache(ctx context.Context, req *requests.MonthYearTotalAmountApiKey, data []*salemongo.MerchantMonthlyTotalAmountRow)

	GetYearlyTotalAmountByApikeysCache(ctx context.Context, req *requests.MonthYearTotalAmountApiKey) ([]*salemongo.MerchantYearlyTotalAmountRow, bool)
	SetYearlyTotalAmountByApikeysCache(ctx context.Context, req *requests.MonthYearTotalAmountApiKey, data []*salemongo.MerchantYearlyTotalAmountRow)
}
