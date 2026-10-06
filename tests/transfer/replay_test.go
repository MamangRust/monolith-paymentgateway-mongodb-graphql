package transfer_test

import (
	"context"
	"testing"
	"time"

	models "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-transfer/repository"
	"github.com/MamangRust/monolith-payment-gateway-transfer/service"
	"github.com/stretchr/testify/require"
)

// replayTransferCommandRepo mirrors the repository contract for
// CreateTransferAtomic: the same idempotency key returns the ORIGINAL result
// (Replayed=true), never a second settlement.
type replayTransferCommandRepo struct {
	seq     int32
	created map[string]*repository.TransferAtomicResult
}

func (r *replayTransferCommandRepo) CreateTransferAtomic(_ context.Context, req *requests.CreateTransferRequest) (*repository.TransferAtomicResult, error) {
	if req.IdempotencyKey != "" {
		if existing, ok := r.created[req.IdempotencyKey]; ok {
			existing.Replayed = true
			return existing, nil
		}
	}
	res := &repository.TransferAtomicResult{
		Row: &models.TransferAllFieldsRow{
			TransferID:     r.seq + 1,
			TransferFrom:   req.TransferFrom,
			TransferTo:     req.TransferTo,
			TransferAmount: int32(req.TransferAmount),
			Status:         "success",
		},
	}
	r.seq++
	if req.IdempotencyKey != "" {
		r.created[req.IdempotencyKey] = res
	}
	return res, nil
}

func (r *replayTransferCommandRepo) CreateTransfer(_ context.Context, req *requests.CreateTransferRequest) (*models.Transfer, error) {
	r.seq++
	return &models.Transfer{TransferID: r.seq, TransferFrom: req.TransferFrom, TransferTo: req.TransferTo, TransferAmount: int64(req.TransferAmount), Status: "pending"}, nil
}
func (replayTransferCommandRepo) TransitionStatus(_ context.Context, id int, _, toStatus, _ string) (*models.Transfer, error) {
	return &models.Transfer{TransferID: int32(id), Status: toStatus}, nil
}
func (replayTransferCommandRepo) GuardStatus(_ context.Context, _ int, _, _, _ string) (bool, error) {
	return true, nil
}
func (replayTransferCommandRepo) ListStuck(_ context.Context, _ time.Duration, _ int32) ([]*models.Transfer, error) {
	return []*models.Transfer{}, nil
}
func (replayTransferCommandRepo) UpdateTransferStatus(context.Context, *requests.UpdateTransferStatus) (*models.Transfer, error) {
	return &models.Transfer{}, nil
}
func (replayTransferCommandRepo) UpdateTransferSettlementDelta(context.Context, string, string, int, int) error {
	return nil
}
func (replayTransferCommandRepo) UpdateTransfer(context.Context, *requests.UpdateTransferRequest) (*models.Transfer, error) {
	return &models.Transfer{}, nil
}
func (replayTransferCommandRepo) UpdateTransferAmount(context.Context, *requests.UpdateTransferAmountRequest) (*models.Transfer, error) {
	return &models.Transfer{}, nil
}
func (replayTransferCommandRepo) TrashedTransfer(context.Context, int) (*models.Transfer, error) {
	return &models.Transfer{}, nil
}
func (replayTransferCommandRepo) RestoreTransfer(context.Context, int) (*models.Transfer, error) {
	return &models.Transfer{}, nil
}
func (replayTransferCommandRepo) DeleteTransferPermanent(context.Context, int) (bool, error) {
	return true, nil
}
func (replayTransferCommandRepo) RestoreAllTransfer(context.Context) (bool, error) {
	return true, nil
}
func (replayTransferCommandRepo) DeleteAllTransferPermanent(context.Context) (bool, error) {
	return true, nil
}

func newReplayTransferService(command repository.TransferCommandRepository) service.TransferCommandService {
	return newP0TransferService(map[string]int64{"sender": 100000, "receiver": 50000}, command)
}

func TestCreateTransferReplayReturnsOriginalSettlement(t *testing.T) {
	command := &replayTransferCommandRepo{created: map[string]*repository.TransferAtomicResult{}}
	svc := newReplayTransferService(command)

	req := &requests.CreateTransferRequest{
		TransferFrom:   "sender",
		TransferTo:     "receiver",
		TransferAmount: 50000,
		IdempotencyKey: "transfer:create:user:7:op-1",
	}

	first, err := svc.CreateTransaction(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, first)
	require.Equal(t, int32(1), first.TransferID)
	require.Equal(t, "success", first.Status)

	second, err := svc.CreateTransaction(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, second)
	require.Equal(t, first.TransferID, second.TransferID, "replay must return the original settlement id")
	require.Equal(t, first.TransferAmount, second.TransferAmount)
}

func TestCreateTransferWithoutKeySettlesTwice(t *testing.T) {
	command := &replayTransferCommandRepo{created: map[string]*repository.TransferAtomicResult{}}
	svc := newReplayTransferService(command)

	req := &requests.CreateTransferRequest{
		TransferFrom:   "sender",
		TransferTo:     "receiver",
		TransferAmount: 50000,
	}

	first, err := svc.CreateTransaction(context.Background(), req)
	require.NoError(t, err)
	second, err := svc.CreateTransaction(context.Background(), req)
	require.NoError(t, err)

	require.NotEqual(t, first.TransferID, second.TransferID, "without a key every request creates a new settlement")
}
