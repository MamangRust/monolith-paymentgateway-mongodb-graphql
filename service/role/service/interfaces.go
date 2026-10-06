package service

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

// RoleQueryService is an interface for querying role records
type RoleQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllRoles) ([]*salemongo.Role, *int, error)
	FindByActiveRole(ctx context.Context, req *requests.FindAllRoles) ([]*salemongo.Role, *int, error)
	FindByTrashedRole(ctx context.Context, req *requests.FindAllRoles) ([]*salemongo.Role, *int, error)
	FindById(ctx context.Context, role_id int) (*salemongo.Role, error)
	FindByUserId(ctx context.Context, id int) ([]*salemongo.Role, error)
	FindByName(ctx context.Context, name string) (*salemongo.Role, error)
}

// RoleCommandService is an interface for creating, updating, and deleting role records
type RoleCommandService interface {
	CreateRole(ctx context.Context, request *requests.CreateRoleRequest) (*salemongo.Role, error)
	UpdateRole(ctx context.Context, request *requests.UpdateRoleRequest) (*salemongo.Role, error)
	TrashedRole(ctx context.Context, role_id int) (*salemongo.Role, error)
	RestoreRole(ctx context.Context, role_id int) (*salemongo.Role, error)
	DeleteRolePermanent(ctx context.Context, role_id int) (bool, error)

	RestoreAllRole(ctx context.Context) (bool, error)
	DeleteAllRolePermanent(ctx context.Context) (bool, error)
}

// UserRoleService is an interface for (un)assigning roles to a user.
type UserRoleService interface {
	AssignRoleToUser(ctx context.Context, userID, roleID int) (*salemongo.UserRole, error)
	RemoveRoleFromUser(ctx context.Context, userID, roleID int) error
}
