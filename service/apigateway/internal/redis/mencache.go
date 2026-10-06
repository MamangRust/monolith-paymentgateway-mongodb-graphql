package mencache

import (
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	sharedcachehelpers "github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	"go.uber.org/zap"

	"github.com/redis/go-redis/v9"
)

type CacheApiGateway interface {
	MerchantCache
	RoleCache
	GetStore() *sharedcachehelpers.CacheStore
}

type mencacheApiGateay struct {
	MerchantCache
	RoleCache
	store *sharedcachehelpers.CacheStore
}

func (m *mencacheApiGateay) GetStore() *sharedcachehelpers.CacheStore {
	return m.store
}

type Deps struct {
	Redis  *redis.Client
	Logger logger.LoggerInterface
}

func NewCacheApiGateway(deps *Deps) CacheApiGateway {
	metrics, err := observability.NewCacheMetrics("apigateway")
	if err != nil {
		deps.Logger.Error("Failed to initialize cache metrics for apigateway cache store", zap.Error(err))
	}

	store := sharedcachehelpers.NewCacheStore(deps.Redis, deps.Logger, metrics)

	return &mencacheApiGateay{
		MerchantCache: NewMerchantCache(store),
		RoleCache:     NewRoleCache(store),
		store:         store,
	}
}
