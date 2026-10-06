package transactionstatscache

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type TransactionStatsAmountCache interface {
	GetMonthlyAmountsCache(ctx context.Context, year int) ([]*salemongo.TransactionMonthlyAmountRow, bool)
	SetMonthlyAmountsCache(ctx context.Context, year int, data []*salemongo.TransactionMonthlyAmountRow)

	GetYearlyAmountsCache(ctx context.Context, year int) ([]*salemongo.TransactionYearlyAmountRow, bool)
	SetYearlyAmountsCache(ctx context.Context, year int, data []*salemongo.TransactionYearlyAmountRow)
}

type TransactionStatsMethodCache interface {
	GetMonthlyPaymentMethodsCache(ctx context.Context, year int) ([]*salemongo.TransactionMonthlyPaymentMethodRow, bool)
	SetMonthlyPaymentMethodsCache(ctx context.Context, year int, data []*salemongo.TransactionMonthlyPaymentMethodRow)

	GetYearlyPaymentMethodsCache(ctx context.Context, year int) ([]*salemongo.TransactionYearlyPaymentMethodRow, bool)
	SetYearlyPaymentMethodsCache(ctx context.Context, year int, data []*salemongo.TransactionYearlyPaymentMethodRow)
}

type TransactionStatsStatusCache interface {
	GetMonthTransactionStatusSuccessCache(ctx context.Context, req *requests.MonthStatusTransaction) ([]*salemongo.TransactionMonthlyStatusSuccessRow, bool)
	SetMonthTransactionStatusSuccessCache(ctx context.Context, req *requests.MonthStatusTransaction, data []*salemongo.TransactionMonthlyStatusSuccessRow)

	GetYearTransactionStatusSuccessCache(ctx context.Context, year int) ([]*salemongo.TransactionYearlyStatusSuccessRow, bool)
	SetYearTransactionStatusSuccessCache(ctx context.Context, year int, data []*salemongo.TransactionYearlyStatusSuccessRow)

	GetMonthTransactionStatusFailedCache(ctx context.Context, req *requests.MonthStatusTransaction) ([]*salemongo.TransactionMonthlyStatusFailedRow, bool)
	SetMonthTransactionStatusFailedCache(ctx context.Context, req *requests.MonthStatusTransaction, data []*salemongo.TransactionMonthlyStatusFailedRow)

	GetYearTransactionStatusFailedCache(ctx context.Context, year int) ([]*salemongo.TransactionYearlyStatusFailedRow, bool)
	SetYearTransactionStatusFailedCache(ctx context.Context, year int, data []*salemongo.TransactionYearlyStatusFailedRow)
}
