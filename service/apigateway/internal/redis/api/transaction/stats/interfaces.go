package transaction_stats_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type TransactionStatsAmountCache interface {
	GetMonthlyAmountsCache(ctx context.Context, year int) (*model.APIResponseTransactionMonthAmount, bool)
	SetMonthlyAmountsCache(ctx context.Context, year int, data *model.APIResponseTransactionMonthAmount)

	GetYearlyAmountsCache(ctx context.Context, year int) (*model.APIResponseTransactionYearAmount, bool)
	SetYearlyAmountsCache(ctx context.Context, year int, data *model.APIResponseTransactionYearAmount)
}

type TransactionStatsMethodCache interface {
	GetMonthlyPaymentMethodsCache(ctx context.Context, year int) (*model.APIResponseTransactionMonthMethod, bool)
	SetMonthlyPaymentMethodsCache(ctx context.Context, year int, data *model.APIResponseTransactionMonthMethod)

	GetYearlyPaymentMethodsCache(ctx context.Context, year int) (*model.APIResponseTransactionYearMethod, bool)
	SetYearlyPaymentMethodsCache(ctx context.Context, year int, data *model.APIResponseTransactionYearMethod)
}

type TransactionStatsStatusCache interface {
	GetMonthTransactionStatusSuccessCache(ctx context.Context, req *requests.MonthStatusTransaction) (*model.APIResponseTransactionMonthStatusSuccess, bool)
	SetMonthTransactionStatusSuccessCache(ctx context.Context, req *requests.MonthStatusTransaction, data *model.APIResponseTransactionMonthStatusSuccess)

	GetYearTransactionStatusSuccessCache(ctx context.Context, year int) (*model.APIResponseTransactionYearStatusSuccess, bool)
	SetYearTransactionStatusSuccessCache(ctx context.Context, year int, data *model.APIResponseTransactionYearStatusSuccess)

	GetMonthTransactionStatusFailedCache(ctx context.Context, req *requests.MonthStatusTransaction) (*model.APIResponseTransactionMonthStatusFailed, bool)
	SetMonthTransactionStatusFailedCache(ctx context.Context, req *requests.MonthStatusTransaction, data *model.APIResponseTransactionMonthStatusFailed)

	GetYearTransactionStatusFailedCache(ctx context.Context, year int) (*model.APIResponseTransactionYearStatusFailed, bool)
	SetYearTransactionStatusFailedCache(ctx context.Context, year int, data *model.APIResponseTransactionYearStatusFailed)
}
