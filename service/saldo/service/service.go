package service

import (
	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	"github.com/MamangRust/monolith-payment-gateway-pkg/kafka"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	mencache "github.com/MamangRust/monolith-payment-gateway-saldo/redis"
	statscache "github.com/MamangRust/monolith-payment-gateway-saldo/redis/stats"
	"github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	statsrepository "github.com/MamangRust/monolith-payment-gateway-saldo/repository/stats"
	saldostatsservice "github.com/MamangRust/monolith-payment-gateway-saldo/service/stats"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	"go.mongodb.org/mongo-driver/mongo"
)

type service struct {
	SaldoQueryService
	SaldoCommandService
	saldostatsservice.SaldoStatsService
}

type Service interface {
	SaldoQueryService
	SaldoCommandService
	saldostatsservice.SaldoStatsService
}

type Deps struct {
	Repositories repository.Repositories
	CardAdapter  adapter.CardAdapter
	Logger       logger.LoggerInterface
	Cache        *cache.CacheStore
	Kafka        *kafka.Kafka
	DB           *mongo.Database
}

func NewService(deps *Deps) Service {
	observability, _ := observability.NewObservability("saldo-service", deps.Logger)
	cache := mencache.NewMencache(deps.Cache)

	return &service{
		SaldoQueryService:   newSaldoQueryService(deps, observability, cache),
		SaldoCommandService: newSaldoCommandService(deps, observability, cache),
		SaldoStatsService: saldostatsservice.NewSaldoStatsService(&saldostatsservice.DepsStats{
			Mencache:      statscache.NewSaldoStatsCache(deps.Cache),
			Logger:        deps.Logger,
			Repository:    statsrepository.NewSaldoStatsRepository(deps.DB),
			Observability: observability,
		}),
	}
}

func newSaldoQueryService(deps *Deps, observabilty observability.TraceLoggerObservability, cache mencache.Mencache) SaldoQueryService {
	return NewSaldoQueryService(&saldoQueryParams{
		Cache:         cache,
		Repository:    deps.Repositories,
		Logger:        deps.Logger,
		Observability: observabilty,
	})
}

func newSaldoCommandService(deps *Deps, observabilty observability.TraceLoggerObservability, cache mencache.Mencache) SaldoCommandService {
	return NewSaldoCommandService(&saldoCommandParams{
		Cache:                  cache,
		saldoCommandRepository: deps.Repositories,
		CardAdapter:            deps.CardAdapter,
		Logger:                 deps.Logger,
		Observability:          observabilty,
		Kafka:                  deps.Kafka,
	})
}
