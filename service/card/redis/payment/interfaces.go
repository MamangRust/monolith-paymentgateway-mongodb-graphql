package cardpaymentmencache

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
)

type CardPaymentCache interface {
	GetPaymentHistory(ctx context.Context, cardNumber string, page, pageSize int) ([]*salemongo.CardPaymentListRow, bool)
	SetPaymentHistory(ctx context.Context, cardNumber string, page, pageSize int, data []*salemongo.CardPaymentListRow)
	DeletePaymentHistory(ctx context.Context, cardNumber string)

	GetPaymentCount(ctx context.Context, cardNumber string) (int, bool)
	SetPaymentCount(ctx context.Context, cardNumber string, count int)
	DeletePaymentCount(ctx context.Context, cardNumber string)
}
