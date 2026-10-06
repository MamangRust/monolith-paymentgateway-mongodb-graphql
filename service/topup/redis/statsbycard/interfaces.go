package topupstatsbycardcache

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type TopupStatsStatusByCardCache interface {
	GetMonthTopupStatusSuccessByCardNumberCache(ctx context.Context, req *requests.MonthTopupStatusCardNumber) ([]*salemongo.TopupMonthlyStatusRow, bool)
	SetMonthTopupStatusSuccessByCardNumberCache(ctx context.Context, req *requests.MonthTopupStatusCardNumber, data []*salemongo.TopupMonthlyStatusRow)

	GetYearlyTopupStatusSuccessByCardNumberCache(ctx context.Context, req *requests.YearTopupStatusCardNumber) ([]*salemongo.TopupYearlyStatusRow, bool)
	SetYearlyTopupStatusSuccessByCardNumberCache(ctx context.Context, req *requests.YearTopupStatusCardNumber, data []*salemongo.TopupYearlyStatusRow)

	GetMonthTopupStatusFailedByCardNumberCache(ctx context.Context, req *requests.MonthTopupStatusCardNumber) ([]*salemongo.TopupMonthlyStatusRow, bool)
	SetMonthTopupStatusFailedByCardNumberCache(ctx context.Context, req *requests.MonthTopupStatusCardNumber, data []*salemongo.TopupMonthlyStatusRow)

	GetYearlyTopupStatusFailedByCardNumberCache(ctx context.Context, req *requests.YearTopupStatusCardNumber) ([]*salemongo.TopupYearlyStatusRow, bool)
	SetYearlyTopupStatusFailedByCardNumberCache(ctx context.Context, req *requests.YearTopupStatusCardNumber, data []*salemongo.TopupYearlyStatusRow)
}

type TopupStatsMethodByCardCache interface {
	GetMonthlyTopupMethodsByCardNumberCache(ctx context.Context, req *requests.YearMonthMethod) ([]*salemongo.TopupMonthlyMethodByCardRow, bool)
	SetMonthlyTopupMethodsByCardNumberCache(ctx context.Context, req *requests.YearMonthMethod, data []*salemongo.TopupMonthlyMethodByCardRow)

	GetYearlyTopupMethodsByCardNumberCache(ctx context.Context, req *requests.YearMonthMethod) ([]*salemongo.TopupYearlyMethodByCardRow, bool)
	SetYearlyTopupMethodsByCardNumberCache(ctx context.Context, req *requests.YearMonthMethod, data []*salemongo.TopupYearlyMethodByCardRow)
}

type TopupStatsAmountByCardCache interface {
	GetMonthlyTopupAmountsByCardNumberCache(ctx context.Context, req *requests.YearMonthMethod) ([]*salemongo.TopupMonthlyAmountRow, bool)
	SetMonthlyTopupAmountsByCardNumberCache(ctx context.Context, req *requests.YearMonthMethod, data []*salemongo.TopupMonthlyAmountRow)

	GetYearlyTopupAmountsByCardNumberCache(ctx context.Context, req *requests.YearMonthMethod) ([]*salemongo.TopupYearlyAmountRow, bool)
	SetYearlyTopupAmountsByCardNumberCache(ctx context.Context, req *requests.YearMonthMethod, data []*salemongo.TopupYearlyAmountRow)
}
