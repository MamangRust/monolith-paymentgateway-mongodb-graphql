package service

import (
	mencache "github.com/MamangRust/monolith-payment-gateway-card/redis"
	cardauthmencache "github.com/MamangRust/monolith-payment-gateway-card/redis/auth"
	carddashboardmencache "github.com/MamangRust/monolith-payment-gateway-card/redis/dashboard"
	cardstatsmencache "github.com/MamangRust/monolith-payment-gateway-card/redis/stats"
	cardstatsbycardmencache "github.com/MamangRust/monolith-payment-gateway-card/redis/statsbycard"
	"github.com/MamangRust/monolith-payment-gateway-card/repository"
	cardstatsservice "github.com/MamangRust/monolith-payment-gateway-card/service/stats"
	cardstatsbycard "github.com/MamangRust/monolith-payment-gateway-card/service/statsbycard"
	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	"github.com/MamangRust/monolith-payment-gateway-pkg/kafka"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
)

type Service interface {
	CardQueryService
	CardCommandService
	CardDashboardService
	cardstatsservice.CardStatsService
	cardstatsbycard.CardStatsByCardService
	CardAuthorizationService
	CardPaymentService
	CardRewardService
	BillingEngineService
}

type service struct {
	CardQueryService
	CardCommandService
	CardDashboardService
	cardstatsservice.CardStatsService
	cardstatsbycard.CardStatsByCardService
	CardAuthorizationService
	CardPaymentService
	CardRewardService
	BillingEngineService
}

type Deps struct {
	Cache           *cache.CacheStore
	Repositories    *repository.Repositories
	UserAdapter     adapter.UserAdapter
	Logger          logger.LoggerInterface
	Kafka           *kafka.Kafka
	BillingCycleDay int
}

func NewService(deps *Deps) Service {
	observability, _ := observability.NewObservability("card-server", deps.Logger)

	cache := mencache.NewMencache(deps.Cache)
	dashCache := carddashboardmencache.NewMencacheDashboard(deps.Cache)
	statsCache := cardstatsmencache.NewMencacheStats(deps.Cache)
	statsByCardCache := cardstatsbycardmencache.NewMencacheStatsByCard(deps.Cache)

	return &service{
		CardQueryService:     newCardQuery(deps, observability, cache),
		CardCommandService:   newCardCommand(deps, observability, cache),
		CardDashboardService: newCardDashboard(deps, observability, dashCache),
		CardStatsService: cardstatsservice.NewCardStatsService(&cardstatsservice.DepsStats{
			Mencache:      statsCache,
			Repositories:  deps.Repositories.CardStatistic,
			Logger:        deps.Logger,
			Observability: observability,
		}),
		CardStatsByCardService: cardstatsbycard.NewCardStatsByCardService(&cardstatsbycard.DepsStatsByCard{
			Mencache:      statsByCardCache,
			Repositories:  deps.Repositories.CardStatisticByCard,
			Logger:        deps.Logger,
			Observability: observability,
		}),
		CardAuthorizationService: newCardAuthorization(deps, observability, cache),
		CardPaymentService:       newCardPayment(deps, observability),
		CardRewardService:        newCardReward(deps, observability),
		BillingEngineService:     newBillingEngine(deps, observability),
	}
}

func newCardQuery(deps *Deps, observability observability.TraceLoggerObservability, cache mencache.Mencache) CardQueryService {
	return NewCardQueryService(&cardQueryServiceDeps{
		Cache:               cache,
		CardQueryRepository: deps.Repositories.CardQuery,
		UserRepository:      deps.Repositories.User,
		Logger:              deps.Logger,
		Observability:       observability,
	})
}

func newCardDashboard(deps *Deps, observability observability.TraceLoggerObservability, cache carddashboardmencache.CardDashboardCache) CardDashboardService {
	return NewCardDashboardService(&cardDashboardDeps{
		Cache:                   cache,
		CardDashboardRepository: deps.Repositories.CardDashboard,
		Logger:                  deps.Logger,
		Observability:           observability,
	})
}

func newCardCommand(deps *Deps, observability observability.TraceLoggerObservability, cache mencache.Mencache) CardCommandService {
	billingEngine := newBillingEngine(deps, observability)
	return NewCardCommandService(&cardCommandServiceDeps{
		Cache:                 cache,
		Kafka:                 deps.Kafka,
		UserAdapter:           deps.UserAdapter,
		CardCommandRepository: deps.Repositories.CardCommand,
		BillingEngine:         billingEngine,
		BillingCycleDay:       deps.BillingCycleDay,
		Logger:                deps.Logger,
		Observability:         observability,
	})
}

func newCardAuthorization(deps *Deps, observability observability.TraceLoggerObservability, cache mencache.Mencache) CardAuthorizationService {
	return NewCardAuthorizationService(&CardAuthorizationServiceDeps{
		Cache:                  cache,
		AuthCache:              cardauthmencache.NewCardAuthCache(deps.Cache),
		CardCommandRepository:  deps.Repositories.CardCommand,
		CardAuthRepository:     deps.Repositories.CardAuthTransaction,
		CardPaymentRepository:  deps.Repositories.CardPayment,
		CardRewardRepository:   deps.Repositories.CardReward,
		BillingCycleRepository: deps.Repositories.BillingCycle,
		CardQueryRepository:    deps.Repositories.CardQuery,
		Logger:                 deps.Logger,
		Observability:          observability,
	})
}

func newCardPayment(deps *Deps, observability observability.TraceLoggerObservability) CardPaymentService {
	return NewCardPaymentService(&CardPaymentServiceDeps{
		CardPaymentRepository:  deps.Repositories.CardPayment,
		BillingCycleRepository: deps.Repositories.BillingCycle,
		CardCommandRepository:  deps.Repositories.CardCommand,
		Logger:                 deps.Logger,
		Observability:          observability,
	})
}

func newCardReward(deps *Deps, observability observability.TraceLoggerObservability) CardRewardService {
	return NewCardRewardService(&CardRewardServiceDeps{
		CardRewardRepository:  deps.Repositories.CardReward,
		CardCommandRepository: deps.Repositories.CardCommand,
		Logger:                deps.Logger,
		Observability:         observability,
	})
}

func newBillingEngine(deps *Deps, observability observability.TraceLoggerObservability) BillingEngineService {
	return NewBillingEngineService(&BillingEngineServiceDeps{
		BillingCycleRepository: deps.Repositories.BillingCycle,
		Kafka:                  deps.Kafka,
		Logger:                 deps.Logger,
		Observability:          observability,
	})
}
