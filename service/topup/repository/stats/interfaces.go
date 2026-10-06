package topupstatsrepository

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type TopupStatsAmountRepository interface {
	GetMonthlyTopupAmounts(ctx context.Context, year int) ([]*salemongo.TopupMonthlyAmountRow, error)
	GetYearlyTopupAmounts(ctx context.Context, year int) ([]*salemongo.TopupYearlyAmountRow, error)
}

type TopupStatsStatusRepository interface {
	GetMonthTopupStatusSuccess(ctx context.Context, req *requests.MonthTopupStatus) ([]*salemongo.TopupMonthlyStatusRow, error)
	GetYearlyTopupStatusSuccess(ctx context.Context, year int) ([]*salemongo.TopupYearlyStatusRow, error)

	GetMonthTopupStatusFailed(ctx context.Context, req *requests.MonthTopupStatus) ([]*salemongo.TopupMonthlyStatusRow, error)
	GetYearlyTopupStatusFailed(ctx context.Context, year int) ([]*salemongo.TopupYearlyStatusRow, error)
}

type TopupStatsMethodRepository interface {
	GetMonthlyTopupMethods(ctx context.Context, year int) ([]*salemongo.TopupMonthlyMethodRow, error)
	GetYearlyTopupMethods(ctx context.Context, year int) ([]*salemongo.TopupYearlyMethodRow, error)
}
