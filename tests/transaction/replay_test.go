package transaction_test

import (
	"context"
	"testing"
	"time"

	pbcard "github.com/MamangRust/monolith-payment-gateway-pb/card"
	pbmerchant "github.com/MamangRust/monolith-payment-gateway-pb/merchant"
	pbsaldo "github.com/MamangRust/monolith-payment-gateway-pb/saldo"
	adapter "github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	models "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/idempotency"
	"github.com/MamangRust/monolith-payment-gateway-transaction/repository"
	"github.com/MamangRust/monolith-payment-gateway-transaction/service"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type replayTransactionMerchantRepo struct{}

func (replayTransactionMerchantRepo) FindByApiKey(context.Context, string) (*models.MerchantAllFieldsRow, error) {
	return &models.MerchantAllFieldsRow{MerchantID: 1, UserID: 7, ApiKey: "key"}, nil
}

type replayTransactionCardRepo struct{}

func (replayTransactionCardRepo) FindUserCardByCardNumber(context.Context, string) (*models.CardByEmailRow, error) {
	return &models.CardByEmailRow{CardNumber: "4111111111111111", Email: "user@example.com"}, nil
}

func (replayTransactionCardRepo) FindCardByUserId(context.Context, int) (*models.CardAllFieldsRow, error) {
	return &models.CardAllFieldsRow{CardNumber: "5111111111111111"}, nil
}

func (replayTransactionCardRepo) FindCardByCardNumber(context.Context, string) (*models.CardAllFieldsRow, error) {
	return &models.CardAllFieldsRow{}, nil
}

func (replayTransactionCardRepo) UpdateCard(context.Context, *requests.UpdateCardRequest) (*models.CardUpdateRow, error) {
	return &models.CardUpdateRow{}, nil
}

func (replayTransactionCardRepo) UpdateCardOutstandingBalance(context.Context, int, int) (*models.UpdateOutstandingBalanceRow, error) {
	return &models.UpdateOutstandingBalanceRow{}, nil
}

func (replayTransactionCardRepo) AddRewardPoints(context.Context, int, int) (*models.AddRewardPointsRow, error) {
	return &models.AddRewardPointsRow{}, nil
}

type replayTransactionCommandRepo struct {
	seq     int32
	created map[string]*models.TransactionAllFieldsRow
}

func (r *replayTransactionCommandRepo) CreateTransactionAtomicIdempotent(_ context.Context, params models.TransactionAllFieldsRow) (*repository.TransactionAtomicResult, error) {
	if params.IdempotencyKey != "" {
		if existing, ok := r.created[params.IdempotencyKey]; ok {
			return &repository.TransactionAtomicResult{Row: existing, Replayed: true}, nil
		}
	}
	res := &models.TransactionAllFieldsRow{
		TransactionID:   r.seq + 1,
		CardNumber:      params.CardNumber,
		Amount:          params.Amount,
		PaymentMethod:   params.PaymentMethod,
		MerchantID:      params.MerchantID,
		TransactionTime: params.TransactionTime,
		Status:          "success",
	}
	r.seq++
	if params.IdempotencyKey != "" {
		r.created[params.IdempotencyKey] = res
	}
	return &repository.TransactionAtomicResult{Row: res}, nil
}

func (replayTransactionCommandRepo) UpdateSaldoBalanceDelta(context.Context, string, int) error {
	return nil
}

func (r *replayTransactionCommandRepo) CreateTransaction(_ context.Context, req *requests.CreateTransactionRequest) (*models.Transaction, error) {
	r.seq++
	return &models.Transaction{TransactionID: r.seq, CardNumber: req.CardNumber, Amount: int64(req.Amount), PaymentMethod: req.PaymentMethod, MerchantID: int32(*req.MerchantID), TransactionTime: req.TransactionTime, Status: "pending"}, nil
}

func (replayTransactionCommandRepo) UpdateTransaction(context.Context, *requests.UpdateTransactionRequest) (*models.Transaction, error) {
	return &models.Transaction{}, nil
}

func (replayTransactionCommandRepo) UpdateTransactionStatus(context.Context, *requests.UpdateTransactionStatus) (*models.Transaction, error) {
	return &models.Transaction{}, nil
}

func (replayTransactionCommandRepo) TransitionStatus(_ context.Context, id int, _, toStatus, _ string) (*models.Transaction, error) {
	return &models.Transaction{TransactionID: int32(id), Status: toStatus}, nil
}
func (replayTransactionCommandRepo) GuardStatus(_ context.Context, _ int, _, _, _ string) (bool, error) {
	return true, nil
}
func (replayTransactionCommandRepo) ListStuck(_ context.Context, _ time.Duration, _ int32) ([]*models.Transaction, error) {
	return []*models.Transaction{}, nil
}

func (replayTransactionCommandRepo) TrashedTransaction(_ context.Context, _ int) (*models.Transaction, error) {
	return &models.Transaction{}, nil
}

func (replayTransactionCommandRepo) RestoreTransaction(_ context.Context, _ int) (*models.Transaction, error) {
	return &models.Transaction{}, nil
}

func (replayTransactionCommandRepo) DeleteTransactionPermanent(context.Context, int) (bool, error) {
	return true, nil
}

func (replayTransactionCommandRepo) RestoreAllTransaction(context.Context) (bool, error) {
	return true, nil
}

func (replayTransactionCommandRepo) DeleteAllTransactionPermanent(context.Context) (bool, error) {
	return true, nil
}

type replayTransactionQueryRepo struct{}

func (replayTransactionQueryRepo) FindAllTransactions(context.Context, *requests.FindAllTransactions) ([]*repository.TransactionQueryResult, error) {
	return []*repository.TransactionQueryResult{}, nil
}

func (replayTransactionQueryRepo) FindByActive(context.Context, *requests.FindAllTransactions) ([]*repository.TransactionQueryResult, error) {
	return []*repository.TransactionQueryResult{}, nil
}

func (replayTransactionQueryRepo) FindByTrashed(context.Context, *requests.FindAllTransactions) ([]*repository.TransactionQueryResult, error) {
	return []*repository.TransactionQueryResult{}, nil
}

func (replayTransactionQueryRepo) FindAllTransactionByCardNumber(context.Context, *requests.FindAllTransactionCardNumber) ([]*repository.TransactionByCardResult, error) {
	return []*repository.TransactionByCardResult{}, nil
}

func (replayTransactionQueryRepo) FindById(context.Context, int) (*models.Transaction, error) {
	return &models.Transaction{}, nil
}

func (replayTransactionQueryRepo) FindTransactionByMerchantId(context.Context, int) ([]*models.Transaction, error) {
	return []*models.Transaction{}, nil
}

type replayTransactionSaldoRepo struct{}

func (replayTransactionSaldoRepo) FindByCardNumber(context.Context, string) (*models.Saldo, error) {
	return &models.Saldo{}, nil
}

func (replayTransactionSaldoRepo) UpdateSaldoBalance(context.Context, *requests.UpdateSaldoBalance) (*models.UpdateSaldoBalanceRow, error) {
	return &models.UpdateSaldoBalanceRow{}, nil
}

type replayTransactionObservability struct{}

func (replayTransactionObservability) StartTracingAndLogging(ctx context.Context, _ string, _ ...attribute.KeyValue) (context.Context, trace.Span, func(string, ...string), string, func(string, ...zap.Field)) {
	ctx, span := trace.NewNoopTracerProvider().Tracer("transaction-replay").Start(ctx, "transaction-replay")
	return ctx, span, func(string, ...string) { span.End() }, "success", func(string, ...zap.Field) {}
}
func (replayTransactionObservability) RecordMetrics(context.Context, string, string, time.Time) {}

type replayTransactionMerchantQueryClient struct {
	pbmerchant.MerchantQueryServiceClient
}

func (replayTransactionMerchantQueryClient) FindByApiKey(_ context.Context, _ *pbmerchant.FindByApiKeyRequest, _ ...grpc.CallOption) (*pbmerchant.ApiResponseMerchant, error) {
	return &pbmerchant.ApiResponseMerchant{Data: &pbmerchant.MerchantResponse{Id: 1, UserId: 7, ApiKey: "key"}}, nil
}

type replayTransactionCardQueryClient struct{ pbcard.CardQueryServiceClient }

func (replayTransactionCardQueryClient) FindUserCardByCardNumber(_ context.Context, _ *pbcard.FindByCardNumberRequest, _ ...grpc.CallOption) (*pbcard.CardWithEmailResponse, error) {
	return &pbcard.CardWithEmailResponse{CardNumber: "4111111111111111", Email: "user@example.com"}, nil
}
func (replayTransactionCardQueryClient) FindByUserIdCard(_ context.Context, _ *pbcard.FindByUserIdCardRequest, _ ...grpc.CallOption) (*pbcard.ApiResponseCard, error) {
	return &pbcard.ApiResponseCard{Data: &pbcard.CardResponse{Id: 1, UserId: 7, CardNumber: "5111111111111111", CardType: "debit"}}, nil
}

type replayTransactionCardCommandClient struct {
	pbcard.CardCommandServiceClient
}

type replayTransactionSaldoQueryClient struct {
	pbsaldo.SaldoQueryServiceClient
}

func (replayTransactionSaldoQueryClient) FindByCardNumber(_ context.Context, _ *pbcard.FindByCardNumberRequest, _ ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	return &pbsaldo.ApiResponseSaldo{Data: &pbsaldo.SaldoResponse{TotalBalance: 10000000}}, nil
}

type replayTransactionSaldoCommandClient struct {
	pbsaldo.SaldoCommandServiceClient
}

func (replayTransactionSaldoCommandClient) UpdateSaldoBalance(_ context.Context, req *pbsaldo.UpdateSaldoBalanceRequest, _ ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	return &pbsaldo.ApiResponseSaldo{Data: &pbsaldo.SaldoResponse{CardNumber: req.CardNumber, TotalBalance: req.TotalBalance}}, nil
}
func (replayTransactionSaldoCommandClient) DebitSaldo(_ context.Context, req *pbsaldo.DebitSaldoRequest, _ ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	return &pbsaldo.ApiResponseSaldo{Data: &pbsaldo.SaldoResponse{CardNumber: req.CardNumber, TotalBalance: 10000000 - req.Amount}}, nil
}
func (replayTransactionSaldoCommandClient) CreditSaldo(_ context.Context, req *pbsaldo.CreditSaldoRequest, _ ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	return &pbsaldo.ApiResponseSaldo{Data: &pbsaldo.SaldoResponse{CardNumber: req.CardNumber, TotalBalance: 10000000 + req.Amount}}, nil
}

type replayIdempotencyStore struct {
	records map[string]*idempotency.Record
}

func newFakeIdempotencyStore() *replayIdempotencyStore {
	return &replayIdempotencyStore{records: map[string]*idempotency.Record{}}
}

func (s *replayIdempotencyStore) Claim(_ context.Context, scope, key, requestHash string) (*idempotency.Record, error) {
	id := scope + ":" + key
	if _, ok := s.records[id]; ok {
		return nil, idempotency.ErrKeyInUse
	}
	rec := &idempotency.Record{Key: key, RequestHash: requestHash, Status: idempotency.StatusProcessing}
	s.records[id] = rec
	return rec, nil
}
func (s *replayIdempotencyStore) Get(_ context.Context, scope, key string) (*idempotency.Record, error) {
	return s.records[scope+":"+key], nil
}
func (s *replayIdempotencyStore) Complete(_ context.Context, scope, key, requestHash string, outcome idempotency.Outcome) error {
	rec := s.records[scope+":"+key]
	if rec == nil {
		return nil
	}
	rec.Status = outcome.Status
	rec.ResponseJSON = outcome.ResponseJSON
	rec.RequestHash = requestHash
	return nil
}
func (s *replayIdempotencyStore) Release(_ context.Context, scope, key, _ string) error {
	delete(s.records, scope+":"+key)
	return nil
}

func newReplayTransactionService(command *replayTransactionCommandRepo) service.TransactionCommandService {
	ca := adapter.NewCardAdapter(&replayTransactionCardQueryClient{}, &replayTransactionCardCommandClient{})
	sa := adapter.NewSaldoAdapter(&replayTransactionSaldoQueryClient{}, &replayTransactionSaldoCommandClient{})
	return service.NewTransactionCommandService(&service.TransactionCommandServiceDeps{
		MerchantAdapter:              adapter.NewMerchantAdapter(&replayTransactionMerchantQueryClient{}),
		CardAdapter:                  ca,
		SaldoAdapter:                 sa,
		TransactionQueryRepository:   replayTransactionQueryRepo{},
		TransactionCommandRepository: command,
		IdempotencyStore:             newFakeIdempotencyStore(),
		Logger:                       &logger.Logger{Log: zap.NewNop()},
		Observability:                replayTransactionObservability{},
	})
}

func TestCreateTransactionReplayReturnsOriginal(t *testing.T) {
	command := &replayTransactionCommandRepo{created: map[string]*models.TransactionAllFieldsRow{}}
	svc := newReplayTransactionService(command)

	req := &requests.CreateTransactionRequest{
		CardNumber:      "4111111111111111",
		Amount:          100000,
		PaymentMethod:   "gopay",
		TransactionTime: time.Now(),
		IdempotencyKey:  "transaction:create:merchant:1:op-1",
	}

	first, err := svc.Create(context.Background(), "api-key", req)
	require.NoError(t, err)
	require.NotNil(t, first)
	require.Equal(t, int32(1), first.TransactionID)
	require.Equal(t, "success", first.Status)

	// Create a fresh request: the service mutates req.MerchantID in-place,
	// so using the same object would change the idempotency hash.
	secondReq := &requests.CreateTransactionRequest{
		CardNumber:      "4111111111111111",
		Amount:          100000,
		PaymentMethod:   "gopay",
		TransactionTime: req.TransactionTime,
		IdempotencyKey:  "transaction:create:merchant:1:op-1",
	}
	second, err := svc.Create(context.Background(), "api-key", secondReq)
	require.NoError(t, err)
	require.NotNil(t, second)
	require.Equal(t, first.TransactionID, second.TransactionID, "replay must return the original transaction id")
	require.Equal(t, first.Amount, second.Amount)
}

func TestCreateTransactionWithoutKeyCreatesTwice(t *testing.T) {
	command := &replayTransactionCommandRepo{created: map[string]*models.TransactionAllFieldsRow{}}
	svc := newReplayTransactionService(command)

	req := &requests.CreateTransactionRequest{
		CardNumber:      "4111111111111111",
		Amount:          100000,
		PaymentMethod:   "gopay",
		TransactionTime: time.Now(),
	}

	first, err := svc.Create(context.Background(), "api-key", req)
	require.NoError(t, err)
	second, err := svc.Create(context.Background(), "api-key", req)
	require.NoError(t, err)

	require.NotEqual(t, first.TransactionID, second.TransactionID, "without a key every request creates a new transaction")
}
