package mencache

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type CardQueryCache interface {
	GetByIdCache(ctx context.Context, cardID int) (*salemongo.Card, bool)
	GetByUserIDCache(ctx context.Context, userID int) (*salemongo.Card, bool)
	GetByCardNumberCache(ctx context.Context, cardNumber string) (*salemongo.Card, bool)
	GetUserCardByCardNumberCache(ctx context.Context, cardNumber string) (*salemongo.Card, bool)
	GetFindAllCache(ctx context.Context, req *requests.FindAllCards) ([]*salemongo.Card, *int, bool)
	GetByActiveCache(ctx context.Context, req *requests.FindAllCards) ([]*salemongo.Card, *int, bool)
	GetByTrashedCache(ctx context.Context, req *requests.FindAllCards) ([]*salemongo.Card, *int, bool)
	SetByIdCache(ctx context.Context, cardID int, data *salemongo.Card)
	SetByUserIDCache(ctx context.Context, userID int, data *salemongo.Card)
	SetByCardNumberCache(ctx context.Context, cardNumber string, data *salemongo.Card)
	SetFindAllCache(ctx context.Context, req *requests.FindAllCards, data []*salemongo.Card, totalRecords *int)
	SetByActiveCache(ctx context.Context, req *requests.FindAllCards, data []*salemongo.Card, totalRecords *int)
	SetUserCardByCardNumberCache(ctx context.Context, cardNumber string, data *salemongo.Card)
	SetByTrashedCache(ctx context.Context, req *requests.FindAllCards, data []*salemongo.Card, totalRecords *int)
	DeleteByIdCache(ctx context.Context, cardID int)
	DeleteByUserIDCache(ctx context.Context, userID int)
	DeleteByCardNumberCache(ctx context.Context, cardNumber string)
}

type CardCommandCache interface {
	DeleteCardCommandCache(ctx context.Context, id int)
}
