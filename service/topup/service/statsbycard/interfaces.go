package topupstatsbycardservice

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type TopupStatsByCardAmountService interface {
	FindMonthlyTopupAmountsByCardNumber(ctx context.Context, req *requests.YearMonthMethod) ([]*salemongo.TopupMonthlyAmountRow, error)
	FindYearlyTopupAmountsByCardNumber(ctx context.Context, req *requests.YearMonthMethod) ([]*salemongo.TopupYearlyAmountRow, error)
}

type TopupStatsByCardMethodService interface {
	FindMonthlyTopupMethodsByCardNumber(ctx context.Context, req *requests.YearMonthMethod) ([]*salemongo.TopupMonthlyMethodByCardRow, error)
	FindYearlyTopupMethodsByCardNumber(ctx context.Context, req *requests.YearMonthMethod) ([]*salemongo.TopupYearlyMethodByCardRow, error)
}

type TopupStatsByCardStatusService interface {
	FindMonthTopupStatusSuccessByCardNumber(ctx context.Context, req *requests.MonthTopupStatusCardNumber) ([]*salemongo.TopupMonthlyStatusRow, error)
	FindYearlyTopupStatusSuccessByCardNumber(ctx context.Context, req *requests.YearTopupStatusCardNumber) ([]*salemongo.TopupYearlyStatusRow, error)
	FindMonthTopupStatusFailedByCardNumber(ctx context.Context, req *requests.MonthTopupStatusCardNumber) ([]*salemongo.TopupMonthlyStatusRow, error)
	FindYearlyTopupStatusFailedByCardNumber(ctx context.Context, req *requests.YearTopupStatusCardNumber) ([]*salemongo.TopupYearlyStatusRow, error)
}
