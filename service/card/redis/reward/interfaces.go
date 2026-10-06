package cardrewardmencache

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
)

type CardRewardCache interface {
	GetBalance(ctx context.Context, cardNumber string) (int64, bool)
	SetBalance(ctx context.Context, cardNumber string, balance int64)
	DeleteBalance(ctx context.Context, cardNumber string)

	GetHistory(ctx context.Context, cardNumber string) ([]*salemongo.CardReward, bool)
	SetHistory(ctx context.Context, cardNumber string, data []*salemongo.CardReward)
	DeleteHistory(ctx context.Context, cardNumber string)
}
