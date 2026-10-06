package withdrawstatsbycardcache

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type WithdrawStatsByCardStatusCache interface {
	GetCachedMonthWithdrawStatusSuccessByCardNumber(ctx context.Context, req *requests.MonthStatusWithdrawCardNumber) ([]*salemongo.WithdrawMonthlyStatusSuccessByCardRow, bool)
	SetCachedMonthWithdrawStatusSuccessByCardNumber(ctx context.Context, req *requests.MonthStatusWithdrawCardNumber, data []*salemongo.WithdrawMonthlyStatusSuccessByCardRow)

	GetCachedYearlyWithdrawStatusSuccessByCardNumber(ctx context.Context, req *requests.YearStatusWithdrawCardNumber) ([]*salemongo.WithdrawYearlyStatusSuccessByCardRow, bool)
	SetCachedYearlyWithdrawStatusSuccessByCardNumber(ctx context.Context, req *requests.YearStatusWithdrawCardNumber, data []*salemongo.WithdrawYearlyStatusSuccessByCardRow)

	GetCachedMonthWithdrawStatusFailedByCardNumber(ctx context.Context, req *requests.MonthStatusWithdrawCardNumber) ([]*salemongo.WithdrawMonthlyStatusFailedByCardRow, bool)
	SetCachedMonthWithdrawStatusFailedByCardNumber(ctx context.Context, req *requests.MonthStatusWithdrawCardNumber, data []*salemongo.WithdrawMonthlyStatusFailedByCardRow)

	GetCachedYearlyWithdrawStatusFailedByCardNumber(ctx context.Context, req *requests.YearStatusWithdrawCardNumber) ([]*salemongo.WithdrawYearlyStatusFailedByCardRow, bool)
	SetCachedYearlyWithdrawStatusFailedByCardNumber(ctx context.Context, req *requests.YearStatusWithdrawCardNumber, data []*salemongo.WithdrawYearlyStatusFailedByCardRow)
}

type WithdrawStatsByCardAmountCache interface {
	GetCachedMonthlyWithdrawsByCardNumber(ctx context.Context, req *requests.YearMonthCardNumber) ([]*salemongo.WithdrawMonthlyAmountByCardRow, bool)
	SetCachedMonthlyWithdrawsByCardNumber(ctx context.Context, req *requests.YearMonthCardNumber, data []*salemongo.WithdrawMonthlyAmountByCardRow)

	GetCachedYearlyWithdrawsByCardNumber(ctx context.Context, req *requests.YearMonthCardNumber) ([]*salemongo.WithdrawYearlyAmountByCardRow, bool)
	SetCachedYearlyWithdrawsByCardNumber(ctx context.Context, req *requests.YearMonthCardNumber, data []*salemongo.WithdrawYearlyAmountByCardRow)
}
