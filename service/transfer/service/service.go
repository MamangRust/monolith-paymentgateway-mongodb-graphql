package service

import (
	"github.com/MamangRust/monolith-payment-gateway-pkg/kafka"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"

	"github.com/MamangRust/monolith-payment-gateway-pb/ai_security"
	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	mencache "github.com/MamangRust/monolith-payment-gateway-transfer/redis"
	statscache "github.com/MamangRust/monolith-payment-gateway-transfer/redis/stats"
	statsbycardcache "github.com/MamangRust/monolith-payment-gateway-transfer/redis/statsbycard"
	"github.com/MamangRust/monolith-payment-gateway-transfer/repository"
	statsrepository "github.com/MamangRust/monolith-payment-gateway-transfer/repository/stats"
	statsbycardrepository "github.com/MamangRust/monolith-payment-gateway-transfer/repository/statsbycard"
	transferstatsservice "github.com/MamangRust/monolith-payment-gateway-transfer/service/stats"
	transferstatsbycardservice "github.com/MamangRust/monolith-payment-gateway-transfer/service/statsbycard"
	"go.mongodb.org/mongo-driver/mongo"
)

type Service interface {
	TransferQueryService
	TransferCommandService
	transferstatsservice.TransferStatsService
	transferstatsbycardservice.TransferStatsByCardService
}

type service struct {
	TransferQueryService
	TransferCommandService
	transferstatsservice.TransferStatsService
	transferstatsbycardservice.TransferStatsByCardService
}

type Deps struct {
	Kafka            *kafka.Kafka
	Cache            *cache.CacheStore
	Repositories     repository.Repositories
	CardAdapter      adapter.CardAdapter
	SaldoAdapter     adapter.SaldoAdapter
	Logger           logger.LoggerInterface
	AISecurityClient ai_security.AISecurityServiceClient
	DB               *mongo.Database
}

func NewService(deps *Deps) Service {
	cache := mencache.NewMencache(deps.Cache)

	observability, _ := observability.NewObservability("transfer-service", deps.Logger)

	return &service{
		TransferQueryService:   newTransferQueryService(deps, observability, cache),
		TransferCommandService: newTransferCommandService(deps, observability, cache),
		TransferStatsService: transferstatsservice.NewTransferStatsService(&transferstatsservice.DepsStats{
			Cache:         statscache.NewTransferStatsCache(deps.Cache),
			Logger:        deps.Logger,
			Repository:    statsrepository.NewTransferStatsRepository(deps.DB),
			Observability: observability,
		}),
		TransferStatsByCardService: transferstatsbycardservice.NewTransferStatsByCardService(&transferstatsbycardservice.DepsStats{
			Cache:         statsbycardcache.NewTransferStatsByCardCache(deps.Cache),
			Logger:        deps.Logger,
			Repository:    statsbycardrepository.NewTransferStatsByCardRepository(deps.DB),
			Observability: observability,
		}),
	}
}

func newTransferQueryService(deps *Deps, observability observability.TraceLoggerObservability, cache mencache.Mencache) TransferQueryService {
	return NewTransferQueryService(
		&transferQueryDeps{
			Cache:         cache,
			Repository:    deps.Repositories,
			Logger:        deps.Logger,
			Observability: observability,
		},
	)
}

func newTransferCommandService(deps *Deps, observability observability.TraceLoggerObservability, cache mencache.Mencache) TransferCommandService {
	return NewTransferCommandService(
		&TransferCommandDeps{
			Kafka:                     deps.Kafka,
			Cache:                     cache,
			CardAdapter:               deps.CardAdapter,
			SaldoAdapter:              deps.SaldoAdapter,
			TransferQueryRepository:   deps.Repositories,
			TransferCommandRepository: deps.Repositories,
			IdempotencyStore:          deps.Repositories,
			OutboxStore:               deps.Repositories,
			Logger:                    deps.Logger,
			Observability:             observability,
			AISecurityClient:          deps.AISecurityClient,
		},
	)
}
