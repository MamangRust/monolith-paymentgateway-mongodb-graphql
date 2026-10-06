package service

import (
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"

	mencache "github.com/MamangRust/monolith-payment-gateway-merchant/redis"
	statcaches "github.com/MamangRust/monolith-payment-gateway-merchant/redis/stats"
	statcachea "github.com/MamangRust/monolith-payment-gateway-merchant/redis/statsbyapikey"
	statcachem "github.com/MamangRust/monolith-payment-gateway-merchant/redis/statsbymerchant"
	"github.com/MamangRust/monolith-payment-gateway-merchant/repository"
	merchantstatsrepository "github.com/MamangRust/monolith-payment-gateway-merchant/repository/stats"
	merchantstatsbyapikeyrepository "github.com/MamangRust/monolith-payment-gateway-merchant/repository/statsbyapikey"
	merchantstatsbymerchantrepository "github.com/MamangRust/monolith-payment-gateway-merchant/repository/statsbymerchant"
	merchantstatsservice "github.com/MamangRust/monolith-payment-gateway-merchant/service/stats"
	merchantstatsbyapikeyservice "github.com/MamangRust/monolith-payment-gateway-merchant/service/statsbyapikey"
	merchantstatsbymerchantservice "github.com/MamangRust/monolith-payment-gateway-merchant/service/statsbymerchant"
	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	"github.com/MamangRust/monolith-payment-gateway-pkg/kafka"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

type Service interface {
	MerchantQueryService() MerchantQueryService
	MerchantTransactionService() MerchantTransactionService
	MerchantCommandService() MerchantCommandService
	MerchantDocumentCommandService() MerchantDocumentCommandService
	MerchantDocumentQueryService() MerchantDocumentQueryService
	MerchantStatsService() merchantstatsservice.MerchantStatsService
	MerchantStatsByMerchantService() merchantstatsbymerchantservice.MerchantStatsByMerchantService
	MerchantStatsByApiKeyService() merchantstatsbyapikeyservice.MerchantStatsByApiKeyService
}

type service struct {
	merchantQuery           MerchantQueryService
	merchantTransaction     MerchantTransactionService
	merchantCommand         MerchantCommandService
	merchantDocumentCommand MerchantDocumentCommandService
	merchantDocumentQuery   MerchantDocumentQueryService
	merchantStats           merchantstatsservice.MerchantStatsService
	merchantStatsByMerchant merchantstatsbymerchantservice.MerchantStatsByMerchantService
	merchantStatsByApiKey   merchantstatsbyapikeyservice.MerchantStatsByApiKeyService
}

type Deps struct {
	Kafka        *kafka.Kafka
	DB           *mongo.Database
	Repositories repository.Repositories
	UserAdapter  adapter.UserAdapter
	Logger       logger.LoggerInterface
	Cache        *cache.CacheStore
}

func NewService(deps *Deps) Service {
	obs, _ := observability.NewObservability("merchant-service", deps.Logger)
	c := mencache.NewMencache(deps.Cache)

	return &service{
		merchantQuery:           newMerchantQueryService(deps, obs, c),
		merchantTransaction:     newMerchantTransactionService(deps, obs, c),
		merchantCommand:         newMerchantCommandService(deps, obs, c),
		merchantDocumentCommand: newMerchantDocumentCommandService(deps, obs, c),
		merchantDocumentQuery:   newMerchantDocumentQueryService(deps, obs, c),
		merchantStats:           newMerchantStatsService(deps.DB, obs),
		merchantStatsByMerchant: newMerchantStatsByMerchantService(deps.DB, obs),
		merchantStatsByApiKey:   newMerchantStatsByApiKeyService(deps.DB, obs),
	}
}

func (s *service) MerchantQueryService() MerchantQueryService { return s.merchantQuery }
func (s *service) MerchantTransactionService() MerchantTransactionService {
	return s.merchantTransaction
}
func (s *service) MerchantCommandService() MerchantCommandService { return s.merchantCommand }
func (s *service) MerchantDocumentCommandService() MerchantDocumentCommandService {
	return s.merchantDocumentCommand
}
func (s *service) MerchantDocumentQueryService() MerchantDocumentQueryService {
	return s.merchantDocumentQuery
}
func (s *service) MerchantStatsService() merchantstatsservice.MerchantStatsService {
	return s.merchantStats
}
func (s *service) MerchantStatsByMerchantService() merchantstatsbymerchantservice.MerchantStatsByMerchantService {
	return s.merchantStatsByMerchant
}
func (s *service) MerchantStatsByApiKeyService() merchantstatsbyapikeyservice.MerchantStatsByApiKeyService {
	return s.merchantStatsByApiKey
}

func newMerchantQueryService(deps *Deps, obs observability.TraceLoggerObservability, c mencache.Mencache) MerchantQueryService {
	return NewMerchantQueryService(&merchantQueryDeps{
		Repository:    deps.Repositories,
		Cache:         c,
		Logger:        deps.Logger,
		Observability: obs,
	})
}

func newMerchantDocumentQueryService(deps *Deps, obs observability.TraceLoggerObservability, c mencache.Mencache) MerchantDocumentQueryService {
	return NewMerchantDocumentQueryService(&merchantDocumentQueryDeps{
		Repository:    deps.Repositories,
		Cache:         c,
		Logger:        deps.Logger,
		Observability: obs,
	})
}

func newMerchantTransactionService(deps *Deps, obs observability.TraceLoggerObservability, c mencache.Mencache) MerchantTransactionService {
	return NewMerchantTransactionService(&merchantTransactionDeps{
		Repository:    deps.Repositories,
		Cache:         c,
		Logger:        deps.Logger,
		Observability: obs,
	})
}

func newMerchantCommandService(deps *Deps, obs observability.TraceLoggerObservability, c mencache.Mencache) MerchantCommandService {
	return NewMerchantCommandService(&merchantCommandServiceDeps{
		Kafka:                     deps.Kafka,
		UserAdapter:               deps.UserAdapter,
		MerchantQueryRepository:   deps.Repositories,
		MerchantCommandRepository: deps.Repositories,
		Logger:                    deps.Logger,
		Observability:             obs,
		Cache:                     c,
	})
}

func newMerchantDocumentCommandService(deps *Deps, obs observability.TraceLoggerObservability, c mencache.Mencache) MerchantDocumentCommandService {
	return NewMerchantDocumentCommandService(&merchantDocumentCommandDeps{
		Kafka:                   deps.Kafka,
		CommandRepository:       deps.Repositories,
		MerchantQueryRepository: deps.Repositories,
		UserAdapter:             deps.UserAdapter,
		Logger:                  deps.Logger,
		Observability:           obs,
		Cache:                   c,
	})
}

func newMerchantStatsService(db *mongo.Database, obs observability.TraceLoggerObservability) merchantstatsservice.MerchantStatsService {
	return merchantstatsservice.NewMerchantStatsService(&merchantstatsservice.DepsStats{
		Mencache:      statcaches.NewMerchantStatsCache(nil),
		Repository:    merchantstatsrepository.NewMerchantStatsRepository(db),
		Observability: obs,
	})
}

func newMerchantStatsByMerchantService(db *mongo.Database, obs observability.TraceLoggerObservability) merchantstatsbymerchantservice.MerchantStatsByMerchantService {
	return merchantstatsbymerchantservice.NewMerchantStatsByMerchantService(&merchantstatsbymerchantservice.DepsStatsByMerchant{
		Mencache:      statcachem.NewMerchantStatsByMerchantCache(nil),
		Repository:    merchantstatsbymerchantrepository.NewMerchantStatsByMerchantRepository(db),
		Observability: obs,
	})
}

func newMerchantStatsByApiKeyService(db *mongo.Database, obs observability.TraceLoggerObservability) merchantstatsbyapikeyservice.MerchantStatsByApiKeyService {
	return merchantstatsbyapikeyservice.NewMerchantStatsByApiKeyService(&merchantstatsbyapikeyservice.DepsStatsByApiKey{
		Mencache:      statcachea.NewMerchantStatsByApiKeyCache(nil),
		Repository:    merchantstatsbyapikeyrepository.NewMerchantStatsByApiKeyRepository(db),
		Observability: obs,
	})
}
