package service

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-transfer/repository"
)

type TransferQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllTransfers) ([]*repository.TransferQueryResult, *int, error)
	FindById(ctx context.Context, transferId int) (*salemongo.Transfer, error)
	FindByActive(ctx context.Context, req *requests.FindAllTransfers) ([]*repository.TransferQueryResult, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllTransfers) ([]*repository.TransferQueryResult, *int, error)
	FindTransferByTransferFrom(ctx context.Context, transfer_from string) ([]*salemongo.Transfer, error)
	FindTransferByTransferTo(ctx context.Context, transfer_to string) ([]*salemongo.Transfer, error)
}

type TransferCommandService interface {
	CreateTransaction(ctx context.Context, request *requests.CreateTransferRequest) (*salemongo.Transfer, error)
	UpdateTransaction(ctx context.Context, request *requests.UpdateTransferRequest) (*salemongo.Transfer, error)
	TrashedTransfer(ctx context.Context, transfer_id int) (*salemongo.Transfer, error)
	RestoreTransfer(ctx context.Context, transfer_id int) (*salemongo.Transfer, error)
	DeleteTransferPermanent(ctx context.Context, transfer_id int) (bool, error)

	RestoreAllTransfer(ctx context.Context) (bool, error)
	DeleteAllTransferPermanent(ctx context.Context) (bool, error)
	RecoverStuckOperations(ctx context.Context, olderThan time.Duration, maxRows int32) error
	StartRecoveryWorker(ctx context.Context, interval, olderThan time.Duration, maxRows int32)
}
