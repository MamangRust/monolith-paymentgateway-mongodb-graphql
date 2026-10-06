package transactionstatsservice

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type TransactionStatsAmountService interface {
	FindMonthlyAmounts(ctx context.Context, year int) ([]*salemongo.TransactionMonthlyAmountRow, error)
	FindYearlyAmounts(ctx context.Context, year int) ([]*salemongo.TransactionYearlyAmountRow, error)
}

type TransactionStatsMethodService interface {
	FindMonthlyPaymentMethods(ctx context.Context, year int) ([]*salemongo.TransactionMonthlyPaymentMethodRow, error)
	FindYearlyPaymentMethods(ctx context.Context, year int) ([]*salemongo.TransactionYearlyPaymentMethodRow, error)
}

type TransactionStatsStatusService interface {
	FindMonthTransactionStatusSuccess(ctx context.Context, req *requests.MonthStatusTransaction) ([]*salemongo.TransactionMonthlyStatusSuccessRow, error)
	FindYearlyTransactionStatusSuccess(ctx context.Context, year int) ([]*salemongo.TransactionYearlyStatusSuccessRow, error)
	FindMonthTransactionStatusFailed(ctx context.Context, req *requests.MonthStatusTransaction) ([]*salemongo.TransactionMonthlyStatusFailedRow, error)
	FindYearlyTransactionStatusFailed(ctx context.Context, year int) ([]*salemongo.TransactionYearlyStatusFailedRow, error)
}
