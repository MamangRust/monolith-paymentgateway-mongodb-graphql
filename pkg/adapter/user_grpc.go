package adapter

import (
	"context"

	pbuser "github.com/MamangRust/monolith-payment-gateway-pb/user"
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/resilience"
)

type UserAdapter interface {
	FindById(ctx context.Context, userID int) (*salemongo.User, error)
}

type userGRPCAdapter struct {
	queryClient pbuser.UserQueryServiceClient
	guard       *resilience.DependencyGuard
}

func (a *userGRPCAdapter) SetGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

func NewUserAdapter(queryClient pbuser.UserQueryServiceClient, opts ...GuardOption) UserAdapter {
	a := &userGRPCAdapter{
		queryClient: queryClient,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *userGRPCAdapter) FindById(ctx context.Context, userID int) (*salemongo.User, error) {
	var resp *pbuser.ApiResponseUser
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.queryClient.FindById(callCtx, &pbuser.FindByIdUserRequest{
			Id: int32(userID),
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Data == nil {
		return nil, nil
	}

	return &salemongo.User{
		UserID:    resp.Data.Id,
		Email:     resp.Data.Email,
		Firstname: resp.Data.Firstname,
		Lastname:  resp.Data.Lastname,
	}, nil
}
