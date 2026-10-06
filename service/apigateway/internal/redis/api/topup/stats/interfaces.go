package topup_stats_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type TopupStatsStatusCache interface {
	GetMonthTopupStatusSuccessCache(ctx context.Context, req *requests.MonthTopupStatus) (*model.APIResponseTopupMonthStatusSuccess, bool)
	SetMonthTopupStatusSuccessCache(ctx context.Context, req *requests.MonthTopupStatus, data *model.APIResponseTopupMonthStatusSuccess)

	GetYearlyTopupStatusSuccessCache(ctx context.Context, year int) (*model.APIResponseTopupYearStatusSuccess, bool)
	SetYearlyTopupStatusSuccessCache(ctx context.Context, year int, data *model.APIResponseTopupYearStatusSuccess)

	GetMonthTopupStatusFailedCache(ctx context.Context, req *requests.MonthTopupStatus) (*model.APIResponseTopupMonthStatusFailed, bool)
	SetMonthTopupStatusFailedCache(ctx context.Context, req *requests.MonthTopupStatus, data *model.APIResponseTopupMonthStatusFailed)

	GetYearlyTopupStatusFailedCache(ctx context.Context, year int) (*model.APIResponseTopupYearStatusFailed, bool)
	SetYearlyTopupStatusFailedCache(ctx context.Context, year int, data *model.APIResponseTopupYearStatusFailed)
}

type TopupStatsMethodCache interface {
	GetMonthlyTopupMethodsCache(ctx context.Context, year int) (*model.APIResponseTopupMonthMethod, bool)
	SetMonthlyTopupMethodsCache(ctx context.Context, year int, data *model.APIResponseTopupMonthMethod)

	GetYearlyTopupMethodsCache(ctx context.Context, year int) (*model.APIResponseTopupYearMethod, bool)
	SetYearlyTopupMethodsCache(ctx context.Context, year int, data *model.APIResponseTopupYearMethod)
}

type TopupStatsAmountCache interface {
	GetMonthlyTopupAmountsCache(ctx context.Context, year int) (*model.APIResponseTopupMonthAmount, bool)
	SetMonthlyTopupAmountsCache(ctx context.Context, year int, data *model.APIResponseTopupMonthAmount)

	GetYearlyTopupAmountsCache(ctx context.Context, year int) (*model.APIResponseTopupYearAmount, bool)
	SetYearlyTopupAmountsCache(ctx context.Context, year int, data *model.APIResponseTopupYearAmount)
}
