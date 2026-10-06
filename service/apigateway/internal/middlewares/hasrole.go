package middlewares

import (
	"context"
	"errors"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	mycontext "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/context"
)

// ErrRoleCheckerNotConfigured is returned when the running schema was built
// without an RBAC checker, so an annotated field can never be authorised.
var ErrRoleCheckerNotConfigured = errors.New("rbac: role checker is not configured")

// RoleChecker verifies that a user owns at least one of the required roles.
// *rolepermission.RolePermission (Kafka + cache) satisfies it.
type RoleChecker interface {
	CheckRole(ctx context.Context, userID int, requiredRoles ...string) error
}

// HasRole implements the GraphQL field directive
// @hasRole(roles: ["Admin", ...]) declared in graphql/directives.graphqls.
//
// Role based access control cannot be enforced as an HTTP middleware on this
// gateway: every operation goes through a single POST /query endpoint, so the
// HTTP path/route tells nothing about which GraphQL operation is being
// executed. The directive runs inside the GraphQL executor instead, once per
// annotated field, where the user stored in the request context by
// AuthMiddleware is available.
//
// The directive only *restricts* authenticated traffic. A request without a
// user in the context passes through: deciding whether anonymous traffic is
// allowed belongs to AuthMiddleware, which whitelists the public operations
// (loginUser/registerUser/refreshToken) and rejects everything else with HTTP
// 401. The one exception is the test harness, which mounts the schema without
// the auth layer and therefore never carries a user.
func HasRole(checker RoleChecker) func(ctx context.Context, obj any, next graphql.Resolver, roles []string) (any, error) {
	return func(ctx context.Context, obj any, next graphql.Resolver, roles []string) (any, error) {
		userID, ok := mycontext.UserForContext(ctx)
		if !ok {
			return next(ctx)
		}

		if checker == nil {
			return nil, ErrRoleCheckerNotConfigured
		}

		if len(roles) > 0 {
			if err := checker.CheckRole(ctx, userID, roles...); err != nil {
				return nil, fmt.Errorf("forbidden: %w", err)
			}
		}

		return next(ctx)
	}
}
