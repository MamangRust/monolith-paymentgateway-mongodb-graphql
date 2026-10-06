package repository

import (
	carddashboardrepo "github.com/MamangRust/monolith-payment-gateway-card/repository/dashboard"
	repositorystats "github.com/MamangRust/monolith-payment-gateway-card/repository/stats"
	repositorystatsbycard "github.com/MamangRust/monolith-payment-gateway-card/repository/statsbycard"
	pbuser "github.com/MamangRust/monolith-payment-gateway-pb/user"
	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	"go.mongodb.org/mongo-driver/mongo"
)

// Repositories contains all repositories used by the card service.
type Repositories struct {
	CardCommand         CardCommandRepository
	CardQuery           CardQueryRepository
	User                UserRepository
	CardAuthTransaction CardAuthTransactionRepository
	CardPayment         CardPaymentRepository
	CardReward          CardRewardRepository
	BillingCycle        BillingCycleRepository
	CardDashboard       carddashboardrepo.CardDashboardRepository
	CardStatistic       repositorystats.CardStatsRepository
	CardStatisticByCard repositorystatsbycard.CardStatsByCardRepository
}

// GuardOptions collects the dependency guards applied to the adapters built
// inside NewRepositories, so every outbound gRPC call runs under a per-call
// deadline, bulkhead and circuit breaker.
type GuardOptions struct {
	User []adapter.GuardOption
}

// NewRepositories assembles the card repositories: Mongo-backed card/token
// repos plus the shared user gRPC adapter, constructed here under its guard.
// The same guarded instance backs UserAdapter() so the service wiring and the
// repository share one breaker.
func NewRepositories(db *mongo.Database, userClient pbuser.UserQueryServiceClient, guards ...GuardOptions) *Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	return &Repositories{
		CardQuery:           NewCardQueryRepository(db.Collection("cards")),
		CardCommand:         NewCardCommandRepository(db.Collection("cards")),
		User:                adapter.NewUserAdapter(userClient, g.User...),
		CardAuthTransaction: NewCardAuthTransactionRepository(db.Collection("card_auth_transactions")),
		CardPayment:         NewCardPaymentRepository(db.Collection("card_payments")),
		CardReward:          NewCardRewardRepository(db.Collection("card_rewards")),
		BillingCycle:        NewBillingCycleRepository(db.Collection("billing_cycles")),
		CardDashboard:       carddashboardrepo.NewCardDashboardRepository(db),
		CardStatistic:       repositorystats.NewCardStatsRepository(db),
		CardStatisticByCard: repositorystatsbycard.NewCardStatsByCardRepository(db),
	}
}

// UserAdapter returns the guarded user gRPC adapter built by NewRepositories.
// It satisfies the full adapter contract (including methods beyond the
// UserRepository query surface) for the card command service.
func (r *Repositories) UserAdapter() adapter.UserAdapter {
	a, _ := r.User.(adapter.UserAdapter)
	return a
}
