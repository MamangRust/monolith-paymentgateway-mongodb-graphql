package withdraw_test

import (
	"context"
	"testing"
	"time"

	pbcard "github.com/MamangRust/monolith-payment-gateway-pb/card"
	pbsaldo "github.com/MamangRust/monolith-payment-gateway-pb/saldo"
	adapter "github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	models "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/idempotency"
	"github.com/MamangRust/monolith-payment-gateway-withdraw/repository"
	"github.com/MamangRust/monolith-payment-gateway-withdraw/service"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// ---------------------------------------------------------------------------
// Fake idempotency store
// ---------------------------------------------------------------------------

type fakeWithdrawIdempotencyStore struct {
	records map[string]*idempotency.Record
}

func newFakeWithdrawIdempotencyStore() *fakeWithdrawIdempotencyStore {
	return &fakeWithdrawIdempotencyStore{records: map[string]*idempotency.Record{}}
}

func (s *fakeWithdrawIdempotencyStore) Claim(_ context.Context, scope, key, requestHash string) (*idempotency.Record, error) {
	id := scope + ":" + key
	if _, ok := s.records[id]; ok {
		return nil, idempotency.ErrKeyInUse
	}
	rec := &idempotency.Record{
		Key:         key,
		RequestHash: requestHash,
		OperationID: "op-1",
		Status:      idempotency.StatusProcessing,
	}
	s.records[id] = rec
	return rec, nil
}

func (s *fakeWithdrawIdempotencyStore) Get(_ context.Context, scope, key string) (*idempotency.Record, error) {
	return s.records[scope+":"+key], nil
}

func (s *fakeWithdrawIdempotencyStore) Complete(_ context.Context, scope, key, requestHash string, outcome idempotency.Outcome) error {
	rec := s.records[scope+":"+key]
	if rec == nil {
		return nil
	}
	rec.Status = outcome.Status
	rec.ResponseJSON = outcome.ResponseJSON
	rec.RequestHash = requestHash
	return nil
}

func (s *fakeWithdrawIdempotencyStore) Release(_ context.Context, scope, key, _ string) error {
	delete(s.records, scope+":"+key)
	return nil
}

// ---------------------------------------------------------------------------
// Fake gRPC clients
// ---------------------------------------------------------------------------

type fakeWithdrawCardQueryClient struct {
	pbcard.CardQueryServiceClient
}

func (f *fakeWithdrawCardQueryClient) FindUserCardByCardNumber(_ context.Context, _ *pbcard.FindByCardNumberRequest, _ ...grpc.CallOption) (*pbcard.CardWithEmailResponse, error) {
	return &pbcard.CardWithEmailResponse{
		Id:         1,
		UserId:     1,
		Email:      "withdraw@example.com",
		CardNumber: "4111111111111111",
		CardType:   "debit",
	}, nil
}

type fakeWithdrawCardCommandClient struct {
	pbcard.CardCommandServiceClient
}

func (f *fakeWithdrawCardCommandClient) UpdateCard(_ context.Context, _ *pbcard.UpdateCardRequest, _ ...grpc.CallOption) (*pbcard.ApiResponseCard, error) {
	return &pbcard.ApiResponseCard{Data: &pbcard.CardResponse{Id: 1, UserId: 1, CardNumber: "4111111111111111", CardType: "debit"}}, nil
}

type fakeWithdrawSaldoQueryClient struct {
	pbsaldo.SaldoQueryServiceClient
}

func (f *fakeWithdrawSaldoQueryClient) FindByCardNumber(_ context.Context, _ *pbcard.FindByCardNumberRequest, _ ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	return &pbsaldo.ApiResponseSaldo{Data: &pbsaldo.SaldoResponse{SaldoId: 1, CardNumber: "4111111111111111", TotalBalance: 100000}}, nil
}

type fakeWithdrawSaldoCommandClient struct {
	pbsaldo.SaldoCommandServiceClient
	balances map[string]int32
}

func newFakeWithdrawSaldoCommandClient() *fakeWithdrawSaldoCommandClient {
	return &fakeWithdrawSaldoCommandClient{balances: map[string]int32{"4111111111111111": 100000}}
}

func (f *fakeWithdrawSaldoCommandClient) UpdateSaldoBalance(_ context.Context, in *pbsaldo.UpdateSaldoBalanceRequest, _ ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	f.balances[in.CardNumber] = in.TotalBalance
	return &pbsaldo.ApiResponseSaldo{Data: &pbsaldo.SaldoResponse{SaldoId: 1, CardNumber: in.CardNumber, TotalBalance: in.TotalBalance}}, nil
}

func (f *fakeWithdrawSaldoCommandClient) DebitSaldo(_ context.Context, in *pbsaldo.DebitSaldoRequest, _ ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	balance := f.balances[in.CardNumber] - in.Amount
	f.balances[in.CardNumber] = balance
	return &pbsaldo.ApiResponseSaldo{Data: &pbsaldo.SaldoResponse{SaldoId: 1, CardNumber: in.CardNumber, TotalBalance: balance}}, nil
}

func (f *fakeWithdrawSaldoCommandClient) CreditSaldo(_ context.Context, in *pbsaldo.CreditSaldoRequest, _ ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	balance := f.balances[in.CardNumber] + in.Amount
	f.balances[in.CardNumber] = balance
	return &pbsaldo.ApiResponseSaldo{Data: &pbsaldo.SaldoResponse{SaldoId: 1, CardNumber: in.CardNumber, TotalBalance: balance}}, nil
}

// ---------------------------------------------------------------------------
// Fake repositories
// ---------------------------------------------------------------------------

type replayWithdrawCommandRepo struct {
	seq     int32
	created map[string]*repository.WithdrawAtomicResult
}

func newReplayWithdrawCommandRepo() *replayWithdrawCommandRepo {
	return &replayWithdrawCommandRepo{created: map[string]*repository.WithdrawAtomicResult{}}
}

func (r *replayWithdrawCommandRepo) CreateWithdrawAtomic(_ context.Context, req *requests.CreateWithdrawRequest) (*repository.WithdrawAtomicResult, error) {
	if req.IdempotencyKey != "" {
		if existing, ok := r.created[req.IdempotencyKey]; ok {
			existing.Replayed = true
			return existing, nil
		}
	}
	res := &repository.WithdrawAtomicResult{
		Row: &models.WithdrawAllFieldsRow{
			WithdrawID:     r.seq + 1,
			CardNumber:     req.CardNumber,
			WithdrawAmount: int32(req.WithdrawAmount),
			Status:         "success",
		},
	}
	r.seq++
	if req.IdempotencyKey != "" {
		r.created[req.IdempotencyKey] = res
	}
	return res, nil
}

func (r *replayWithdrawCommandRepo) CreateWithdraw(_ context.Context, req *requests.CreateWithdrawRequest) (*models.Withdraw, error) {
	r.seq++
	return &models.Withdraw{WithdrawID: r.seq, CardNumber: req.CardNumber, WithdrawAmount: int64(req.WithdrawAmount), Status: "pending"}, nil
}
func (r *replayWithdrawCommandRepo) UpdateWithdraw(_ context.Context, req *requests.UpdateWithdrawRequest) (*models.Withdraw, error) {
	return &models.Withdraw{WithdrawID: int32(*req.WithdrawID), Status: "success"}, nil
}
func (replayWithdrawCommandRepo) UpdateSaldoBalanceDelta(context.Context, string, int) error {
	return nil
}
func (replayWithdrawCommandRepo) UpdateWithdrawAtomic(context.Context, *requests.UpdateWithdrawRequest) (*models.Withdraw, error) {
	return &models.Withdraw{}, nil
}
func (replayWithdrawCommandRepo) UpdateWithdrawStatus(_ context.Context, req *requests.UpdateWithdrawStatus) (*models.Withdraw, error) {
	return &models.Withdraw{WithdrawID: int32(req.WithdrawID), Status: "success"}, nil
}
func (replayWithdrawCommandRepo) TrashedWithdraw(context.Context, int) (*models.Withdraw, error) {
	return &models.Withdraw{}, nil
}
func (replayWithdrawCommandRepo) RestoreWithdraw(context.Context, int) (*models.Withdraw, error) {
	return &models.Withdraw{}, nil
}
func (replayWithdrawCommandRepo) DeleteWithdrawPermanent(context.Context, int) (bool, error) {
	return true, nil
}
func (replayWithdrawCommandRepo) RestoreAllWithdraw(context.Context) (bool, error) {
	return true, nil
}
func (replayWithdrawCommandRepo) DeleteAllWithdrawPermanent(context.Context) (bool, error) {
	return true, nil
}

type replayWithdrawQueryRepo struct{}

func (replayWithdrawQueryRepo) FindAll(context.Context, *requests.FindAllWithdraws) ([]*repository.WithdrawQueryResult, error) {
	return []*repository.WithdrawQueryResult{}, nil
}
func (replayWithdrawQueryRepo) FindByActive(context.Context, *requests.FindAllWithdraws) ([]*repository.WithdrawQueryResult, error) {
	return []*repository.WithdrawQueryResult{}, nil
}
func (replayWithdrawQueryRepo) FindByTrashed(context.Context, *requests.FindAllWithdraws) ([]*repository.WithdrawQueryResult, error) {
	return []*repository.WithdrawQueryResult{}, nil
}
func (replayWithdrawQueryRepo) FindAllByCardNumber(context.Context, *requests.FindAllWithdrawCardNumber) ([]*repository.WithdrawQueryResult, error) {
	return []*repository.WithdrawQueryResult{}, nil
}
func (replayWithdrawQueryRepo) FindById(context.Context, int) (*models.Withdraw, error) {
	return &models.Withdraw{}, nil
}
func (replayWithdrawQueryRepo) GetTodayWithdrawSumByCardNumber(context.Context, string) (int64, error) {
	return 0, nil
}

type replayWithdrawObservability struct{}

func (replayWithdrawObservability) StartTracingAndLogging(ctx context.Context, _ string, _ ...attribute.KeyValue) (context.Context, trace.Span, func(string, ...string), string, func(string, ...zap.Field)) {
	ctx, span := trace.NewNoopTracerProvider().Tracer("withdraw-replay").Start(ctx, "withdraw-replay")
	return ctx, span, func(string, ...string) { span.End() }, "success", func(string, ...zap.Field) {}
}
func (replayWithdrawObservability) RecordMetrics(context.Context, string, string, time.Time) {}

func newReplayWithdrawService(command *replayWithdrawCommandRepo) service.WithdrawCommandService {
	fakeSaldoCmd := newFakeWithdrawSaldoCommandClient()
	ca := adapter.NewCardAdapter(&fakeWithdrawCardQueryClient{}, &fakeWithdrawCardCommandClient{})
	sa := adapter.NewSaldoAdapter(&fakeWithdrawSaldoQueryClient{}, fakeSaldoCmd)

	return service.NewWithdrawCommandService(&service.WithdrawCommandServiceDeps{
		CardAdapter:       ca,
		SaldoAdapter:      sa,
		CommandRepository: command,
		QueryRepository:   replayWithdrawQueryRepo{},
		IdempotencyStore:  newFakeWithdrawIdempotencyStore(),
		Logger:            &logger.Logger{Log: zap.NewNop()},
		Observability:     replayWithdrawObservability{},
	})
}

func TestCreateWithdrawReplayReturnsOriginal(t *testing.T) {
	command := newReplayWithdrawCommandRepo()
	svc := newReplayWithdrawService(command)

	req := &requests.CreateWithdrawRequest{
		CardNumber:     "4111111111111111",
		WithdrawAmount: 100000,
		WithdrawTime:   time.Now(),
		IdempotencyKey: "withdraw:create:user:42:op-1",
	}

	first, err := svc.Create(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, first)
	require.Equal(t, "success", first.Status)

	// Replay with the same key: same withdrawal id, no second record.
	second, err := svc.Create(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, second)
	require.Equal(t, first.WithdrawID, second.WithdrawID, "replay must return the original withdrawal id")
	require.Equal(t, first.WithdrawAmount, second.WithdrawAmount)
}

func TestCreateWithdrawWithoutKeyDebitsTwice(t *testing.T) {
	command := newReplayWithdrawCommandRepo()
	svc := newReplayWithdrawService(command)

	req := &requests.CreateWithdrawRequest{
		CardNumber:     "4111111111111111",
		WithdrawAmount: 100000,
		WithdrawTime:   time.Now(),
	}

	first, err := svc.Create(context.Background(), req)
	require.NoError(t, err)
	second, err := svc.Create(context.Background(), req)
	require.NoError(t, err)

	require.NotEqual(t, first.WithdrawID, second.WithdrawID, "without a key every request creates a new withdrawal")
}
