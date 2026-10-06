package rolepermission

import "context"

type RolePermission interface {
	ValidateRole(ctx context.Context, userID int) ([]string, error)
	CheckRole(ctx context.Context, userID int, requiredRoles ...string) error
}
