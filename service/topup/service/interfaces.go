package service

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-topup/repository"
)

// TopupQueryService defines the read-only operations for querying topup data.
type TopupQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllTopups) ([]*repository.TopupQueryResult, *int, error)
	FindAllByCardNumber(ctx context.Context, req *requests.FindAllTopupsByCardNumber) ([]*repository.TopupByCardResult, *int, error)
	FindById(ctx context.Context, topupID int) (*salemongo.Topup, error)
	FindByActive(ctx context.Context, req *requests.FindAllTopups) ([]*repository.TopupQueryResult, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllTopups) ([]*repository.TopupQueryResult, *int, error)
}

type TopupCommandService interface {
	CreateTopup(ctx context.Context, request *requests.CreateTopupRequest) (*salemongo.Topup, error)
	UpdateTopup(ctx context.Context, request *requests.UpdateTopupRequest) (*salemongo.Topup, error)
	TrashedTopup(ctx context.Context, topup_id int) (*salemongo.Topup, error)
	RestoreTopup(ctx context.Context, topup_id int) (*salemongo.Topup, error)
	DeleteTopupPermanent(ctx context.Context, topup_id int) (bool, error)

	RestoreAllTopup(ctx context.Context) (bool, error)
	DeleteAllTopupPermanent(ctx context.Context) (bool, error)
}
