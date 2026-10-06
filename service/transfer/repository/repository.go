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
	Saldo []adapter.GuardOption
	Card  []adapter.GuardOption
}

type Repositories interface {
	SaldoRepository
	TransferQueryRepository
	TransferCommandRepository
	CardRepository
	IdempotencyRepository
	OutboxRepository
	SaldoAdapter() adapter.SaldoAdapter
	CardAdapter() adapter.CardAdapter
}

type repositories struct {
	SaldoRepository
	TransferQueryRepository
	TransferCommandRepository
	CardRepository
	IdempotencyRepository
	OutboxRepository
	saldoAdapter adapter.SaldoAdapter
	cardAdapter  adapter.CardAdapter
}

// NewRepositories assembles the transfer repositories and builds the
// saldo/card gRPC adapters from the raw clients under their guards. The same
// guarded instances back SaldoAdapter()/CardAdapter() so the service wiring
// and the repository share one breaker per dependency.
func NewRepositories(
	db *mongo.Database,
	saldoQuery pbsaldo.SaldoQueryServiceClient,
	saldoCmd pbsaldo.SaldoCommandServiceClient,
	cardQuery pbcard.CardQueryServiceClient,
	cardCmd pbcard.CardCommandServiceClient,
	guards ...GuardOptions,
) Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	transferColl := db.Collection("transfers")
	saldoAdapter := adapter.NewSaldoAdapter(saldoQuery, saldoCmd, g.Saldo...)
	cardAdapter := adapter.NewCardAdapter(cardQuery, cardCmd, g.Card...)

	return &repositories{
		SaldoRepository:           saldoAdapter,
		TransferQueryRepository:   NewTransferQueryRepository(transferColl),
		TransferCommandRepository: NewTransferCommandRepository(db, transferColl, NewTransferIdempotencyRepository(db.Collection("idempotency_records"))),
		CardRepository:            cardAdapter,
		IdempotencyRepository:     NewTransferIdempotencyRepository(db.Collection("idempotency_records")),
		OutboxRepository:          NewOutboxMongoStore(db.Collection("outbox_events")),
		saldoAdapter:              saldoAdapter,
		cardAdapter:               cardAdapter,
	}
}

// NewRepositoriesWithAdapters assembles the transfer repositories around
// pre-built saldo/card implementations. Tests use it to inject local or fake
// implementations directly; production goes through NewRepositories.
func NewRepositoriesWithAdapters(db *mongo.Database, saldo SaldoRepository, card CardRepository) Repositories {
	transferColl := db.Collection("transfers")
	r := &repositories{
		SaldoRepository:           saldo,
		TransferQueryRepository:   NewTransferQueryRepository(transferColl),
		TransferCommandRepository: NewTransferCommandRepository(db, transferColl, NewTransferIdempotencyRepository(db.Collection("idempotency_records"))),
		CardRepository:            card,
		IdempotencyRepository:     NewTransferIdempotencyRepository(db.Collection("idempotency_records")),
		OutboxRepository:          NewOutboxMongoStore(db.Collection("outbox_events")),
	}
	r.saldoAdapter, _ = saldo.(adapter.SaldoAdapter)
	r.cardAdapter, _ = card.(adapter.CardAdapter)
	return r
}

// SaldoAdapter returns the guarded saldo gRPC adapter, or nil when constructed
// via NewRepositoriesWithAdapters around a non-adapter implementation.
func (r *repositories) SaldoAdapter() adapter.SaldoAdapter { return r.saldoAdapter }

// CardAdapter returns the guarded card gRPC adapter, or nil when constructed
// via NewRepositoriesWithAdapters around a non-adapter implementation.
func (r *repositories) CardAdapter() adapter.CardAdapter { return r.cardAdapter }
