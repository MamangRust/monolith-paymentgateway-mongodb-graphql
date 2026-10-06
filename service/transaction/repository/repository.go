package repository

import (
	pbcard "github.com/MamangRust/monolith-payment-gateway-pb/card"
	pbmerchant "github.com/MamangRust/monolith-payment-gateway-pb/merchant"
	pbsaldo "github.com/MamangRust/monolith-payment-gateway-pb/saldo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	"go.mongodb.org/mongo-driver/mongo"
)

// GuardOptions collects the dependency guards applied to the adapters built
// inside NewRepositories, so every outbound gRPC call runs under a per-call
// deadline, bulkhead and circuit breaker.
type GuardOptions struct {
	Saldo    []adapter.GuardOption
	Card     []adapter.GuardOption
	Merchant []adapter.GuardOption
}

type Repositories interface {
	SaldoRepository
	MerchantRepository
	CardRepository
	TransactionQueryRepository
	TransactionCommandRepository
	IdempotencyRepository
	OutboxRepository
	SaldoAdapter() adapter.SaldoAdapter
	CardAdapter() adapter.CardAdapter
	MerchantAdapter() adapter.MerchantAdapter
}

type repositories struct {
	SaldoRepository
	MerchantRepository
	CardRepository
	TransactionQueryRepository
	TransactionCommandRepository
	IdempotencyRepository
	OutboxRepository
	saldoAdapter    adapter.SaldoAdapter
	cardAdapter     adapter.CardAdapter
	merchantAdapter adapter.MerchantAdapter
}

// NewRepositories assembles the transaction repositories and builds the
// saldo/card/merchant gRPC adapters from the raw clients under their guards.
// The same guarded instances back SaldoAdapter()/CardAdapter()/MerchantAdapter()
// so the service wiring and the repository share one breaker per dependency.
func NewRepositories(
	db *mongo.Database,
	saldoQuery pbsaldo.SaldoQueryServiceClient,
	saldoCmd pbsaldo.SaldoCommandServiceClient,
	cardQuery pbcard.CardQueryServiceClient,
	cardCmd pbcard.CardCommandServiceClient,
	merchantQuery pbmerchant.MerchantQueryServiceClient,
	guards ...GuardOptions,
) Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	txColl := db.Collection("transactions")
	saldoAdapter := adapter.NewSaldoAdapter(saldoQuery, saldoCmd, g.Saldo...)
	cardAdapter := adapter.NewCardAdapter(cardQuery, cardCmd, g.Card...)
	merchantAdapter := adapter.NewMerchantAdapter(merchantQuery, g.Merchant...)

	return &repositories{
		SaldoRepository:              saldoAdapter,
		MerchantRepository:           merchantAdapter,
		CardRepository:               cardAdapter,
		TransactionQueryRepository:   NewTransactionQueryRepository(txColl),
		TransactionCommandRepository: NewTransactionCommandRepository(db, txColl),
		IdempotencyRepository:        NewTransactionIdempotencyRepository(db.Collection("idempotency_records")),
		OutboxRepository:             NewOutboxMongoStore(db.Collection("outbox_events")),
		saldoAdapter:                 saldoAdapter,
		cardAdapter:                  cardAdapter,
		merchantAdapter:              merchantAdapter,
	}
}

// NewRepositoriesWithAdapters assembles the transaction repositories around
// pre-built saldo/card/merchant implementations. Tests use it to inject local
// or fake implementations directly; production goes through NewRepositories.
func NewRepositoriesWithAdapters(db *mongo.Database, saldo SaldoRepository, card CardRepository, merchant MerchantRepository) Repositories {
	txColl := db.Collection("transactions")
	r := &repositories{
		SaldoRepository:              saldo,
		MerchantRepository:           merchant,
		CardRepository:               card,
		TransactionQueryRepository:   NewTransactionQueryRepository(txColl),
		TransactionCommandRepository: NewTransactionCommandRepository(db, txColl),
		IdempotencyRepository:        NewTransactionIdempotencyRepository(db.Collection("idempotency_records")),
		OutboxRepository:             NewOutboxMongoStore(db.Collection("outbox_events")),
	}
	r.saldoAdapter, _ = saldo.(adapter.SaldoAdapter)
	r.cardAdapter, _ = card.(adapter.CardAdapter)
	r.merchantAdapter, _ = merchant.(adapter.MerchantAdapter)
	return r
}

// SaldoAdapter returns the guarded saldo gRPC adapter, or nil when constructed
// via NewRepositoriesWithAdapters around a non-adapter implementation.
func (r *repositories) SaldoAdapter() adapter.SaldoAdapter { return r.saldoAdapter }

// CardAdapter returns the guarded card gRPC adapter, or nil when constructed
// via NewRepositoriesWithAdapters around a non-adapter implementation.
func (r *repositories) CardAdapter() adapter.CardAdapter { return r.cardAdapter }

// MerchantAdapter returns the guarded merchant gRPC adapter, or nil when
// constructed via NewRepositoriesWithAdapters around a non-adapter
// implementation.
func (r *repositories) MerchantAdapter() adapter.MerchantAdapter { return r.merchantAdapter }
