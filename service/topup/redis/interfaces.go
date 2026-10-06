package mencache

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-topup/repository"
)

type TopupQueryCache interface {
	GetCachedTopupsCache(ctx context.Context, req *requests.FindAllTopups) ([]*repository.TopupQueryResult, *int, bool)
	SetCachedTopupsCache(ctx context.Context, req *requests.FindAllTopups, data []*repository.TopupQueryResult, total *int)

	GetCacheTopupByCardCache(ctx context.Context, req *requests.FindAllTopupsByCardNumber) ([]*repository.TopupByCardResult, *int, bool)
	SetCacheTopupByCardCache(ctx context.Context, req *requests.FindAllTopupsByCardNumber, data []*repository.TopupByCardResult, total *int)

	GetCachedTopupActiveCache(ctx context.Context, req *requests.FindAllTopups) ([]*repository.TopupQueryResult, *int, bool)
	SetCachedTopupActiveCache(ctx context.Context, req *requests.FindAllTopups, data []*repository.TopupQueryResult, total *int)

	GetCachedTopupTrashedCache(ctx context.Context, req *requests.FindAllTopups) ([]*repository.TopupQueryResult, *int, bool)
	SetCachedTopupTrashedCache(ctx context.Context, req *requests.FindAllTopups, data []*repository.TopupQueryResult, total *int)

	GetCachedTopupCache(ctx context.Context, id int) (*salemongo.Topup, bool)
	SetCachedTopupCache(ctx context.Context, data *salemongo.Topup)
}

type TopupCommandCache interface {
	DeleteCachedTopupCache(ctx context.Context, id int)
}
