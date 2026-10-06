package topupstatscache

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type TopupStatsStatusCache interface {
	GetMonthTopupStatusSuccessCache(ctx context.Context, req *requests.MonthTopupStatus) ([]*salemongo.TopupMonthlyStatusRow, bool)
	SetMonthTopupStatusSuccessCache(ctx context.Context, req *requests.MonthTopupStatus, data []*salemongo.TopupMonthlyStatusRow)

	GetYearlyTopupStatusSuccessCache(ctx context.Context, year int) ([]*salemongo.TopupYearlyStatusRow, bool)
	SetYearlyTopupStatusSuccessCache(ctx context.Context, year int, data []*salemongo.TopupYearlyStatusRow)

	GetMonthTopupStatusFailedCache(ctx context.Context, req *requests.MonthTopupStatus) ([]*salemongo.TopupMonthlyStatusRow, bool)
	SetMonthTopupStatusFailedCache(ctx context.Context, req *requests.MonthTopupStatus, data []*salemongo.TopupMonthlyStatusRow)

	GetYearlyTopupStatusFailedCache(ctx context.Context, year int) ([]*salemongo.TopupYearlyStatusRow, bool)
	SetYearlyTopupStatusFailedCache(ctx context.Context, year int, data []*salemongo.TopupYearlyStatusRow)
}

type TopupStatsMethodCache interface {
	GetMonthlyTopupMethodsCache(ctx context.Context, year int) ([]*salemongo.TopupMonthlyMethodRow, bool)
	SetMonthlyTopupMethodsCache(ctx context.Context, year int, data []*salemongo.TopupMonthlyMethodRow)

	GetYearlyTopupMethodsCache(ctx context.Context, year int) ([]*salemongo.TopupYearlyMethodRow, bool)
	SetYearlyTopupMethodsCache(ctx context.Context, year int, data []*salemongo.TopupYearlyMethodRow)
}

type TopupStatsAmountCache interface {
	GetMonthlyTopupAmountsCache(ctx context.Context, year int) ([]*salemongo.TopupMonthlyAmountRow, bool)
	SetMonthlyTopupAmountsCache(ctx context.Context, year int, data []*salemongo.TopupMonthlyAmountRow)

	GetYearlyTopupAmountsCache(ctx context.Context, year int) ([]*salemongo.TopupYearlyAmountRow, bool)
	SetYearlyTopupAmountsCache(ctx context.Context, year int, data []*salemongo.TopupYearlyAmountRow)
}
