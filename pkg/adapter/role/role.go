// Package role adapts the Role service gRPC API into the shared domain model.
// It exposes two narrow adapters: RoleAdapter for role reads and UserRoleAdapter
// for user_roles writes.
//
// It lives in its own package (instead of the flat pkg/adapter package) so that
// service/user/repository can import it without pulling in the other adapters
// under pkg/adapter (import cycle).
package role

import (
	"context"

	pbrole "github.com/MamangRust/monolith-payment-gateway-pb/role"
	adapter "github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/resilience"
	role_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/role_errors/repository"
)

// QueryRepository is the contract consumers depend on for role reads.
type QueryRepository interface {
	FindById(ctx context.Context, roleID int) (*salemongo.Role, error)
	FindByName(ctx context.Context, name string) (*salemongo.Role, error)
}

// RoleAdapter implements QueryRepository on top of the generated role query
// client.
type RoleAdapter struct {
	client pbrole.RoleQueryServiceClient
	guard  *resilience.DependencyGuard
}

// NewRoleAdapter wraps the generated role query client. Passing zero options
// leaves the guard nil, which makes DependencyGuard.Call a plain passthrough.
func NewRoleAdapter(client pbrole.RoleQueryServiceClient, opts ...adapter.GuardOption) *RoleAdapter {
	r := &RoleAdapter{client: client}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// SetGuard implements adapter.GuardSetter.
func (r *RoleAdapter) SetGuard(g *resilience.DependencyGuard) {
	r.guard = g
}

func (r *RoleAdapter) FindById(ctx context.Context, roleID int) (*salemongo.Role, error) {
	var res *pbrole.ApiResponseRole
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = r.client.FindByIdRole(ctx, &pbrole.FindByIdRoleRequest{RoleId: int32(roleID)})
		return callErr
	})
	if err != nil || res == nil || res.Data == nil {
		return nil, role_errors.ErrRoleNotFound.WithInternal(err)
	}
	return &salemongo.Role{RoleID: res.Data.Id, RoleName: res.Data.Name}, nil
}

func (r *RoleAdapter) FindByName(ctx context.Context, name string) (*salemongo.Role, error) {
	var res *pbrole.ApiResponseRole
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = r.client.FindByNameRole(ctx, &pbrole.FindByNameRoleRequest{Name: name})
		return callErr
	})
	if err != nil || res == nil || res.Data == nil {
		return nil, role_errors.ErrRoleNotFound.WithInternal(err)
	}
	return &salemongo.Role{RoleID: res.Data.Id, RoleName: res.Data.Name}, nil
}
