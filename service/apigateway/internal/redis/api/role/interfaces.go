package role_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
)

type RoleQueryCache interface {
	SetCachedRoles(ctx context.Context, req *model.FindAllRoleInput, data *model.APIResponsePaginationRole)
	GetCachedRoles(ctx context.Context, req *model.FindAllRoleInput) (*model.APIResponsePaginationRole, bool)

	SetCachedRoleById(ctx context.Context, id int, data *model.APIResponseRole)
	GetCachedRoleById(ctx context.Context, id int) (*model.APIResponseRole, bool)

	SetCachedRoleByUserId(ctx context.Context, userId int, data *model.APIResponsesRole)
	GetCachedRoleByUserId(ctx context.Context, userId int) (*model.APIResponsesRole, bool)

	SetCachedRoleActive(ctx context.Context, req *model.FindAllRoleInput, data *model.APIResponsePaginationRoleDeleteAt)
	GetCachedRoleActive(ctx context.Context, req *model.FindAllRoleInput) (*model.APIResponsePaginationRoleDeleteAt, bool)

	SetCachedRoleTrashed(ctx context.Context, req *model.FindAllRoleInput, data *model.APIResponsePaginationRoleDeleteAt)
	GetCachedRoleTrashed(ctx context.Context, req *model.FindAllRoleInput) (*model.APIResponsePaginationRoleDeleteAt, bool)
}

type RoleCommandCache interface {
	DeleteCachedRole(ctx context.Context, id int)
}
