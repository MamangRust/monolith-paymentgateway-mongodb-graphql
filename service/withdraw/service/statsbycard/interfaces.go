package withdrawstatsbycardservice

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type WithdrawStatsByCardStatusService interface {
	FindMonthWithdrawStatusSuccessByCardNumber(ctx context.Context, req *requests.MonthStatusWithdrawCardNumber) ([]*salemongo.WithdrawMonthlyStatusSuccessByCardRow, error)
	FindYearlyWithdrawStatusSuccessByCardNumber(ctx context.Context, req *requests.YearStatusWithdrawCardNumber) ([]*salemongo.WithdrawYearlyStatusSuccessByCardRow, error)
	FindMonthWithdrawStatusFailedByCardNumber(ctx context.Context, req *requests.MonthStatusWithdrawCardNumber) ([]*salemongo.WithdrawMonthlyStatusFailedByCardRow, error)
	FindYearlyWithdrawStatusFailedByCardNumber(ctx context.Context, req *requests.YearStatusWithdrawCardNumber) ([]*salemongo.WithdrawYearlyStatusFailedByCardRow, error)
}

type WithdrawStatsByCardAmountService interface {
	FindMonthlyWithdrawsByCardNumber(ctx context.Context, req *requests.YearMonthCardNumber) ([]*salemongo.WithdrawMonthlyAmountByCardRow, error)
	FindYearlyWithdrawsByCardNumber(ctx context.Context, req *requests.YearMonthCardNumber) ([]*salemongo.WithdrawYearlyAmountByCardRow, error)
}
