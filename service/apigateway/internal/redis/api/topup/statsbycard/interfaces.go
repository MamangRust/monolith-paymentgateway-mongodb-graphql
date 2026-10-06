package topup_stats_bycard_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
)

type TopupStatsStatusByCardCache interface {
	GetMonthTopupStatusSuccessByCardNumberCache(ctx context.Context, req *model.FindMonthlyTopupStatusCardNumberInput) (*model.APIResponseTopupMonthStatusSuccess, bool)
	SetMonthTopupStatusSuccessByCardNumberCache(ctx context.Context, req *model.FindMonthlyTopupStatusCardNumberInput, data *model.APIResponseTopupMonthStatusSuccess)

	GetYearlyTopupStatusSuccessByCardNumberCache(ctx context.Context, req *model.FindYearTopupStatusCardNumberInput) (*model.APIResponseTopupYearStatusSuccess, bool)
	SetYearlyTopupStatusSuccessByCardNumberCache(ctx context.Context, req *model.FindYearTopupStatusCardNumberInput, data *model.APIResponseTopupYearStatusSuccess)

	GetMonthTopupStatusFailedByCardNumberCache(ctx context.Context, req *model.FindMonthlyTopupStatusCardNumberInput) (*model.APIResponseTopupMonthStatusFailed, bool)
	SetMonthTopupStatusFailedByCardNumberCache(ctx context.Context, req *model.FindMonthlyTopupStatusCardNumberInput, data *model.APIResponseTopupMonthStatusFailed)

	GetYearlyTopupStatusFailedByCardNumberCache(ctx context.Context, req *model.FindYearTopupStatusCardNumberInput) (*model.APIResponseTopupYearStatusFailed, bool)
	SetYearlyTopupStatusFailedByCardNumberCache(ctx context.Context, req *model.FindYearTopupStatusCardNumberInput, data *model.APIResponseTopupYearStatusFailed)
}

type TopupStatsMethodByCardCache interface {
	GetMonthlyTopupMethodsByCardNumberCache(ctx context.Context, req *model.FindYearTopupCardNumberInput) (*model.APIResponseTopupMonthMethod, bool)
	SetMonthlyTopupMethodsByCardNumberCache(ctx context.Context, req *model.FindYearTopupCardNumberInput, data *model.APIResponseTopupMonthMethod)

	GetYearlyTopupMethodsByCardNumberCache(ctx context.Context, req *model.FindYearTopupCardNumberInput) (*model.APIResponseTopupYearMethod, bool)
	SetYearlyTopupMethodsByCardNumberCache(ctx context.Context, req *model.FindYearTopupCardNumberInput, data *model.APIResponseTopupYearMethod)
}

type TopupStatsAmountByCardCache interface {
	GetMonthlyTopupAmountsByCardNumberCache(ctx context.Context, req *model.FindYearTopupCardNumberInput) (*model.APIResponseTopupMonthAmount, bool)
	SetMonthlyTopupAmountsByCardNumberCache(ctx context.Context, req *model.FindYearTopupCardNumberInput, data *model.APIResponseTopupMonthAmount)

	GetYearlyTopupAmountsByCardNumberCache(ctx context.Context, req *model.FindYearTopupCardNumberInput) (*model.APIResponseTopupYearAmount, bool)
	SetYearlyTopupAmountsByCardNumberCache(ctx context.Context, req *model.FindYearTopupCardNumberInput, data *model.APIResponseTopupYearAmount)
}
