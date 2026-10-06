package service

import (
	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	"github.com/MamangRust/monolith-payment-gateway-pkg/kafka"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	mencache "github.com/MamangRust/monolith-payment-gateway-topup/redis"
	statscache "github.com/MamangRust/monolith-payment-gateway-topup/redis/stats"
	statsbycardcache "github.com/MamangRust/monolith-payment-gateway-topup/redis/statsbycard"
	"github.com/MamangRust/monolith-payment-gateway-topup/repository"
	statsrepository "github.com/MamangRust/monolith-payment-gateway-topup/repository/stats"
	statsbycardrepository "github.com/MamangRust/monolith-payment-gateway-topup/repository/statsbycard"
	topupstatsservice "github.com/MamangRust/monolith-payment-gateway-topup/service/stats"
	topupstatsbycardservice "github.com/MamangRust/monolith-payment-gateway-topup/service/statsbycard"
	"go.mongodb.org/mongo-driver/mongo"
)

type Service interface {
	TopupQueryService
	TopupCommandService
	topupstatsservice.TopupStatsService
	topupstatsbycardservice.TopupStatsByCardService
}

type service struct {
	TopupQueryService
	TopupCommandService
	topupstatsservice.TopupStatsService
	topupstatsbycardservice.TopupStatsByCardService
}

type Deps struct {
	Kafka        *kafka.Kafka
	Cache        *cache.CacheStore
	Repositories repository.Repositories
	CardAdapter  adapter.CardAdapter
	SaldoAdapter adapter.SaldoAdapter
	Logger       logger.LoggerInterface
	DB           *mongo.Database
}

func NewService(deps *Deps) Service {
	cache := mencache.NewMencache(deps.Cache)

	observability, _ := observability.NewObservability("topup-service", deps.Logger)

	return &service{
		TopupQueryService:   newTopupQueryService(deps, observability, cache),
		TopupCommandService: newTopupCommandService(deps, observability, cache),
		TopupStatsService: topupstatsservice.NewTopupStatsService(&topupstatsservice.DepsStats{
			Cache:         statscache.NewTopupStatsCache(deps.Cache),
			Logger:        deps.Logger,
			Repository:    statsrepository.NewTopupStatsRepository(deps.DB),
			Observability: observability,
		}),
		TopupStatsByCardService: topupstatsbycardservice.NewTopupStatsByCardService(&topupstatsbycardservice.DepsStatsByCard{
			Cache:         statsbycardcache.NewTopupStatsByCardCache(deps.Cache),
			Logger:        deps.Logger,
			Repository:    statsbycardrepository.NewTopupStatsByCardRepository(deps.DB),
			Observability: observability,
		}),
	}
}

func newTopupQueryService(deps *Deps, observability observability.TraceLoggerObservability, cache mencache.TopupQueryCache) TopupQueryService {
	return NewTopupQueryService(&topupQueryDeps{
		Cache:         cache,
		Repository:    deps.Repositories,
		Logger:        deps.Logger,
		Observability: observability,
	})
}

func newTopupCommandService(deps *Deps, observability observability.TraceLoggerObservability, cache mencache.TopupCommandCache) TopupCommandService {
	return NewTopupCommandService(&TopupCommandDeps{
		Kafka:                  deps.Kafka,
		Cache:                  cache,
		CardAdapter:            deps.CardAdapter,
		TopupQueryRepository:   deps.Repositories,
		TopupCommandRepository: deps.Repositories,
		SaldoAdapter:           deps.SaldoAdapter,
		IdempotencyStore:       deps.Repositories,
		OutboxStore:            deps.Repositories,
		Logger:                 deps.Logger,
		Observability:          observability,
	})
}
