package transactionstatsbycarcache

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type TransactionStatsByCardAmountCache interface {
	GetMonthlyAmountsByCardCache(ctx context.Context, req *requests.MonthYearPaymentMethod) ([]*salemongo.TransactionMonthlyAmountByCardRow, bool)
	SetMonthlyAmountsByCardCache(ctx context.Context, req *requests.MonthYearPaymentMethod, data []*salemongo.TransactionMonthlyAmountByCardRow)

	GetYearlyAmountsByCardCache(ctx context.Context, req *requests.MonthYearPaymentMethod) ([]*salemongo.TransactionYearlyAmountByCardRow, bool)
	SetYearlyAmountsByCardCache(ctx context.Context, req *requests.MonthYearPaymentMethod, data []*salemongo.TransactionYearlyAmountByCardRow)
}

type TransactionStatsByCardMethodCache interface {
	GetMonthlyPaymentMethodsByCardCache(ctx context.Context, req *requests.MonthYearPaymentMethod) ([]*salemongo.TransactionMonthlyPaymentMethodByCardRow, bool)
	SetMonthlyPaymentMethodsByCardCache(ctx context.Context, req *requests.MonthYearPaymentMethod, data []*salemongo.TransactionMonthlyPaymentMethodByCardRow)

	GetYearlyPaymentMethodsByCardCache(ctx context.Context, req *requests.MonthYearPaymentMethod) ([]*salemongo.TransactionYearlyPaymentMethodByCardRow, bool)
	SetYearlyPaymentMethodsByCardCache(ctx context.Context, req *requests.MonthYearPaymentMethod, data []*salemongo.TransactionYearlyPaymentMethodByCardRow)
}

type TransactionStatsByCardStatusCache interface {
	GetMonthTransactionStatusSuccessByCardCache(ctx context.Context, req *requests.MonthStatusTransactionCardNumber) ([]*salemongo.TransactionMonthlyStatusSuccessByCardRow, bool)
	SetMonthTransactionStatusSuccessByCardCache(ctx context.Context, req *requests.MonthStatusTransactionCardNumber, data []*salemongo.TransactionMonthlyStatusSuccessByCardRow)

	GetYearTransactionStatusSuccessByCardCache(ctx context.Context, req *requests.YearStatusTransactionCardNumber) ([]*salemongo.TransactionYearlyStatusSuccessByCardRow, bool)
	SetYearTransactionStatusSuccessByCardCache(ctx context.Context, req *requests.YearStatusTransactionCardNumber, data []*salemongo.TransactionYearlyStatusSuccessByCardRow)

	GetMonthTransactionStatusFailedByCardCache(ctx context.Context, req *requests.MonthStatusTransactionCardNumber) ([]*salemongo.TransactionMonthlyStatusFailedByCardRow, bool)
	SetMonthTransactionStatusFailedByCardCache(ctx context.Context, req *requests.MonthStatusTransactionCardNumber, data []*salemongo.TransactionMonthlyStatusFailedByCardRow)

	GetYearTransactionStatusFailedByCardCache(ctx context.Context, req *requests.YearStatusTransactionCardNumber) ([]*salemongo.TransactionYearlyStatusFailedByCardRow, bool)
	SetYearTransactionStatusFailedByCardCache(ctx context.Context, req *requests.YearStatusTransactionCardNumber, data []*salemongo.TransactionYearlyStatusFailedByCardRow)
}
