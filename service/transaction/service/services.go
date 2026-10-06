package service

import (
	"github.com/MamangRust/monolith-payment-gateway-pb/ai_security"
	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	"github.com/MamangRust/monolith-payment-gateway-pkg/kafka"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	mencache "github.com/MamangRust/monolith-payment-gateway-transaction/redis"
	statscache "github.com/MamangRust/monolith-payment-gateway-transaction/redis/stats"
	statsbycardcache "github.com/MamangRust/monolith-payment-gateway-transaction/redis/statsbycard"
	"github.com/MamangRust/monolith-payment-gateway-transaction/repository"
	statsrepository "github.com/MamangRust/monolith-payment-gateway-transaction/repository/stats"
	statsbycardrepository "github.com/MamangRust/monolith-payment-gateway-transaction/repository/statsbycard"
	transactionstatsservice "github.com/MamangRust/monolith-payment-gateway-transaction/service/stats"
	transactionstatsbycardservice "github.com/MamangRust/monolith-payment-gateway-transaction/service/statsbycard"
	"go.mongodb.org/mongo-driver/mongo"
)

// Service is a struct that contains all the services
type service struct {
	TransactionQueryService
	TransactionCommandService
	transactionstatsservice.TransactionStatsService
	transactionstatsbycardservice.TransactionStatsByCardService
}

type Service interface {
	TransactionQueryService
	TransactionCommandService
	transactionstatsservice.TransactionStatsService
	transactionstatsbycardservice.TransactionStatsByCardService
}

type Deps struct {
	Kafka            *kafka.Kafka
	Repositories     repository.Repositories
	MerchantAdapter  adapter.MerchantAdapter
	CardAdapter      adapter.CardAdapter
	SaldoAdapter     adapter.SaldoAdapter
	Logger           logger.LoggerInterface
	Cache            *cache.CacheStore
	AISecurityClient ai_security.AISecurityServiceClient
	DB               *mongo.Database
}

func NewService(deps *Deps) Service {
	cache := mencache.NewMencache(deps.Cache)
	observability, _ := observability.NewObservability("transaction-service", deps.Logger)

	return &service{
		TransactionQueryService:   newTransactionQueryService(deps, observability, cache),
		TransactionCommandService: newTransactionCommandService(deps, observability, cache),
		TransactionStatsService: transactionstatsservice.NewTransactionStatsService(&transactionstatsservice.DepsStats{
			Cache:         statscache.NewTransactionStatsCache(deps.Cache),
			Logger:        deps.Logger,
			Repository:    statsrepository.NewTransactionStatsRepository(deps.DB),
			Observability: observability,
		}),
		TransactionStatsByCardService: transactionstatsbycardservice.NewTransactionStatsByCardService(&transactionstatsbycardservice.DepsStats{
			Cache:         statsbycardcache.NewTransactionStatsByCardCache(deps.Cache),
			Logger:        deps.Logger,
			Repository:    statsbycardrepository.NewTransactionStatsByCardRepository(deps.DB),
			Observability: observability,
		}),
	}
}

func newTransactionQueryService(deps *Deps, observability observability.TraceLoggerObservability, cache mencache.Mencache) TransactionQueryService {
	return NewTransactionQueryService(&transactionQueryServiceDeps{
		Cache:                      cache,
		TransactionQueryRepository: deps.Repositories,
		Logger:                     deps.Logger,
		Observability:              observability,
	})
}

func newTransactionCommandService(deps *Deps, observability observability.TraceLoggerObservability, cache mencache.Mencache) TransactionCommandService {
	return NewTransactionCommandService(&TransactionCommandServiceDeps{
		Kafka:                        deps.Kafka,
		Mencache:                     cache,
		MerchantAdapter:              deps.MerchantAdapter,
		CardAdapter:                  deps.CardAdapter,
		SaldoAdapter:                 deps.SaldoAdapter,
		TransactionCommandRepository: deps.Repositories,
		TransactionQueryRepository:   deps.Repositories,
		IdempotencyStore:             deps.Repositories,
		OutboxStore:                  deps.Repositories,
		Logger:                       deps.Logger,
		Observability:                observability,
		AISecurityClient:             deps.AISecurityClient,
	})
}
