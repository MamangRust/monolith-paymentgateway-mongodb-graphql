package service

import (
	"github.com/MamangRust/monolith-payment-gateway-pb/ai_security"
	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	"github.com/MamangRust/monolith-payment-gateway-pkg/kafka"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	mencache "github.com/MamangRust/monolith-payment-gateway-withdraw/redis"
	statscache "github.com/MamangRust/monolith-payment-gateway-withdraw/redis/stats"
	statsbycardcache "github.com/MamangRust/monolith-payment-gateway-withdraw/redis/statsbycard"
	"github.com/MamangRust/monolith-payment-gateway-withdraw/repository"
	statsrepository "github.com/MamangRust/monolith-payment-gateway-withdraw/repository/stats"
	statsbycardrepository "github.com/MamangRust/monolith-payment-gateway-withdraw/repository/statsbycard"
	withdrawstatsservice "github.com/MamangRust/monolith-payment-gateway-withdraw/service/stats"
	withdrawstatsbycardservice "github.com/MamangRust/monolith-payment-gateway-withdraw/service/statsbycard"
	"go.mongodb.org/mongo-driver/mongo"
)

type Service interface {
	WithdrawQueryService
	WithdrawCommandService
	withdrawstatsservice.WithdrawStatsService
	withdrawstatsbycardservice.WithdrawStatsByCardService
}

type service struct {
	WithdrawQueryService
	WithdrawCommandService
	withdrawstatsservice.WithdrawStatsService
	withdrawstatsbycardservice.WithdrawStatsByCardService
}

type Deps struct {
	Kafka                *kafka.Kafka
	Repositories         repository.Repositories
	CardAdapter          adapter.CardAdapter
	SaldoAdapter         adapter.SaldoAdapter
	Logger               logger.LoggerInterface
	Cache                *cache.CacheStore
	AISecurityClient     ai_security.AISecurityServiceClient
	DailyWithdrawalLimit int64
	DB                   *mongo.Database
}

func NewService(deps *Deps) Service {

	cache := mencache.NewMencache(deps.Cache)

	observability, _ := observability.NewObservability("withdraw-service", deps.Logger)

	return &service{
		WithdrawQueryService:   newWithdrawQueryService(deps, observability, cache),
		WithdrawCommandService: newWithdrawCommandService(deps, observability, cache),
		WithdrawStatsService: withdrawstatsservice.NewWithdrawStatsService(&withdrawstatsservice.DepsStats{
			Cache:         statscache.NewWithdrawStatsCache(deps.Cache),
			Logger:        deps.Logger,
			Repository:    statsrepository.NewWithdrawStatsRepository(deps.DB),
			Observability: observability,
		}),
		WithdrawStatsByCardService: withdrawstatsbycardservice.NewWithdrawStatsByCardService(&withdrawstatsbycardservice.DepsStatsByCard{
			Cache:         statsbycardcache.NewWithdrawStatsByCardCache(deps.Cache),
			Logger:        deps.Logger,
			Repository:    statsbycardrepository.NewWithdrawStatsByCardRepository(deps.DB),
			Observability: observability,
		}),
	}
}

func newWithdrawQueryService(deps *Deps, observability observability.TraceLoggerObservability, cache mencache.Mencache) WithdrawQueryService {
	return NewWithdrawQueryService(
		&withdrawQueryServiceDeps{
			Cache:         cache,
			Repository:    deps.Repositories,
			Logger:        deps.Logger,
			Observability: observability,
		},
	)
}

func newWithdrawCommandService(deps *Deps, observability observability.TraceLoggerObservability, cache mencache.Mencache) WithdrawCommandService {
	return NewWithdrawCommandService(
		&WithdrawCommandServiceDeps{
			Cache:                cache,
			Kafka:                deps.Kafka,
			CardAdapter:          deps.CardAdapter,
			SaldoAdapter:         deps.SaldoAdapter,
			CommandRepository:    deps.Repositories,
			QueryRepository:      deps.Repositories,
			IdempotencyStore:     deps.Repositories,
			OutboxStore:          deps.Repositories,
			Logger:               deps.Logger,
			Observability:        observability,
			AISecurityClient:     deps.AISecurityClient,
			DailyWithdrawalLimit: deps.DailyWithdrawalLimit,
		},
	)
}
