package mencache

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-transfer/repository"
)

type TransferQueryCache interface {
	GetCachedTransfersCache(ctx context.Context, req *requests.FindAllTransfers) ([]*repository.TransferQueryResult, *int, bool)
	SetCachedTransfersCache(ctx context.Context, req *requests.FindAllTransfers, data []*repository.TransferQueryResult, total *int)

	GetCachedTransferActiveCache(ctx context.Context, req *requests.FindAllTransfers) ([]*repository.TransferQueryResult, *int, bool)
	SetCachedTransferActiveCache(ctx context.Context, req *requests.FindAllTransfers, data []*repository.TransferQueryResult, total *int)

	GetCachedTransferTrashedCache(ctx context.Context, req *requests.FindAllTransfers) ([]*repository.TransferQueryResult, *int, bool)
	SetCachedTransferTrashedCache(ctx context.Context, req *requests.FindAllTransfers, data []*repository.TransferQueryResult, total *int)

	GetCachedTransferCache(ctx context.Context, id int) (*salemongo.Transfer, bool)
	SetCachedTransferCache(ctx context.Context, data *salemongo.Transfer)

	GetCachedTransferByFrom(ctx context.Context, from string) ([]*salemongo.Transfer, bool)
	SetCachedTransferByFrom(ctx context.Context, from string, data []*salemongo.Transfer)

	GetCachedTransferByTo(ctx context.Context, to string) ([]*salemongo.Transfer, bool)
	SetCachedTransferByTo(ctx context.Context, to string, data []*salemongo.Transfer)
}

type TransferCommandCache interface {
	DeleteTransferCache(ctx context.Context, id int)
}
