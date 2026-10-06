package repository

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type UserQueryRepository interface {
	FindAllUsers(ctx context.Context, req *requests.FindAllUsers) ([]*salemongo.User, error)
	FindByActive(ctx context.Context, req *requests.FindAllUsers) ([]*salemongo.User, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllUsers) ([]*salemongo.User, error)
	FindById(ctx context.Context, user_id int) (*salemongo.User, error)
	FindByEmail(ctx context.Context, email string) (*salemongo.User, error)
	FindByVerificationCode(ctx context.Context, code string) (*salemongo.User, error)
	CountAllUsers(ctx context.Context, search string) (int64, error)
	CountActiveUsers(ctx context.Context, search string) (int64, error)
	CountTrashedUsers(ctx context.Context, search string) (int64, error)
}

type UserCommandRepository interface {
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
