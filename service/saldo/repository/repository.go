package repository

import (
	pbcard "github.com/MamangRust/monolith-payment-gateway-pb/card"
	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	"go.mongodb.org/mongo-driver/mongo"
)

// GuardOptions collects the dependency guards applied to the card adapter
// built inside NewRepositories, so every outbound gRPC call runs under a
// per-call deadline, bulkhead and circuit breaker.
type GuardOptions struct {
	Card []adapter.GuardOption
}

type Repositories interface {
	SaldoQueryRepository
	SaldoCommandRepository
	CardRepository
	CardAdapter() adapter.CardAdapter
}

type repositories struct {
	SaldoQueryRepository
	SaldoCommandRepository
	CardRepository
	cardAdapter adapter.CardAdapter
}

// NewRepositories assembles the saldo repositories: Mongo-backed saldo/card
// lookups plus the card gRPC adapter, constructed here under its guard. The
// same guarded instance backs CardAdapter() so the service wiring and the
// repository share one breaker.
func NewRepositories(
	db *mongo.Database,
	cardQuery pbcard.CardQueryServiceClient,
	cardCmd pbcard.CardCommandServiceClient,
	guards ...GuardOptions,
) Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	return &repositories{
		SaldoQueryRepository:   NewSaldoQueryRepository(db.Collection("saldos")),
		SaldoCommandRepository: NewSaldoCommandRepository(db.Collection("saldos")),
		CardRepository:         NewCardRepository(db),
		cardAdapter:            adapter.NewCardAdapter(cardQuery, cardCmd, g.Card...),
	}
}

// CardAdapter returns the guarded card gRPC adapter built by
// NewRepositories.
func (r *repositories) CardAdapter() adapter.CardAdapter {
	return r.cardAdapter
}
