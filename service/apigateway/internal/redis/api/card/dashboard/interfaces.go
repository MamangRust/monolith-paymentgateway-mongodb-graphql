package card_dashboard_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
)

type CardDashboardTotalCache interface {
	GetDashboardCardCache(ctx context.Context) (*model.APIResponseDashboardCard, bool)
	SetDashboardCardCache(ctx context.Context, data *model.APIResponseDashboardCard)
	DeleteDashboardCardCache(ctx context.Context)
}

type CardDashboardByCardNumberCache interface {
	SetDashboardCardCardNumberCache(ctx context.Context, cardNumber string, data *model.APIResponseDashboardCardNumber)
	GetDashboardCardCardNumberCache(ctx context.Context, cardNumber string) (*model.APIResponseDashboardCardNumber, bool)
	DeleteDashboardCardCardNumberCache(ctx context.Context, cardNumber string)
}
