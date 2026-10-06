package transactionstatsbycardservice

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type TransactionStatsByCardAmountService interface {
	FindMonthlyAmountsByCardNumber(ctx context.Context, req *requests.MonthYearPaymentMethod) ([]*salemongo.TransactionMonthlyAmountByCardRow, error)
	FindYearlyAmountsByCardNumber(ctx context.Context, req *requests.MonthYearPaymentMethod) ([]*salemongo.TransactionYearlyAmountByCardRow, error)
}

type TransactionStatsByCardMethodService interface {
	FindMonthlyPaymentMethodsByCardNumber(ctx context.Context, req *requests.MonthYearPaymentMethod) ([]*salemongo.TransactionMonthlyPaymentMethodByCardRow, error)
	FindYearlyPaymentMethodsByCardNumber(ctx context.Context, req *requests.MonthYearPaymentMethod) ([]*salemongo.TransactionYearlyPaymentMethodByCardRow, error)
}

type TransactionStatsByCardStatusService interface {
	FindMonthTransactionStatusSuccessByCardNumber(ctx context.Context, req *requests.MonthStatusTransactionCardNumber) ([]*salemongo.TransactionMonthlyStatusSuccessByCardRow, error)
	FindYearlyTransactionStatusSuccessByCardNumber(ctx context.Context, req *requests.YearStatusTransactionCardNumber) ([]*salemongo.TransactionYearlyStatusSuccessByCardRow, error)
	FindMonthTransactionStatusFailedByCardNumber(ctx context.Context, req *requests.MonthStatusTransactionCardNumber) ([]*salemongo.TransactionMonthlyStatusFailedByCardRow, error)
	FindYearlyTransactionStatusFailedByCardNumber(ctx context.Context, req *requests.YearStatusTransactionCardNumber) ([]*salemongo.TransactionYearlyStatusFailedByCardRow, error)
}
