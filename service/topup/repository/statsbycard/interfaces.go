package topupstatsbycardrepository

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type TopupStatsByCardAmountRepository interface {
	GetMonthlyTopupAmountsByCardNumber(ctx context.Context, req *requests.YearMonthMethod) ([]*salemongo.TopupMonthlyAmountRow, error)
	GetYearlyTopupAmountsByCardNumber(ctx context.Context, req *requests.YearMonthMethod) ([]*salemongo.TopupYearlyAmountRow, error)
}

type TopupStatsByCardStatusRepository interface {
	GetMonthTopupStatusSuccessByCardNumber(ctx context.Context, req *requests.MonthTopupStatusCardNumber) ([]*salemongo.TopupMonthlyStatusRow, error)
	GetYearlyTopupStatusSuccessByCardNumber(ctx context.Context, req *requests.YearTopupStatusCardNumber) ([]*salemongo.TopupYearlyStatusRow, error)

	GetMonthTopupStatusFailedByCardNumber(ctx context.Context, req *requests.MonthTopupStatusCardNumber) ([]*salemongo.TopupMonthlyStatusRow, error)
	GetYearlyTopupStatusFailedByCardNumber(ctx context.Context, req *requests.YearTopupStatusCardNumber) ([]*salemongo.TopupYearlyStatusRow, error)
}

type TopupStatsByCardMethodRepository interface {
	GetMonthlyTopupMethodsByCardNumber(ctx context.Context, req *requests.YearMonthMethod) ([]*salemongo.TopupMonthlyMethodByCardRow, error)
	GetYearlyTopupMethodsByCardNumber(ctx context.Context, req *requests.YearMonthMethod) ([]*salemongo.TopupYearlyMethodByCardRow, error)
}
