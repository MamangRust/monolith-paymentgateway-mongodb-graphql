package repository

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

// UserRepository defines the data access layer for user-related operations.
// This uses gRPC adapters, not direct DB access.
//
//go:generate mockgen -source=interfaces.go -destination=mocks/mock.go
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*salemongo.User, error)

	FindByEmailAndVerify(ctx context.Context, email string) (*salemongo.User, error)

	FindById(ctx context.Context, user_id int) (*salemongo.User, error)

	CreateUser(ctx context.Context, request *requests.RegisterRequest) (*salemongo.User, error)

	UpdateUserIsVerified(ctx context.Context, user_id int, is_verified bool) (*salemongo.User, error)

	UpdateUserPassword(ctx context.Context, user_id int, password string) (*salemongo.User, error)

	FindByVerificationCode(ctx context.Context, verification_code string) (*salemongo.User, error)
}

type ResetTokenRepository interface {
	FindByToken(ctx context.Context, code string) (*salemongo.ResetToken, error)

	CreateResetToken(ctx context.Context, req *requests.CreateResetTokenRequest) (*salemongo.ResetToken, error)

	DeleteResetToken(ctx context.Context, user_id int32) error
}

type RefreshTokenRepository interface {
	FindByToken(ctx context.Context, token string) (*salemongo.RefreshToken, error)

	FindByUserId(ctx context.Context, user_id int) (*salemongo.RefreshToken, error)

	CreateRefreshToken(ctx context.Context, req *requests.CreateRefreshToken) (*salemongo.RefreshToken, error)

	UpdateRefreshToken(ctx context.Context, req *requests.UpdateRefreshToken) (*salemongo.RefreshToken, error)

	DeleteRefreshToken(ctx context.Context, token string) error

	DeleteRefreshTokenByUserId(ctx context.Context, user_id int) error
}

type UserRoleRepository interface {
	AssignRoleToUser(ctx context.Context, req *requests.CreateUserRoleRequest) (*salemongo.UserRole, error)

	RemoveRoleFromUser(ctx context.Context, req *requests.RemoveUserRoleRequest) error
}

type RoleRepository interface {
	FindById(ctx context.Context, id int) (*salemongo.Role, error)

	FindByName(ctx context.Context, name string) (*salemongo.Role, error)
}
