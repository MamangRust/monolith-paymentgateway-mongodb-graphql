package mencache

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type RoleQueryCache interface {
	SetCachedRoles(ctx context.Context, req *requests.FindAllRoles, data []*salemongo.Role, total *int)
	GetCachedRoles(ctx context.Context, req *requests.FindAllRoles) ([]*salemongo.Role, *int, bool)

	GetCachedRoleById(ctx context.Context, id int) (*salemongo.Role, bool)
	SetCachedRoleById(ctx context.Context, id int, data *salemongo.Role)

	GetCachedRoleByUserId(ctx context.Context, userId int) ([]*salemongo.Role, bool)
	SetCachedRoleByUserId(ctx context.Context, userId int, data []*salemongo.Role)

	GetCachedRoleActive(ctx context.Context, req *requests.FindAllRoles) ([]*salemongo.Role, *int, bool)
	SetCachedRoleActive(ctx context.Context, req *requests.FindAllRoles, data []*salemongo.Role, total *int)

	GetCachedRoleTrashed(ctx context.Context, req *requests.FindAllRoles) ([]*salemongo.Role, *int, bool)
	SetCachedRoleTrashed(ctx context.Context, req *requests.FindAllRoles, data []*salemongo.Role, total *int)

	GetCachedRoleByName(ctx context.Context, name string) (*salemongo.Role, bool)
	SetCachedRoleByName(ctx context.Context, name string, data *salemongo.Role)
}

type RoleCommandCache interface {
	DeleteCachedRole(ctx context.Context, id int)
}
