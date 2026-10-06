package transactionbycardrepository

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type TransactionStatsByCardStatusRepository interface {
	GetMonthTransactionStatusSuccessByCardNumber(ctx context.Context, req *requests.MonthStatusTransactionCardNumber) ([]*salemongo.TransactionMonthlyStatusSuccessByCardRow, error)
	GetYearlyTransactionStatusSuccessByCardNumber(ctx context.Context, req *requests.YearStatusTransactionCardNumber) ([]*salemongo.TransactionYearlyStatusSuccessByCardRow, error)
	GetMonthTransactionStatusFailedByCardNumber(ctx context.Context, req *requests.MonthStatusTransactionCardNumber) ([]*salemongo.TransactionMonthlyStatusFailedByCardRow, error)
	GetYearlyTransactionStatusFailedByCardNumber(ctx context.Context, req *requests.YearStatusTransactionCardNumber) ([]*salemongo.TransactionYearlyStatusFailedByCardRow, error)
}

type TransactionStatsByCardMethodRepository interface {
	GetMonthlyPaymentMethodsByCardNumber(ctx context.Context, req *requests.MonthYearPaymentMethod) ([]*salemongo.TransactionMonthlyPaymentMethodByCardRow, error)
	GetYearlyPaymentMethodsByCardNumber(ctx context.Context, req *requests.MonthYearPaymentMethod) ([]*salemongo.TransactionYearlyPaymentMethodByCardRow, error)
}

type TransactionStatsByCardAmountRepository interface {
	GetMonthlyAmountsByCardNumber(ctx context.Context, req *requests.MonthYearPaymentMethod) ([]*salemongo.TransactionMonthlyAmountByCardRow, error)
	GetYearlyAmountsByCardNumber(ctx context.Context, req *requests.MonthYearPaymentMethod) ([]*salemongo.TransactionYearlyAmountByCardRow, error)
}
