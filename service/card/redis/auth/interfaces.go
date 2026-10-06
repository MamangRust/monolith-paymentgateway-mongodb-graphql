package cardauthmencache

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
)

type CardAuthCache interface {
	GetByTxnID(ctx context.Context, txnID string) (*salemongo.CardAuthTransaction, bool)
	SetByTxnID(ctx context.Context, txnID string, data *salemongo.CardAuthTransaction)
	DeleteByTxnID(ctx context.Context, txnID string)
}
