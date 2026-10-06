package transaction_stats_bycard_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
)

type TransactionStatsByCardAmountCache interface {
	GetMonthlyAmountsByCardCache(ctx context.Context, req *model.FindByYearCardNumberTransactionInput) (*model.APIResponseTransactionMonthAmount, bool)
	SetMonthlyAmountsByCardCache(ctx context.Context, req *model.FindByYearCardNumberTransactionInput, data *model.APIResponseTransactionMonthAmount)

	GetYearlyAmountsByCardCache(ctx context.Context, req *model.FindByYearCardNumberTransactionInput) (*model.APIResponseTransactionYearAmount, bool)
	SetYearlyAmountsByCardCache(ctx context.Context, req *model.FindByYearCardNumberTransactionInput, data *model.APIResponseTransactionYearAmount)
}

type TransactionStatsByCardMethodCache interface {
	GetMonthlyPaymentMethodsByCardCache(ctx context.Context, req *model.FindByYearCardNumberTransactionInput) (*model.APIResponseTransactionMonthMethod, bool)
	SetMonthlyPaymentMethodsByCardCache(ctx context.Context, req *model.FindByYearCardNumberTransactionInput, data *model.APIResponseTransactionMonthMethod)

	GetYearlyPaymentMethodsByCardCache(ctx context.Context, req *model.FindByYearCardNumberTransactionInput) (*model.APIResponseTransactionYearMethod, bool)
	SetYearlyPaymentMethodsByCardCache(ctx context.Context, req *model.FindByYearCardNumberTransactionInput, data *model.APIResponseTransactionYearMethod)
}

type TransactionStatsByCardStatusCache interface {
	GetMonthTransactionStatusSuccessByCardCache(ctx context.Context, req *model.FindMonthlyTransactionStatusCardNumberInput) (*model.APIResponseTransactionMonthStatusSuccess, bool)
	SetMonthTransactionStatusSuccessByCardCache(ctx context.Context, req *model.FindMonthlyTransactionStatusCardNumberInput, data *model.APIResponseTransactionMonthStatusSuccess)

	GetYearTransactionStatusSuccessByCardCache(ctx context.Context, req *model.FindYearTransactionStatusCardNumberInput) (*model.APIResponseTransactionYearStatusSuccess, bool)
	SetYearTransactionStatusSuccessByCardCache(ctx context.Context, req *model.FindYearTransactionStatusCardNumberInput, data *model.APIResponseTransactionYearStatusSuccess)

	GetMonthTransactionStatusFailedByCardCache(ctx context.Context, req *model.FindMonthlyTransactionStatusCardNumberInput) (*model.APIResponseTransactionMonthStatusFailed, bool)
	SetMonthTransactionStatusFailedByCardCache(ctx context.Context, req *model.FindMonthlyTransactionStatusCardNumberInput, data *model.APIResponseTransactionMonthStatusFailed)

	GetYearTransactionStatusFailedByCardCache(ctx context.Context, req *model.FindYearTransactionStatusCardNumberInput) (*model.APIResponseTransactionYearStatusFailed, bool)
	SetYearTransactionStatusFailedByCardCache(ctx context.Context, req *model.FindYearTransactionStatusCardNumberInput, data *model.APIResponseTransactionYearStatusFailed)
}
