package card_cache

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
)

type CardQueryCache interface {
	GetByIdCache(ctx context.Context, cardID int) (*model.APIResponseCard, bool)

	GetByUserIDCache(ctx context.Context, userID int) (*model.APIResponseCard, bool)

	GetByCardNumberCache(ctx context.Context, cardNumber string) (*model.APIResponseCard, bool)

	GetFindAllCache(ctx context.Context, req *model.FindAllCardInput) (*model.APIResponsePaginationCard, bool)

	GetByActiveCache(ctx context.Context, req *model.FindAllCardInput) (*model.APIResponsePaginationCardDeleteAt, bool)

	GetByTrashedCache(ctx context.Context, req *model.FindAllCardInput) (*model.APIResponsePaginationCardDeleteAt, bool)

	SetByIdCache(ctx context.Context, cardID int, data *model.APIResponseCard)

	SetByUserIDCache(ctx context.Context, userID int, data *model.APIResponseCard)

	SetByCardNumberCache(ctx context.Context, cardNumber string, data *model.APIResponseCard)

	SetFindAllCache(ctx context.Context, req *model.FindAllCardInput, data *model.APIResponsePaginationCard)

	SetByActiveCache(ctx context.Context, req *model.FindAllCardInput, data *model.APIResponsePaginationCardDeleteAt)

	SetByTrashedCache(ctx context.Context, req *model.FindAllCardInput, data *model.APIResponsePaginationCardDeleteAt)

	DeleteByIdCache(ctx context.Context, cardID int)
	DeleteByUserIDCache(ctx context.Context, userID int)
	DeleteByCardNumberCache(ctx context.Context, cardNumber string)
}

type CardCommandCache interface {
	DeleteCardCommandCache(ctx context.Context, id int)
}
