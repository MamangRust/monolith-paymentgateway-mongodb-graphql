package repository

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type RoleQueryRepository interface {
	FindAllRoles(ctx context.Context, req *requests.FindAllRoles) ([]*salemongo.Role, error)
	FindByActiveRole(ctx context.Context, req *requests.FindAllRoles) ([]*salemongo.Role, error)
	FindByTrashedRole(ctx context.Context, req *requests.FindAllRoles) ([]*salemongo.Role, error)
	FindById(ctx context.Context, role_id int) (*salemongo.Role, error)
	FindByName(ctx context.Context, name string) (*salemongo.Role, error)
	FindByUserId(ctx context.Context, user_id int) ([]*salemongo.Role, error)
	CountAllRoles(ctx context.Context, search string) (int64, error)
	CountActiveRoles(ctx context.Context, search string) (int64, error)
	CountTrashedRoles(ctx context.Context, search string) (int64, error)
}

type RoleCommandRepository interface {
	CreateRole(ctx context.Context, request *requests.CreateRoleRequest) (*salemongo.Role, error)
	UpdateRole(ctx context.Context, request *requests.UpdateRoleRequest) (*salemongo.Role, error)
	TrashedRole(ctx context.Context, role_id int) (*salemongo.Role, error)
	RestoreRole(ctx context.Context, role_id int) (*salemongo.Role, error)
	DeleteRolePermanent(ctx context.Context, role_id int) (bool, error)
	RestoreAllRole(ctx context.Context) (bool, error)
	DeleteAllRolePermanent(ctx context.Context) (bool, error)
}
