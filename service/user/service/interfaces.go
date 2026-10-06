package service

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

// UserQueryService handles query operations related to user data.
type UserQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllUsers) ([]*salemongo.User, *int, error)
	FindByID(ctx context.Context, id int) (*salemongo.User, error)
	FindByEmail(ctx context.Context, email string) (*salemongo.User, error)
	FindByVerificationCode(ctx context.Context, verificationCode string) (*salemongo.User, error)
	FindByActive(ctx context.Context, req *requests.FindAllUsers) ([]*salemongo.User, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllUsers) ([]*salemongo.User, *int, error)
}

// UserCommandService handles command operations related to user management.
type UserCommandService interface {
	CreateUser(ctx context.Context, request *requests.CreateUserRequest) (*salemongo.User, error)
	UpdateUser(ctx context.Context, request *requests.UpdateUserRequest) (*salemongo.User, error)
	UpdateIsVerified(ctx context.Context, userID int, isVerified bool) (*salemongo.User, error)
	UpdatePassword(ctx context.Context, userID int, password string) (*salemongo.User, error)
	TrashedUser(ctx context.Context, user_id int) (*salemongo.User, error)
	RestoreUser(ctx context.Context, user_id int) (*salemongo.User, error)
	DeleteUserPermanent(ctx context.Context, user_id int) (bool, error)

	RestoreAllUser(ctx context.Context) (bool, error)
	DeleteAllUserPermanent(ctx context.Context) (bool, error)
}
