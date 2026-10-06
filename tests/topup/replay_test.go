package topup_test

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
	"github.com/MamangRust/monolith-payment-gateway-topup/repository"
	"github.com/MamangRust/monolith-payment-gateway-topup/service"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// ---------------------------------------------------------------------------
// Fake idempotency store — an in-memory implementation of idempotency.Store
// that mimics the MongoDB-backed repository: Claim inserts a processing
// record, Get replays a completed response, Complete finalizes it.
// ---------------------------------------------------------------------------

type fakeIdempotencyStore struct {
	records map[string]*idempotency.Record
}

func newFakeIdempotencyStore() *fakeIdempotencyStore {
	return &fakeIdempotencyStore{records: map[string]*idempotency.Record{}}
}

func (s *fakeIdempotencyStore) Claim(_ context.Context, scope, key, requestHash string) (*idempotency.Record, error) {
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

func (s *fakeIdempotencyStore) Get(_ context.Context, scope, key string) (*idempotency.Record, error) {
	return s.records[scope+":"+key], nil
}

func (s *fakeIdempotencyStore) Complete(_ context.Context, scope, key, requestHash string, outcome idempotency.Outcome) error {
	rec := s.records[scope+":"+key]
	if rec == nil {
		return nil
	}
	rec.Status = outcome.Status
	rec.ResponseJSON = outcome.ResponseJSON
	rec.RequestHash = requestHash
	return nil
}

func (s *fakeIdempotencyStore) Release(_ context.Context, scope, key, _ string) error {
	delete(s.records, scope+":"+key)
	return nil
}

// ---------------------------------------------------------------------------
// Fake gRPC clients — satisfy the pb client interfaces the production
// adapters wrap, recording how often the card metadata side-effect runs.
// ---------------------------------------------------------------------------

type fakeCardQueryClient struct {
	pbcard.CardQueryServiceClient
}

func (f *fakeCardQueryClient) FindUserCardByCardNumber(_ context.Context, _ *pbcard.FindByCardNumberRequest, _ ...grpc.CallOption) (*pbcard.CardWithEmailResponse, error) {
	return &pbcard.CardWithEmailResponse{
		Id:         1,
		UserId:     1,
		Email:      "topup@example.com",
		CardNumber: "4111111111111111",
		CardType:   "debit",
	}, nil
}

type fakeCardCommandClient struct {
	pbcard.CardCommandServiceClient
	updateCalls int
}

func (f *fakeCardCommandClient) UpdateCard(_ context.Context, _ *pbcard.UpdateCardRequest, _ ...grpc.CallOption) (*pbcard.ApiResponseCard, error) {
	f.updateCalls++
	return &pbcard.ApiResponseCard{Data: &pbcard.CardResponse{Id: 1, UserId: 1, CardNumber: "4111111111111111", CardType: "debit"}}, nil
}

type fakeSaldoQueryClient struct {
	pbsaldo.SaldoQueryServiceClient
}

func (f *fakeSaldoQueryClient) FindByCardNumber(_ context.Context, _ *pbcard.FindByCardNumberRequest, _ ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	return &pbsaldo.ApiResponseSaldo{Data: &pbsaldo.SaldoResponse{SaldoId: 1, CardNumber: "4111111111111111", TotalBalance: 100000}}, nil
}

type fakeSaldoCommandClient struct {
	pbsaldo.SaldoCommandServiceClient
}

func (f *fakeSaldoCommandClient) UpdateSaldoBalance(_ context.Context, in *pbsaldo.UpdateSaldoBalanceRequest, _ ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	return &pbsaldo.ApiResponseSaldo{Data: &pbsaldo.SaldoResponse{SaldoId: 1, CardNumber: in.CardNumber, TotalBalance: in.TotalBalance}}, nil
}

func (f *fakeSaldoCommandClient) DebitSaldo(_ context.Context, in *pbsaldo.DebitSaldoRequest, _ ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	return &pbsaldo.ApiResponseSaldo{Data: &pbsaldo.SaldoResponse{SaldoId: 1, CardNumber: in.CardNumber, TotalBalance: 100000 - in.Amount}}, nil
}

func (f *fakeSaldoCommandClient) CreditSaldo(_ context.Context, in *pbsaldo.CreditSaldoRequest, _ ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	return &pbsaldo.ApiResponseSaldo{Data: &pbsaldo.SaldoResponse{SaldoId: 1, CardNumber: in.CardNumber, TotalBalance: 100000 + in.Amount}}, nil
}

// ---------------------------------------------------------------------------
// Fake repositories
// ---------------------------------------------------------------------------

// replayTopupCommandRepo mirrors the repository contract: creating a topup
// without an idempotency key always inserts a new record.
type replayTopupCommandRepo struct {
	seq     int32
	updates map[int32]int
}

func newReplayTopupCommandRepo() *replayTopupCommandRepo {
	return &replayTopupCommandRepo{updates: map[int32]int{}}
}

func (r *replayTopupCommandRepo) CreateTopup(_ context.Context, request *requests.CreateTopupRequest) (*models.Topup, error) {
	r.seq++
	return &models.Topup{
		TopupID:     r.seq,
		CardNumber:  request.CardNumber,
		TopupAmount: int64(request.TopupAmount),
		TopupMethod: request.TopupMethod,
		Status:      "pending",
	}, nil
}

func (r *replayTopupCommandRepo) UpdateTopupStatus(_ context.Context, request *requests.UpdateTopupStatus) (*models.Topup, error) {
	r.updates[int32(request.TopupID)]++
	return &models.Topup{
		TopupID:     int32(request.TopupID),
		CardNumber:  "4111111111111111",
		TopupAmount: 100000,
		Status:      request.Status,
	}, nil
}

func (r *replayTopupCommandRepo) UpdateTopup(context.Context, *requests.UpdateTopupRequest) (*models.Topup, error) {
	return &models.Topup{}, nil
}

func (r *replayTopupCommandRepo) UpdateTopupAmount(context.Context, *requests.UpdateTopupAmount) (*models.Topup, error) {
	return &models.Topup{}, nil
}

func (r *replayTopupCommandRepo) CompareAndSwapTopupAmount(context.Context, int, int, int) (bool, error) {
	return true, nil
}

func (r *replayTopupCommandRepo) TrashedTopup(context.Context, int) (*models.Topup, error) {
	return &models.Topup{}, nil
}

func (r *replayTopupCommandRepo) RestoreTopup(context.Context, int) (*models.Topup, error) {
	return &models.Topup{}, nil
}

func (r *replayTopupCommandRepo) DeleteTopupPermanent(context.Context, int) (bool, error) {
	return true, nil
}

func (r *replayTopupCommandRepo) RestoreAllTopup(context.Context) (bool, error) {
	return true, nil
}

func (r *replayTopupCommandRepo) DeleteAllTopupPermanent(context.Context) (bool, error) {
	return true, nil
}

type replayTopupQueryRepo struct{}

func (replayTopupQueryRepo) FindAllTopups(context.Context, *requests.FindAllTopups) ([]*repository.TopupQueryResult, error) {
	return []*repository.TopupQueryResult{}, nil
}

func (replayTopupQueryRepo) FindByActive(context.Context, *requests.FindAllTopups) ([]*repository.TopupQueryResult, error) {
	return []*repository.TopupQueryResult{}, nil
}

func (replayTopupQueryRepo) FindByTrashed(context.Context, *requests.FindAllTopups) ([]*repository.TopupQueryResult, error) {
	return []*repository.TopupQueryResult{}, nil
}

func (replayTopupQueryRepo) FindAllTopupByCardNumber(context.Context, *requests.FindAllTopupsByCardNumber) ([]*repository.TopupByCardResult, error) {
	return []*repository.TopupByCardResult{}, nil
}

func (replayTopupQueryRepo) FindById(context.Context, int) (*models.Topup, error) {
	return &models.Topup{}, nil
}

type replayTopupObservability struct{}

func (replayTopupObservability) StartTracingAndLogging(ctx context.Context, _ string, _ ...attribute.KeyValue) (context.Context, trace.Span, func(string, ...string), string, func(string, ...zap.Field)) {
	ctx, span := trace.NewNoopTracerProvider().Tracer("topup-replay").Start(ctx, "topup-replay")
	return ctx, span, func(string, ...string) { span.End() }, "success", func(string, ...zap.Field) {}
}

func (replayTopupObservability) RecordMetrics(context.Context, string, string, time.Time) {}

// ---------------------------------------------------------------------------
// Service wiring
// ---------------------------------------------------------------------------

func newReplayTopupService(
	command *replayTopupCommandRepo,
	cardCmd *fakeCardCommandClient,
) service.TopupCommandService {
	saldoQ := &fakeSaldoQueryClient{}
	cardQ := &fakeCardQueryClient{}
	return service.NewTopupCommandService(&service.TopupCommandDeps{
		CardAdapter:            adapter.NewCardAdapter(cardQ, cardCmd),
		TopupQueryRepository:   replayTopupQueryRepo{},
		TopupCommandRepository: command,
		SaldoAdapter:           adapter.NewSaldoAdapter(saldoQ, &fakeSaldoCommandClient{}),
		IdempotencyStore:       newFakeIdempotencyStore(),
		Logger:                 &logger.Logger{Log: zap.NewNop()},
		Observability:          replayTopupObservability{},
	})
}

func TestCreateTopupReplayReturnsOriginalAndSkipsSideEffects(t *testing.T) {
	command := newReplayTopupCommandRepo()
	cardCmd := &fakeCardCommandClient{}
	svc := newReplayTopupService(command, cardCmd)

	req := &requests.CreateTopupRequest{
		CardNumber:     "4111111111111111",
		TopupAmount:    100000,
		TopupMethod:    "gopay",
		IdempotencyKey: "topup:create:user:42:op-1",
	}

	first, err := svc.CreateTopup(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, first)
	require.Equal(t, int32(1), first.TopupID)
	require.Equal(t, "success", first.Status)
	require.Equal(t, 1, cardCmd.updateCalls, "first create must run the card metadata side-effect once")
	require.Len(t, command.updates, 1, "the topup status must be persisted exactly once")

	// Replay with the same key: same response, and no side-effects repeat.
	second, err := svc.CreateTopup(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, second)
	require.Equal(t, first.TopupID, second.TopupID, "replay must return the original record id")
	require.Equal(t, first.TopupAmount, second.TopupAmount)
	require.Equal(t, "success", second.Status)
	require.Equal(t, 1, cardCmd.updateCalls, "replay must not repeat the card metadata side-effect")
	require.Len(t, command.updates, 1, "replay must not repeat the status update")
}

func TestCreateTopupWithoutKeyCreatesIndependently(t *testing.T) {
	command := newReplayTopupCommandRepo()
	cardCmd := &fakeCardCommandClient{}
	svc := newReplayTopupService(command, cardCmd)

	req := &requests.CreateTopupRequest{
		CardNumber:  "4111111111111111",
		TopupAmount: 100000,
		TopupMethod: "gopay",
	}

	first, err := svc.CreateTopup(context.Background(), req)
	require.NoError(t, err)
	second, err := svc.CreateTopup(context.Background(), req)
	require.NoError(t, err)

	require.NotEqual(t, first.TopupID, second.TopupID, "without a key every request creates a new record")
	require.Equal(t, 2, cardCmd.updateCalls)
}
