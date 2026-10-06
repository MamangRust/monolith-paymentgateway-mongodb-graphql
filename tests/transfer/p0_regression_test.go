package transfer_test

import (
	"context"
	"errors"
	"testing"
	"time"

	pbcard "github.com/MamangRust/monolith-payment-gateway-pb/card"
	pbsaldo "github.com/MamangRust/monolith-payment-gateway-pb/saldo"
	adapter "github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	models "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/idempotency"
	"github.com/MamangRust/monolith-payment-gateway-transfer/repository"
	"github.com/MamangRust/monolith-payment-gateway-transfer/service"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// ---------------------------------------------------------------------------
// Fake gRPC clients
// ---------------------------------------------------------------------------

type p0FakeCardQueryClient struct {
	pbcard.CardQueryServiceClient
}

func (f *p0FakeCardQueryClient) FindUserCardByCardNumber(_ context.Context, _ *pbcard.FindByCardNumberRequest, _ ...grpc.CallOption) (*pbcard.CardWithEmailResponse, error) {
	return &pbcard.CardWithEmailResponse{
		Id: 1, UserId: 1, Email: "sender@example.com",
		CardNumber: "sender", CardType: "debit",
	}, nil
}

func (f *p0FakeCardQueryClient) FindByCardNumber(_ context.Context, in *pbcard.FindByCardNumberRequest, _ ...grpc.CallOption) (*pbcard.ApiResponseCard, error) {
	return &pbcard.ApiResponseCard{Data: &pbcard.CardResponse{Id: 1, UserId: 1, CardNumber: in.CardNumber, CardType: "debit"}}, nil
}

type p0FakeCardCommandClient struct {
	pbcard.CardCommandServiceClient
}

func (f *p0FakeCardCommandClient) UpdateCard(_ context.Context, _ *pbcard.UpdateCardRequest, _ ...grpc.CallOption) (*pbcard.ApiResponseCard, error) {
	return &pbcard.ApiResponseCard{Data: &pbcard.CardResponse{Id: 1, UserId: 1, CardNumber: "sender", CardType: "debit"}}, nil
}

type p0FakeSaldoQueryClient struct {
	pbsaldo.SaldoQueryServiceClient
	balances map[string]int64
}

func (f *p0FakeSaldoQueryClient) FindByCardNumber(_ context.Context, in *pbcard.FindByCardNumberRequest, _ ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	bal := f.balances[in.CardNumber]
	return &pbsaldo.ApiResponseSaldo{Data: &pbsaldo.SaldoResponse{SaldoId: 1, CardNumber: in.CardNumber, TotalBalance: int32(bal)}}, nil
}

type p0FakeSaldoCommandClient struct {
	pbsaldo.SaldoCommandServiceClient
	balances map[string]int64
}

func (f *p0FakeSaldoCommandClient) UpdateSaldoBalance(_ context.Context, in *pbsaldo.UpdateSaldoBalanceRequest, _ ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	f.balances[in.CardNumber] = int64(in.TotalBalance)
	return &pbsaldo.ApiResponseSaldo{Data: &pbsaldo.SaldoResponse{SaldoId: 1, CardNumber: in.CardNumber, TotalBalance: in.TotalBalance}}, nil
}

func (f *p0FakeSaldoCommandClient) DebitSaldo(_ context.Context, in *pbsaldo.DebitSaldoRequest, _ ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	balance := f.balances[in.CardNumber] - int64(in.Amount)
	f.balances[in.CardNumber] = balance
	return &pbsaldo.ApiResponseSaldo{Data: &pbsaldo.SaldoResponse{SaldoId: 1, CardNumber: in.CardNumber, TotalBalance: int32(balance)}}, nil
}

func (f *p0FakeSaldoCommandClient) CreditSaldo(_ context.Context, in *pbsaldo.CreditSaldoRequest, _ ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	balance := f.balances[in.CardNumber] + int64(in.Amount)
	f.balances[in.CardNumber] = balance
	return &pbsaldo.ApiResponseSaldo{Data: &pbsaldo.SaldoResponse{SaldoId: 1, CardNumber: in.CardNumber, TotalBalance: int32(balance)}}, nil
}

// ---------------------------------------------------------------------------
// Fake repositories
// ---------------------------------------------------------------------------

type p0TransferCommandRepo struct {
	createErr        error
	statusErr        error
	createTransferID int32
	statusCalls      int
	balances         map[string]int64
}

func (r *p0TransferCommandRepo) CreateTransfer(context.Context, *requests.CreateTransferRequest) (*models.Transfer, error) {
	if r.createErr != nil {
		return nil, r.createErr
	}
	return &models.Transfer{
		TransferID:     r.createTransferID,
		Status:         "success",
		TransferAmount: 50000,
	}, nil
}

func (r *p0TransferCommandRepo) CreateTransferAtomic(_ context.Context, req *requests.CreateTransferRequest) (*repository.TransferAtomicResult, error) {
	if r.createErr != nil {
		return nil, r.createErr
	}
	return &repository.TransferAtomicResult{
		Row: &models.TransferAllFieldsRow{
			TransferID:     r.createTransferID,
			TransferFrom:   req.TransferFrom,
			TransferTo:     req.TransferTo,
			TransferAmount: int32(req.TransferAmount),
			Status:         "success",
		},
	}, nil
}

func (p0TransferCommandRepo) TransitionStatus(_ context.Context, id int, fromStatus, toStatus, reason string) (*models.Transfer, error) {
	return &models.Transfer{TransferID: int32(id), Status: toStatus}, nil
}
func (p0TransferCommandRepo) GuardStatus(_ context.Context, _ int, _, _, _ string) (bool, error) {
	return true, nil
}
func (p0TransferCommandRepo) ListStuck(_ context.Context, _ time.Duration, _ int32) ([]*models.Transfer, error) {
	return []*models.Transfer{}, nil
}
func (r *p0TransferCommandRepo) UpdateTransferStatus(context.Context, *requests.UpdateTransferStatus) (*models.Transfer, error) {
	r.statusCalls++
	if r.statusErr != nil {
		return nil, r.statusErr
	}
	return &models.Transfer{}, nil
}

func (r *p0TransferCommandRepo) UpdateTransferAtomic(_ context.Context, _ models.TransferAllFieldsRow) (*models.Transfer, error) {
	return &models.Transfer{}, nil
}

func (r *p0TransferCommandRepo) UpdateTransferSettlementDelta(_ context.Context, senderCard, receiverCard string, senderDelta, receiverDelta int) error {
	if r.balances != nil {
		r.balances[senderCard] += int64(senderDelta)
		r.balances[receiverCard] += int64(receiverDelta)
	}
	return nil
}

func (p0TransferCommandRepo) UpdateTransfer(context.Context, *requests.UpdateTransferRequest) (*models.Transfer, error) {
	return &models.Transfer{}, nil
}
func (p0TransferCommandRepo) UpdateTransferAmount(context.Context, *requests.UpdateTransferAmountRequest) (*models.Transfer, error) {
	return &models.Transfer{}, nil
}
func (p0TransferCommandRepo) TrashedTransfer(context.Context, int) (*models.Transfer, error) {
	return &models.Transfer{}, nil
}
func (p0TransferCommandRepo) RestoreTransfer(context.Context, int) (*models.Transfer, error) {
	return &models.Transfer{}, nil
}
func (p0TransferCommandRepo) DeleteTransferPermanent(context.Context, int) (bool, error) {
	return true, nil
}
func (p0TransferCommandRepo) RestoreAllTransfer(context.Context) (bool, error) {
	return true, nil
}
func (p0TransferCommandRepo) DeleteAllTransferPermanent(context.Context) (bool, error) {
	return true, nil
}

type p0TransferQueryRepo struct{}

func (p0TransferQueryRepo) FindAll(context.Context, *requests.FindAllTransfers) ([]*repository.TransferQueryResult, error) {
	return []*repository.TransferQueryResult{}, nil
}
func (p0TransferQueryRepo) FindByActive(context.Context, *requests.FindAllTransfers) ([]*repository.TransferQueryResult, error) {
	return []*repository.TransferQueryResult{}, nil
}
func (p0TransferQueryRepo) FindByTrashed(context.Context, *requests.FindAllTransfers) ([]*repository.TransferQueryResult, error) {
	return []*repository.TransferQueryResult{}, nil
}
func (p0TransferQueryRepo) FindById(context.Context, int) (*models.Transfer, error) {
	return &models.Transfer{}, nil
}
func (p0TransferQueryRepo) FindTransferByTransferFrom(context.Context, string) ([]*models.Transfer, error) {
	return []*models.Transfer{}, nil
}
func (p0TransferQueryRepo) FindTransferByTransferTo(context.Context, string) ([]*models.Transfer, error) {
	return []*models.Transfer{}, nil
}

type p0TransferObservability struct{}

func (p0TransferObservability) StartTracingAndLogging(ctx context.Context, _ string, _ ...attribute.KeyValue) (context.Context, trace.Span, func(string, ...string), string, func(string, ...zap.Field)) {
	ctx, span := trace.NewNoopTracerProvider().Tracer("transfer-p0").Start(ctx, "transfer-p0")
	return ctx, span, func(string, ...string) { span.End() }, "success", func(string, ...zap.Field) {}
}
func (p0TransferObservability) RecordMetrics(context.Context, string, string, time.Time) {}

type transferIdempotencyStore struct {
	records map[string]*idempotency.Record
}

func newFakeIdempotencyStore() *transferIdempotencyStore {
	return &transferIdempotencyStore{records: map[string]*idempotency.Record{}}
}

func (s *transferIdempotencyStore) Claim(_ context.Context, scope, key, requestHash string) (*idempotency.Record, error) {
	id := scope + ":" + key
	if _, ok := s.records[id]; ok {
		return nil, idempotency.ErrKeyInUse
	}
	rec := &idempotency.Record{Key: key, RequestHash: requestHash, Status: idempotency.StatusProcessing}
	s.records[id] = rec
	return rec, nil
}
func (s *transferIdempotencyStore) Get(_ context.Context, scope, key string) (*idempotency.Record, error) {
	return s.records[scope+":"+key], nil
}
func (s *transferIdempotencyStore) Complete(_ context.Context, scope, key, requestHash string, outcome idempotency.Outcome) error {
	rec := s.records[scope+":"+key]
	if rec == nil {
		return nil
	}
	rec.Status = outcome.Status
	rec.ResponseJSON = outcome.ResponseJSON
	rec.RequestHash = requestHash
	return nil
}
func (s *transferIdempotencyStore) Release(_ context.Context, scope, key, _ string) error {
	delete(s.records, scope+":"+key)
	return nil
}

func newP0TransferService(balances map[string]int64, command repository.TransferCommandRepository) service.TransferCommandService {
	fakeSaldoQuery := &p0FakeSaldoQueryClient{balances: balances}
	fakeSaldoCmd := &p0FakeSaldoCommandClient{balances: balances}
	sa := adapter.NewSaldoAdapter(fakeSaldoQuery, fakeSaldoCmd)
	ca := adapter.NewCardAdapter(&p0FakeCardQueryClient{}, &p0FakeCardCommandClient{})

	return service.NewTransferCommandService(&service.TransferCommandDeps{
		CardAdapter:               ca,
		SaldoAdapter:              sa,
		TransferQueryRepository:   p0TransferQueryRepo{},
		TransferCommandRepository: command,
		IdempotencyStore:          newFakeIdempotencyStore(),
		Logger:                    &logger.Logger{Log: zap.NewNop()},
		Observability:             p0TransferObservability{},
	})
}

func TestCreateTransferAtomicFailureLeavesBalancesUntouched(t *testing.T) {
	balances := map[string]int64{"sender": 100000, "receiver": 50000}
	command := &p0TransferCommandRepo{createErr: errors.New("create transfer failed")}
	svc := newP0TransferService(balances, command)

	result, err := svc.CreateTransaction(context.Background(), &requests.CreateTransferRequest{
		TransferFrom:   "sender",
		TransferTo:     "receiver",
		TransferAmount: 50000,
	})

	require.Error(t, err)
	require.Nil(t, result)
	require.Equal(t, int64(100000), balances["sender"])
	require.Equal(t, int64(50000), balances["receiver"])
}

func TestCreateTransferAtomicSettlesOnceWithoutStatusUpdate(t *testing.T) {
	balances := map[string]int64{"sender": 100000, "receiver": 50000}
	command := &p0TransferCommandRepo{
		createTransferID: 7,
		statusErr:        errors.New("status update must not be called"),
		balances:         balances,
	}
	svc := newP0TransferService(balances, command)

	result, err := svc.CreateTransaction(context.Background(), &requests.CreateTransferRequest{
		TransferFrom:   "sender",
		TransferTo:     "receiver",
		TransferAmount: 50000,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, int32(7), result.TransferID)
	require.Equal(t, "success", result.Status)
	require.Zero(t, command.statusCalls)
	require.Equal(t, int64(50000), balances["sender"])
	require.Equal(t, int64(100000), balances["receiver"])
}
