package transactionstatsrepository

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type TransactionStatsStatusRepository interface {
	GetMonthTransactionStatusSuccess(ctx context.Context, req *requests.MonthStatusTransaction) ([]*salemongo.TransactionMonthlyStatusSuccessRow, error)
	GetYearlyTransactionStatusSuccess(ctx context.Context, year int) ([]*salemongo.TransactionYearlyStatusSuccessRow, error)
	GetMonthTransactionStatusFailed(ctx context.Context, req *requests.MonthStatusTransaction) ([]*salemongo.TransactionMonthlyStatusFailedRow, error)
	GetYearlyTransactionStatusFailed(ctx context.Context, year int) ([]*salemongo.TransactionYearlyStatusFailedRow, error)
}

type TransactionStatsMethodRepository interface {
	GetMonthlyPaymentMethods(ctx context.Context, year int) ([]*salemongo.TransactionMonthlyPaymentMethodRow, error)
	GetYearlyPaymentMethods(ctx context.Context, year int) ([]*salemongo.TransactionYearlyPaymentMethodRow, error)
}

type TransactionStatsAmountRepository interface {
	GetMonthlyAmounts(ctx context.Context, year int) ([]*salemongo.TransactionMonthlyAmountRow, error)
	GetYearlyAmounts(ctx context.Context, year int) ([]*salemongo.TransactionYearlyAmountRow, error)
}
