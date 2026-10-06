package repository

import (
	"context"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type CardCommandRepository interface {
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
}

type CardQueryRepository interface {
	FindAllCards(ctx context.Context, req *requests.FindAllCards) ([]*salemongo.Card, error)
	FindByActive(ctx context.Context, req *requests.FindAllCards) ([]*salemongo.Card, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllCards) ([]*salemongo.Card, error)
	FindById(ctx context.Context, cardID int) (*salemongo.Card, error)
	FindCardByUserId(ctx context.Context, userID int) (*salemongo.Card, error)
	FindCardByCardNumber(ctx context.Context, cardNumber string) (*salemongo.Card, error)
	FindUserCardByCardNumber(ctx context.Context, cardNumber string) (*salemongo.Card, error)
	CountAllCards(ctx context.Context, search string) (int64, error)
	CountActiveCards(ctx context.Context, search string) (int64, error)
	CountTrashedCards(ctx context.Context, search string) (int64, error)
}

type UserRepository interface {
	FindById(ctx context.Context, userID int) (*salemongo.User, error)
}

type CardAuthTransactionRepository interface {
	InsertPending(ctx context.Context, req *requests.AuthorizeCardRequest) (*salemongo.CardAuthTransaction, error)
	Approve(ctx context.Context, txnID string) (*salemongo.CardAuthTransaction, error)
	Decline(ctx context.Context, txnID string) (*salemongo.CardAuthTransaction, error)
	Reverse(ctx context.Context, txnID string) (*salemongo.CardAuthTransaction, error)
	FindByIdempotencyKey(ctx context.Context, key string) (*salemongo.CardAuthTransaction, error)
	FindByTxnID(ctx context.Context, txnID string) (*salemongo.CardAuthTransaction, error)
	FindByCardNumber(ctx context.Context, cardNumber string, page, pageSize int) ([]*salemongo.GetAuthTxnByCardNumberRow, error)
	CountRecentByCardNumber(ctx context.Context, cardNumber string, since time.Time) (int, error)
	UpdateRiskScore(ctx context.Context, txnID string, score int) error
}

// CardPaymentResult contains a payment and whether it was returned from an
// idempotent replay.
type CardPaymentResult struct {
	Payment  *salemongo.CardPayment
	Replayed bool
}

type CardPaymentRepository interface {
	PostPayment(ctx context.Context, req *requests.PostPaymentRequest) (*salemongo.CardPayment, error)
	PostPaymentIdempotent(ctx context.Context, req *requests.PostPaymentRequest, expectedBillingStatus, targetBillingStatus string) (*CardPaymentResult, error)
	GetPaymentByReferenceID(ctx context.Context, referenceID string) (*salemongo.CardPayment, error)
	GetPaymentHistory(ctx context.Context, cardNumber string, page, pageSize int) ([]*salemongo.CardPaymentListRow, error)
	CountPayments(ctx context.Context, cardNumber string) (int, error)
}

type CardRewardRepository interface {
	EarnRewards(ctx context.Context, req *requests.EarnRewardsRequest) (*salemongo.CardReward, error)
	GetBalance(ctx context.Context, cardNumber string) (int64, error)
	GetHistory(ctx context.Context, cardNumber string) ([]*salemongo.CardReward, error)
	RedeemRewards(ctx context.Context, cardNumber string, points int64) (int64, error)
}

type BillingCycleRepository interface {
	GetBillingCyclesByCardNumber(ctx context.Context, cardNumber string) ([]*salemongo.BillingCycle, error)
	GetBillingCycleByID(ctx context.Context, billingID int) (*salemongo.BillingCycle, error)
	CreateBillingCycles(ctx context.Context, cycleStart, cycleEnd, dueDate time.Time) (int, error)
}
