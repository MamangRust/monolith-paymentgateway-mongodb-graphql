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
	statemachine "github.com/MamangRust/monolith-payment-gateway-shared/domain/status"
	"github.com/MamangRust/monolith-payment-gateway-transaction/repository"
	"github.com/MamangRust/monolith-payment-gateway-transaction/service"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Fake gRPC clients for lifecycle tests
type lifecycleFakeCardQueryClient struct{ pbcard.CardQueryServiceClient }

func (lifecycleFakeCardQueryClient) FindUserCardByCardNumber(context.Context, *pbcard.FindByCardNumberRequest, ...grpc.CallOption) (*pbcard.CardWithEmailResponse, error) {
	return &pbcard.CardWithEmailResponse{CardNumber: "4111111111111111", Email: "user@example.com"}, nil
}
func (lifecycleFakeCardQueryClient) FindByUserIdCard(_ context.Context, _ *pbcard.FindByUserIdCardRequest, _ ...grpc.CallOption) (*pbcard.ApiResponseCard, error) {
	return &pbcard.ApiResponseCard{Data: &pbcard.CardResponse{Id: 1, UserId: 7, CardNumber: "5111111111111111", CardType: "debit"}}, nil
}

type lifecycleFakeCardCommandClient struct {
	pbcard.CardCommandServiceClient
}

func (lifecycleFakeCardCommandClient) UpdateCard(context.Context, *pbcard.UpdateCardRequest, ...grpc.CallOption) (*pbcard.ApiResponseCard, error) {
	return &pbcard.ApiResponseCard{}, nil
}

type lifecycleFakeMerchantQueryClient struct {
	pbmerchant.MerchantQueryServiceClient
}

func (lifecycleFakeMerchantQueryClient) FindByApiKey(context.Context, *pbmerchant.FindByApiKeyRequest, ...grpc.CallOption) (*pbmerchant.ApiResponseMerchant, error) {
	return &pbmerchant.ApiResponseMerchant{Data: &pbmerchant.MerchantResponse{Id: 1, UserId: 7, ApiKey: "key"}}, nil
}

type lifecycleFakeSaldoQueryClient struct {
	pbsaldo.SaldoQueryServiceClient
}

func (lifecycleFakeSaldoQueryClient) FindByCardNumber(_ context.Context, _ *pbcard.FindByCardNumberRequest, _ ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	return &pbsaldo.ApiResponseSaldo{Data: &pbsaldo.SaldoResponse{TotalBalance: 10000000}}, nil
}

type lifecycleFakeSaldoCommandClient struct {
	pbsaldo.SaldoCommandServiceClient
}

func (lifecycleFakeSaldoCommandClient) UpdateSaldoBalance(_ context.Context, req *pbsaldo.UpdateSaldoBalanceRequest, _ ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	return &pbsaldo.ApiResponseSaldo{Data: &pbsaldo.SaldoResponse{CardNumber: req.CardNumber, TotalBalance: req.TotalBalance}}, nil
}
func (lifecycleFakeSaldoCommandClient) DebitSaldo(_ context.Context, req *pbsaldo.DebitSaldoRequest, _ ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	return &pbsaldo.ApiResponseSaldo{Data: &pbsaldo.SaldoResponse{CardNumber: req.CardNumber, TotalBalance: 10000000 - req.Amount}}, nil
}
func (lifecycleFakeSaldoCommandClient) CreditSaldo(_ context.Context, req *pbsaldo.CreditSaldoRequest, _ ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	return &pbsaldo.ApiResponseSaldo{Data: &pbsaldo.SaldoResponse{CardNumber: req.CardNumber, TotalBalance: 10000000 + req.Amount}}, nil
}

// lifecycleQueryRepo returns a transaction in the given state so the lifecycle
// state machine is exercised at the service boundary.
type lifecycleQueryRepo struct {
	currentStatus string
}

func (r *lifecycleQueryRepo) FindById(_ context.Context, id int) (*models.Transaction, error) {
	return &models.Transaction{TransactionID: int32(id), Status: r.currentStatus}, nil
}

// The remaining repository methods are not exercised by the lifecycle unit
// tests. They return well-formed dummy values (never errors) so a future
// service change that starts calling them fails with a clear assertion
// instead of a misleading "not implemented" runtime error.
func (lifecycleQueryRepo) FindAllTransactions(context.Context, *requests.FindAllTransactions) ([]*repository.TransactionQueryResult, error) {
	return []*repository.TransactionQueryResult{}, nil
}
func (lifecycleQueryRepo) FindByActive(context.Context, *requests.FindAllTransactions) ([]*repository.TransactionQueryResult, error) {
	return []*repository.TransactionQueryResult{}, nil
}
func (lifecycleQueryRepo) FindByTrashed(context.Context, *requests.FindAllTransactions) ([]*repository.TransactionQueryResult, error) {
	return []*repository.TransactionQueryResult{}, nil
}
func (lifecycleQueryRepo) FindAllTransactionByCardNumber(context.Context, *requests.FindAllTransactionCardNumber) ([]*repository.TransactionByCardResult, error) {
	return []*repository.TransactionByCardResult{}, nil
}
func (lifecycleQueryRepo) FindTransactionByMerchantId(context.Context, int) ([]*models.Transaction, error) {
	return []*models.Transaction{}, nil
}

// lifecycleCommandRepo counts balance deltas so tests can assert that a
// rejected lifecycle operation never moves money.
type lifecycleCommandRepo struct {
	deltaCalls int
}

func (r *lifecycleCommandRepo) UpdateSaldoBalanceDelta(context.Context, string, int) error {
	r.deltaCalls++
	return nil
}

func (lifecycleCommandRepo) CreateTransactionAtomic(context.Context, models.TransactionAllFieldsRow) (*models.Transaction, error) {
	return &models.Transaction{}, nil
}
func (lifecycleCommandRepo) CreateTransactionAtomicIdempotent(context.Context, models.TransactionAllFieldsRow) (*repository.TransactionAtomicResult, error) {
	return &repository.TransactionAtomicResult{Row: &models.TransactionAllFieldsRow{}}, nil
}
func (lifecycleCommandRepo) CreateTransaction(context.Context, *requests.CreateTransactionRequest) (*models.Transaction, error) {
	return &models.Transaction{}, nil
}
func (lifecycleCommandRepo) UpdateTransaction(context.Context, *requests.UpdateTransactionRequest) (*models.Transaction, error) {
	return &models.Transaction{}, nil
}
func (lifecycleCommandRepo) UpdateTransactionStatus(context.Context, *requests.UpdateTransactionStatus) (*models.Transaction, error) {
	return &models.Transaction{}, nil
}

func (lifecycleCommandRepo) AuthorizeTransactionAtomic(context.Context, models.TransactionAllFieldsRow) (*models.Transaction, error) {
	return &models.Transaction{}, nil
}
func (lifecycleCommandRepo) AuthorizeTransactionAtomicIdempotent(context.Context, models.TransactionAllFieldsRow) (*repository.AuthorizeTransactionResult, error) {
	return &repository.AuthorizeTransactionResult{Row: &models.TransactionAllFieldsRow{}}, nil
}

func (lifecycleCommandRepo) UpdateTransactionAtomic(context.Context, models.TransactionAllFieldsRow) (*models.Transaction, error) {
	return &models.Transaction{}, nil
}

func (lifecycleCommandRepo) CaptureTransactionAtomic(context.Context, models.TransactionAllFieldsRow) (*models.Transaction, error) {
	return &models.Transaction{}, nil
}

func (lifecycleCommandRepo) VoidTransactionAtomic(context.Context, int32) (*models.Transaction, error) {
	return &models.Transaction{}, nil
}

func (lifecycleCommandRepo) RefundTransactionAtomic(context.Context, models.TransactionAllFieldsRow) (*models.Transaction, error) {
	return &models.Transaction{}, nil
}

func (lifecycleCommandRepo) TransitionStatus(_ context.Context, _ int, fromStatus, toStatus, _ string) (*models.Transaction, error) {
	return &models.Transaction{Status: toStatus}, nil
}
func (lifecycleCommandRepo) GuardStatus(_ context.Context, _ int, _, _, _ string) (bool, error) {
	return true, nil
}
func (lifecycleCommandRepo) ListStuck(_ context.Context, _ time.Duration, _ int32) ([]*models.Transaction, error) {
	return []*models.Transaction{}, nil
}
func (lifecycleCommandRepo) TrashedTransaction(context.Context, int) (*models.Transaction, error) {
	return &models.Transaction{}, nil
}
func (lifecycleCommandRepo) RestoreTransaction(context.Context, int) (*models.Transaction, error) {
	return &models.Transaction{}, nil
}
func (lifecycleCommandRepo) DeleteTransactionPermanent(context.Context, int) (bool, error) {
	return true, nil
}
func (lifecycleCommandRepo) RestoreAllTransaction(context.Context) (bool, error) {
	return true, nil
}
func (lifecycleCommandRepo) DeleteAllTransactionPermanent(context.Context) (bool, error) {
	return true, nil
}

func newLifecycleService(command *lifecycleCommandRepo, query *lifecycleQueryRepo) service.TransactionCommandService {
	ca := adapter.NewCardAdapter(&lifecycleFakeCardQueryClient{}, &lifecycleFakeCardCommandClient{})
	ma := adapter.NewMerchantAdapter(&lifecycleFakeMerchantQueryClient{})
	sa := adapter.NewSaldoAdapter(&lifecycleFakeSaldoQueryClient{}, &lifecycleFakeSaldoCommandClient{})
	return service.NewTransactionCommandService(&service.TransactionCommandServiceDeps{
		MerchantAdapter:              ma,
		CardAdapter:                  ca,
		SaldoAdapter:                 sa,
		TransactionQueryRepository:   query,
		TransactionCommandRepository: command,
		Logger:                       &logger.Logger{Log: zap.NewNop()},
		Observability:                replayTransactionObservability{},
	})
}

func TestLifecycleOpsRejectDoubleApplication(t *testing.T) {
	cases := []struct {
		name    string
		current string
		call    func(service.TransactionCommandService) (*models.Transaction, error)
	}{
		{"double-capture", statemachine.Captured, func(s service.TransactionCommandService) (*models.Transaction, error) {
			txn, err := s.CaptureTransaction(context.Background(), "api-key", 7)
			return txn, err
		}},
		{"double-void", statemachine.Voided, func(s service.TransactionCommandService) (*models.Transaction, error) {
			txn, err := s.VoidTransaction(context.Background(), "api-key", 7)
			return txn, err
		}},
		{"double-refund", statemachine.Refunded, func(s service.TransactionCommandService) (*models.Transaction, error) {
			txn, err := s.RefundTransaction(context.Background(), "api-key", 7)
			return txn, err
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			command := &lifecycleCommandRepo{}
			svc := newLifecycleService(command, &lifecycleQueryRepo{currentStatus: tc.current})

			_, err := tc.call(svc)
			require.Error(t, err, "double %s must be rejected", tc.name)
			require.Contains(t, err.Error(), "invalid status transition", "expected a 400 state-machine error")
			require.Zero(t, command.deltaCalls, "double %s must not move any balance", tc.name)
		})
	}
}

func TestLifecycleOpsRejectWrongSourceState(t *testing.T) {
	// A refund directly from "authorized" (without capture) is invalid and must
	// not touch balances.
	command := &lifecycleCommandRepo{}
	svc := newLifecycleService(command, &lifecycleQueryRepo{currentStatus: statemachine.Authorized})

	_, err := svc.RefundTransaction(context.Background(), "api-key", 7)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid status transition")
	require.Zero(t, command.deltaCalls)
}
