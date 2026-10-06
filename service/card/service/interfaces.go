package service

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/response"
)

type CardAuthorizationService interface {
	Authorize(ctx context.Context, request *requests.AuthorizeCardRequest) (*salemongo.CardAuthTransaction, error)
	Reverse(ctx context.Context, request *requests.ReverseTransactionRequest) (*salemongo.CardAuthTransaction, error)
	GetAuthTransaction(ctx context.Context, txnID string) (*salemongo.CardAuthTransaction, error)
	GetAuthTransactionsByCardNumber(ctx context.Context, cardNumber string, page, pageSize int) ([]*salemongo.GetAuthTxnByCardNumberRow, error)
}

type CardPaymentService interface {
	PostPayment(ctx context.Context, request *requests.PostPaymentRequest) (*salemongo.CardPayment, error)
	GetPaymentHistory(ctx context.Context, cardNumber string, page, pageSize int) ([]*salemongo.CardPaymentListRow, error)
	CountPayments(ctx context.Context, cardNumber string) (int, error)
}

type CardRewardService interface {
	EarnRewards(ctx context.Context, request *requests.EarnRewardsRequest) (*salemongo.CardReward, error)
	GetBalance(ctx context.Context, cardNumber string) (int64, error)
	GetHistory(ctx context.Context, cardNumber string) ([]*salemongo.CardReward, error)
	RedeemRewards(ctx context.Context, cardNumber string, points int64) (int64, error)
}

type CardDashboardService interface {
	DashboardCard(ctx context.Context) (*response.DashboardCard, error)
	DashboardCardCardNumber(ctx context.Context, cardNumber string) (*response.DashboardCardCardNumber, error)
}

type CardQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllCards) ([]*salemongo.Card, *int, error)
	FindByActive(ctx context.Context, req *requests.FindAllCards) ([]*salemongo.Card, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllCards) ([]*salemongo.Card, *int, error)
	FindById(ctx context.Context, cardID int) (*salemongo.Card, error)
	FindByUserID(ctx context.Context, userID int) (*salemongo.Card, error)
	FindByCardNumber(ctx context.Context, cardNumber string) (*salemongo.Card, error)
	FindUserCardByCardNumber(ctx context.Context, cardNumber string) (*salemongo.Card, error)
}

type CardCommandService interface {
	CreateCard(ctx context.Context, request *requests.CreateCardRequest) (*salemongo.Card, error)
	UpdateCard(ctx context.Context, request *requests.UpdateCardRequest) (*salemongo.Card, error)
	TrashedCard(ctx context.Context, cardID int) (*salemongo.Card, error)
	RestoreCard(ctx context.Context, cardID int) (*salemongo.Card, error)
	DeleteCardPermanent(ctx context.Context, cardID int) (bool, error)
	RestoreAllCard(ctx context.Context) (bool, error)
	DeleteAllCardPermanent(ctx context.Context) (bool, error)
	ToggleCardStatus(ctx context.Context, request *requests.ToggleCardStatusRequest) (*salemongo.Card, error)
	UpdateCreditLimit(ctx context.Context, request *requests.UpdateCreditLimitRequest) (*salemongo.Card, error)
	RedeemPoints(ctx context.Context, request *requests.RedeemPointsRequest) (*salemongo.Card, error)
	ProcessBillingCycles(ctx context.Context) error
}

type BillingEngineService interface {
	TriggerBillingCycle(ctx context.Context, billingCycleDay int) (int, error)
	GetStatement(ctx context.Context, cardNumber string) (*salemongo.BillingCycle, error)
	GetStatementsByCard(ctx context.Context, cardNumber string, page, pageSize int) ([]*salemongo.BillingCycle, error)
	GetBillingCyclesByCardNumber(ctx context.Context, cardNumber string) ([]*salemongo.BillingCycle, error)
}
