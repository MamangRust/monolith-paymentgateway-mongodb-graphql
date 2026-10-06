package repository

import (
	pb_role "github.com/MamangRust/monolith-payment-gateway-pb/role"
	pbuserrole "github.com/MamangRust/monolith-payment-gateway-pb/user_role"
	adapter "github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	roleadapter "github.com/MamangRust/monolith-payment-gateway-pkg/adapter/role"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repositories struct {
	User         UserRepository
	RefreshToken RefreshTokenRepository
	UserRole     UserRoleRepository
	Role         RoleRepository
	ResetToken   ResetTokenRepository
}

// RepositoriesDeps carries the Mongo database plus the raw role-service gRPC
// clients. The guarded role adapters are constructed inside NewRepositories so
// every outbound call runs under a per-call deadline, bulkhead and circuit
// breaker.
type RepositoriesDeps struct {
	DB              *mongo.Database
	RoleQueryClient pb_role.RoleQueryServiceClient
	UserRoleClient  pbuserrole.UserRoleServiceClient
}

// GuardOptions collects the dependency guards applied to the adapters built
// inside NewRepositories. The role query and user_role command sides get
// separate breakers so one failure mode cannot starve the other.
type GuardOptions struct {
	Role     []adapter.GuardOption
	UserRole []adapter.GuardOption
}

// NewRepositories assembles the auth repositories: Mongo-backed user/token
// repos plus the shared role and user_role gRPC adapters, constructed here
// under their guards.
func NewRepositories(deps *RepositoriesDeps, guards ...GuardOptions) *Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	return &Repositories{
		User:         NewUserRepository(deps.DB),
		RefreshToken: NewRefreshTokenRepository(deps.DB),
		UserRole:     roleadapter.NewUserRoleAdapter(deps.UserRoleClient, g.UserRole...),
		Role:         roleadapter.NewRoleAdapter(deps.RoleQueryClient, g.Role...),
		ResetToken:   NewResetTokenRepository(deps.DB),
	}
}
