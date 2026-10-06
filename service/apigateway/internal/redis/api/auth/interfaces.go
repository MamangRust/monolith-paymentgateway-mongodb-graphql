package auth_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
)

type IdentityCache interface {
	SetRefreshToken(
		ctx context.Context,
		token string,
		response *model.APIResponseRefreshToken,
	)
	GetRefreshToken(ctx context.Context, token string) (*model.APIResponseRefreshToken, bool)
	DeleteRefreshToken(ctx context.Context, token string)

	SetCachedUserInfo(ctx context.Context, userId int, data *model.APIResponseGetMe)
	GetCachedUserInfo(ctx context.Context, userId int) (*model.APIResponseGetMe, bool)
	DeleteCachedUserInfo(ctx context.Context, userId int)
}

type LoginCache interface {
	GetCachedLogin(
		ctx context.Context,
		email string,
	) (*model.APIResponseLogin, bool)
	SetCachedLogin(
		ctx context.Context,
		email string,
		data *model.APIResponseLogin,
	)
}
