package repository

import (
	pbrole "github.com/MamangRust/monolith-payment-gateway-pb/role"
	pbuserrole "github.com/MamangRust/monolith-payment-gateway-pb/user_role"
	adapter "github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	roleadapter "github.com/MamangRust/monolith-payment-gateway-pkg/adapter/role"
	"go.mongodb.org/mongo-driver/mongo"
)

// GuardOptions collects the dependency guards applied to the role adapters
// built inside NewRepositories. The role query and user_role command sides get
// separate breakers so one failure mode cannot starve the other.
type GuardOptions struct {
	Role     []adapter.GuardOption
	UserRole []adapter.GuardOption
}

type Repositories interface {
	UserQuery() UserQueryRepository
	UserCommand() UserCommandRepository
	Role() roleadapter.QueryRepository
	UserRole() roleadapter.CommandRepository
}

type repositories struct {
	userQuery   UserQueryRepository
	userCommand UserCommandRepository
	role        roleadapter.QueryRepository
	userRole    roleadapter.CommandRepository
}

// NewRepositories assembles the user Mongo repositories and builds the shared
// role/user_role gRPC adapters from the raw clients under their guards. Role
// reads and user_roles writes go through the role service gRPC API; the user
// service never touches the roles/user_roles collections directly.
func NewRepositories(
	db *mongo.Database,
	roleQueryClient pbrole.RoleQueryServiceClient,
	userRoleClient pbuserrole.UserRoleServiceClient,
	guards ...GuardOptions,
) Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	return &repositories{
		userCommand: NewUserCommandRepository(db.Collection("users")),
		userQuery:   NewUserQueryRepository(db.Collection("users")),
		role:        roleadapter.NewRoleAdapter(roleQueryClient, g.Role...),
		userRole:    roleadapter.NewUserRoleAdapter(userRoleClient, g.UserRole...),
	}
}

func (r *repositories) UserQuery() UserQueryRepository {
	return r.userQuery
}
func (r *repositories) UserCommand() UserCommandRepository {
	return r.userCommand
}
func (r *repositories) Role() roleadapter.QueryRepository {
	return r.role
}
func (r *repositories) UserRole() roleadapter.CommandRepository {
	return r.userRole
}
