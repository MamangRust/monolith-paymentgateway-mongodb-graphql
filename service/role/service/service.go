package service

import (
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	mencache "github.com/MamangRust/monolith-payment-gateway-role/redis"
	"github.com/MamangRust/monolith-payment-gateway-role/repository"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
)

// Service aggregates role-related services.
type Service struct {
	RoleQuery   RoleQueryService
	RoleCommand RoleCommandService
	UserRole    UserRoleService
}

// Deps defines dependencies for role services.
type Deps struct {
	Cache        *cache.CacheStore
	Repositories repository.Repositories
	Logger       logger.LoggerInterface
}

// NewService creates a new role Service.
func NewService(deps *Deps) *Service {
	obs, _ := observability.NewObservability("role-service", deps.Logger)
	cache := mencache.NewMencache(deps.Cache)

	return &Service{
		RoleQuery:   newRoleQueryService(deps, obs, cache.RoleQueryCache),
		RoleCommand: newRoleCommandService(deps, obs, cache.RoleCommandCache),
		UserRole:    newUserRoleService(deps, obs),
	}
}

// newUserRoleService creates a UserRoleService.
func newUserRoleService(
	deps *Deps,
	obs observability.TraceLoggerObservability,
) UserRoleService {
	return NewUserRoleService(&userRoleDeps{
		Repository:    deps.Repositories,
		Logger:        deps.Logger,
		Observability: obs,
	})
}

// newRoleCommandService creates a RoleCommandService.
func newRoleCommandService(
	deps *Deps,
	obs observability.TraceLoggerObservability,
	cache mencache.RoleCommandCache,
) RoleCommandService {
	return NewRoleCommandService(&roleCommandDeps{
		Cache:         cache,
		Repository:    deps.Repositories,
		Logger:        deps.Logger,
		Observability: obs,
	})
}

// newRoleQueryService creates a RoleQueryService.
func newRoleQueryService(
	deps *Deps,
	obs observability.TraceLoggerObservability,
	cache mencache.RoleQueryCache,
) RoleQueryService {
	return NewRoleQueryService(&roleQueryDeps{
		Cache:         cache,
		Repository:    deps.Repositories,
		Logger:        deps.Logger,
		Observability: obs,
	})
}
