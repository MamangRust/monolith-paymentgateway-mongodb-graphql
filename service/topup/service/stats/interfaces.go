package topupstatsservice

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type TopupStatsAmountService interface {
	FindMonthlyTopupAmounts(ctx context.Context, year int) ([]*salemongo.TopupMonthlyAmountRow, error)
	FindYearlyTopupAmounts(ctx context.Context, year int) ([]*salemongo.TopupYearlyAmountRow, error)
}

type TopupStatsMethodService interface {
	FindMonthlyTopupMethods(ctx context.Context, year int) ([]*salemongo.TopupMonthlyMethodRow, error)
	FindYearlyTopupMethods(ctx context.Context, year int) ([]*salemongo.TopupYearlyMethodRow, error)
}

type TopupStatsStatusService interface {
	FindMonthTopupStatusSuccess(ctx context.Context, req *requests.MonthTopupStatus) ([]*salemongo.TopupMonthlyStatusRow, error)
	FindYearlyTopupStatusSuccess(ctx context.Context, year int) ([]*salemongo.TopupYearlyStatusRow, error)
	FindMonthTopupStatusFailed(ctx context.Context, req *requests.MonthTopupStatus) ([]*salemongo.TopupMonthlyStatusRow, error)
	FindYearlyTopupStatusFailed(ctx context.Context, year int) ([]*salemongo.TopupYearlyStatusRow, error)
}
