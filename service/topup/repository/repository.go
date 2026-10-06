package repository

import (
	pbcard "github.com/MamangRust/monolith-payment-gateway-pb/card"
	pbsaldo "github.com/MamangRust/monolith-payment-gateway-pb/saldo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	"go.mongodb.org/mongo-driver/mongo"
)

// GuardOptions collects the dependency guards applied to the adapters built
// inside NewRepositories, so every outbound gRPC call runs under a per-call
// deadline, bulkhead and circuit breaker.
type GuardOptions struct {
	Card  []adapter.GuardOption
	Saldo []adapter.GuardOption
}

type Repositories interface {
	TopupQueryRepository
	TopupCommandRepository
	CardRepository
	SaldoRepository
	IdempotencyRepository
	OutboxRepository
	CardAdapter() adapter.CardAdapter
	SaldoAdapter() adapter.SaldoAdapter
}

type repositories struct {
	TopupQueryRepository
	TopupCommandRepository
	CardRepository
	SaldoRepository
	IdempotencyRepository
	OutboxRepository
	cardAdapter  adapter.CardAdapter
	saldoAdapter adapter.SaldoAdapter
}

// NewRepositories assembles the topup repositories and builds the card/saldo
// gRPC adapters from the raw clients under their guards. The same guarded
// instances back CardAdapter()/SaldoAdapter() so the service wiring and the
// repository share one breaker per dependency.
func NewRepositories(
	db *mongo.Database,
	cardQuery pbcard.CardQueryServiceClient,
	cardCmd pbcard.CardCommandServiceClient,
	saldoQuery pbsaldo.SaldoQueryServiceClient,
	saldoCmd pbsaldo.SaldoCommandServiceClient,
	guards ...GuardOptions,
) Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	topupColl := db.Collection("topups")
	cardAdapter := adapter.NewCardAdapter(cardQuery, cardCmd, g.Card...)
	saldoAdapter := adapter.NewSaldoAdapter(saldoQuery, saldoCmd, g.Saldo...)

	return &repositories{
		TopupQueryRepository:   NewTopupQueryRepository(topupColl),
		TopupCommandRepository: NewTopupCommandRepository(db, topupColl),
		CardRepository:         cardAdapter,
		SaldoRepository:        saldoAdapter,
		IdempotencyRepository:  NewTopupIdempotencyRepository(db.Collection("idempotency_records")),
		OutboxRepository:       NewOutboxMongoStore(db.Collection("outbox_events")),
		cardAdapter:            cardAdapter,
		saldoAdapter:           saldoAdapter,
	}
}

// NewRepositoriesWithAdapters assembles the topup repositories around
// pre-built card/saldo implementations. Tests use it to inject local or fake
// implementations directly; production goes through NewRepositories.
func NewRepositoriesWithAdapters(db *mongo.Database, card CardRepository, saldo SaldoRepository) Repositories {
	topupColl := db.Collection("topups")
	r := &repositories{
		TopupQueryRepository:   NewTopupQueryRepository(topupColl),
		TopupCommandRepository: NewTopupCommandRepository(db, topupColl),
		CardRepository:         card,
		SaldoRepository:        saldo,
		IdempotencyRepository:  NewTopupIdempotencyRepository(db.Collection("idempotency_records")),
		OutboxRepository:       NewOutboxMongoStore(db.Collection("outbox_events")),
	}
	r.cardAdapter, _ = card.(adapter.CardAdapter)
	r.saldoAdapter, _ = saldo.(adapter.SaldoAdapter)
	return r
}

// CardAdapter returns the guarded card gRPC adapter, or nil when constructed
// via NewRepositoriesWithAdapters around a non-adapter implementation.
func (r *repositories) CardAdapter() adapter.CardAdapter { return r.cardAdapter }

// SaldoAdapter returns the guarded saldo gRPC adapter, or nil when constructed
// via NewRepositoriesWithAdapters around a non-adapter implementation.
func (r *repositories) SaldoAdapter() adapter.SaldoAdapter { return r.saldoAdapter }
