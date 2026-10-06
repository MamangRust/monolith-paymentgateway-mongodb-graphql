package transfer_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
)

type TransferQueryCache interface {
	GetCachedTransfersCache(ctx context.Context, req *model.FindAllTransferInput) (*model.APIResponsePaginationTransfer, bool)
	SetCachedTransfersCache(ctx context.Context, req *model.FindAllTransferInput, data *model.APIResponsePaginationTransfer)

	GetCachedTransferActiveCache(ctx context.Context, req *model.FindAllTransferInput) (*model.APIResponsePaginationTransferDeleteAt, bool)
	SetCachedTransferActiveCache(ctx context.Context, req *model.FindAllTransferInput, data *model.APIResponsePaginationTransferDeleteAt)

	GetCachedTransferTrashedCache(ctx context.Context, req *model.FindAllTransferInput) (*model.APIResponsePaginationTransferDeleteAt, bool)
	SetCachedTransferTrashedCache(ctx context.Context, req *model.FindAllTransferInput, data *model.APIResponsePaginationTransferDeleteAt)

	GetCachedTransferCache(ctx context.Context, id int) (*model.APIResponseTransfer, bool)
	SetCachedTransferCache(ctx context.Context, data *model.APIResponseTransfer)

	GetCachedTransferByFrom(ctx context.Context, from string) (*model.APIResponseTransfers, bool)
	SetCachedTransferByFrom(ctx context.Context, from string, data *model.APIResponseTransfers)

	GetCachedTransferByTo(ctx context.Context, to string) (*model.APIResponseTransfers, bool)
	SetCachedTransferByTo(ctx context.Context, to string, data *model.APIResponseTransfers)
}

type TransferCommandCache interface {
	DeleteTransferCache(ctx context.Context, id int)
}
