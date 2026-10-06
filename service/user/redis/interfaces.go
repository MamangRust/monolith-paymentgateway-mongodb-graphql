package mencache

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type UserQueryCache interface {
	GetCachedUsersCache(ctx context.Context, req *requests.FindAllUsers) ([]*salemongo.User, *int, bool)
	SetCachedUsersCache(ctx context.Context, req *requests.FindAllUsers, data []*salemongo.User, total *int)

	GetCachedUserActiveCache(ctx context.Context, req *requests.FindAllUsers) ([]*salemongo.User, *int, bool)
	SetCachedUserActiveCache(ctx context.Context, req *requests.FindAllUsers, data []*salemongo.User, total *int)

	GetCachedUserTrashedCache(ctx context.Context, req *requests.FindAllUsers) ([]*salemongo.User, *int, bool)
	SetCachedUserTrashedCache(ctx context.Context, req *requests.FindAllUsers, data []*salemongo.User, total *int)

	GetCachedUserCache(ctx context.Context, id int) (*salemongo.User, bool)
	SetCachedUserCache(ctx context.Context, data *salemongo.User)
}

type UserCommandCache interface {
	DeleteUserCache(ctx context.Context, id int)
}
