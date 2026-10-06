package repositorydashboard

import "context"

type CardDashboardBalanceRepository interface {
	GetTotalBalances(ctx context.Context) (*int64, error)
	GetTotalBalanceByCardNumber(ctx context.Context, cardNumber string) (*int64, error)
}

type CardDashboardTopupRepository interface {
	GetTotalTopAmount(ctx context.Context) (*int64, error)
	GetTotalTopupAmountByCardNumber(ctx context.Context, cardNumber string) (*int64, error)
}

type CardDashboardWithdrawRepository interface {
	GetTotalWithdrawAmount(ctx context.Context) (*int64, error)
	GetTotalWithdrawAmountByCardNumber(ctx context.Context, cardNumber string) (*int64, error)
}

type CardDashboardTransactionRepository interface {
	GetTotalTransactionAmount(ctx context.Context) (*int64, error)
	GetTotalTransactionAmountByCardNumber(ctx context.Context, cardNumber string) (*int64, error)
}

type CardDashboardTransferRepository interface {
	GetTotalTransferAmount(ctx context.Context) (*int64, error)
	GetTotalTransferAmountBySender(ctx context.Context, senderCardNumber string) (*int64, error)
	GetTotalTransferAmountByReceiver(ctx context.Context, receiverCardNumber string) (*int64, error)
}
