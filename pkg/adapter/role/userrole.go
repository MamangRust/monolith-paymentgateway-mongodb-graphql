package role

import (
	"context"

	pbuserrole "github.com/MamangRust/monolith-payment-gateway-pb/user_role"
	adapter "github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/resilience"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	userrole_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/user_role_errors/repository"
)

// CommandRepository is the contract consumers depend on for user_roles writes.
type CommandRepository interface {
	AssignRoleToUser(ctx context.Context, req *requests.CreateUserRoleRequest) (*salemongo.UserRole, error)
	RemoveRoleFromUser(ctx context.Context, req *requests.RemoveUserRoleRequest) error
}

// UserRoleAdapter implements CommandRepository on top of the generated
// user_role service client. The user_roles collection stays owned by the role
// service; consumers reach it only through this adapter.
type UserRoleAdapter struct {
	client pbuserrole.UserRoleServiceClient
	guard  *resilience.DependencyGuard
}

// NewUserRoleAdapter wraps the generated user_role client. Passing zero
// options leaves the guard nil, which makes DependencyGuard.Call a plain
// passthrough.
func NewUserRoleAdapter(client pbuserrole.UserRoleServiceClient, opts ...adapter.GuardOption) *UserRoleAdapter {
	r := &UserRoleAdapter{client: client}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// SetGuard implements adapter.GuardSetter.
func (r *UserRoleAdapter) SetGuard(g *resilience.DependencyGuard) {
	r.guard = g
}

// AssignRoleToUser implements CommandRepository. The role service treats a
// repeated assignment as a no-op and returns the existing mapping, so this call
// is idempotent.
func (r *UserRoleAdapter) AssignRoleToUser(ctx context.Context, req *requests.CreateUserRoleRequest) (*salemongo.UserRole, error) {
	var res *pbuserrole.ApiResponseUserRole
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = r.client.AssignRoleToUser(ctx, &pbuserrole.AssignRoleToUserRequest{
			UserId: int32(req.UserId),
			RoleId: int32(req.RoleId),
		})
		return callErr
	})
	if err != nil || res == nil || res.Data == nil {
		return nil, userrole_errors.ErrAssignRoleToUser.WithInternal(err)
	}
	return &salemongo.UserRole{
		UserRoleID: res.Data.UserRoleId,
		UserID:     res.Data.UserId,
		RoleID:     res.Data.RoleId,
	}, nil
}

// RemoveRoleFromUser implements CommandRepository.
func (r *UserRoleAdapter) RemoveRoleFromUser(ctx context.Context, req *requests.RemoveUserRoleRequest) error {
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		_, callErr := r.client.RemoveRoleFromUser(ctx, &pbuserrole.RemoveRoleFromUserRequest{
			UserId: int32(req.UserId),
			RoleId: int32(req.RoleId),
		})
		return callErr
	})
	if err != nil {
		return userrole_errors.ErrRemoveRole.WithInternal(err)
	}
	return nil
}
