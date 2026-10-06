package cardbillingmencache

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
)

type CardBillingCache interface {
	GetByCardNumber(ctx context.Context, cardNumber string) ([]*salemongo.BillingCycle, bool)
	SetByCardNumber(ctx context.Context, cardNumber string, data []*salemongo.BillingCycle)
	DeleteByCardNumber(ctx context.Context, cardNumber string)
}
