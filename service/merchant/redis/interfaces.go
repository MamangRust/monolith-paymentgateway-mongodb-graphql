package mencache

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type MerchantQueryCache interface {
	GetCachedMerchants(ctx context.Context, req *requests.FindAllMerchants) ([]*salemongo.Merchant, *int, bool)
	SetCachedMerchants(ctx context.Context, req *requests.FindAllMerchants, data []*salemongo.Merchant, total *int)

	GetCachedMerchantActive(ctx context.Context, req *requests.FindAllMerchants) ([]*salemongo.Merchant, *int, bool)
	SetCachedMerchantActive(ctx context.Context, req *requests.FindAllMerchants, data []*salemongo.Merchant, total *int)

	GetCachedMerchantTrashed(ctx context.Context, req *requests.FindAllMerchants) ([]*salemongo.Merchant, *int, bool)
	SetCachedMerchantTrashed(ctx context.Context, req *requests.FindAllMerchants, data []*salemongo.Merchant, total *int)

	GetCachedMerchant(ctx context.Context, id int) (*salemongo.Merchant, bool)
	SetCachedMerchant(ctx context.Context, data *salemongo.Merchant)

	GetCachedMerchantsByUserId(ctx context.Context, userId int) ([]*salemongo.Merchant, bool)
	SetCachedMerchantsByUserId(ctx context.Context, userId int, data []*salemongo.Merchant)

	GetCachedMerchantByApiKey(ctx context.Context, apiKey string) (*salemongo.Merchant, bool)
	SetCachedMerchantByApiKey(ctx context.Context, apiKey string, data *salemongo.Merchant)
}

type MerchantDocumentQueryCache interface {
	GetCachedMerchantDocuments(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*salemongo.MerchantDocument, *int, bool)
	SetCachedMerchantDocuments(ctx context.Context, req *requests.FindAllMerchantDocuments, data []*salemongo.MerchantDocument, total *int)

	GetCachedMerchantDocumentsActive(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*salemongo.MerchantDocument, *int, bool)
	SetCachedMerchantDocumentsActive(ctx context.Context, req *requests.FindAllMerchantDocuments, data []*salemongo.MerchantDocument, total *int)

	GetCachedMerchantDocumentsTrashed(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*salemongo.MerchantDocument, *int, bool)
	SetCachedMerchantDocumentsTrashed(ctx context.Context, req *requests.FindAllMerchantDocuments, data []*salemongo.MerchantDocument, total *int)

	GetCachedMerchantDocument(ctx context.Context, id int) (*salemongo.MerchantDocument, bool)
	SetCachedMerchantDocument(ctx context.Context, id int, data *salemongo.MerchantDocument)
}

type MerchantCommandCache interface {
	DeleteCachedMerchant(ctx context.Context, id int)
	InvalidateMerchantListCaches(ctx context.Context)
	DeleteCachedMerchantByUserID(ctx context.Context, userId int)
	DeleteCachedMerchantByAPIKey(ctx context.Context, apiKey string)
}

type MerchantDocumentCommandCache interface {
	DeleteCachedMerchantDocuments(ctx context.Context, id int)
}

type MerchantTransactionCache interface {
	GetCacheAllMerchantTransactions(ctx context.Context, req *requests.FindAllMerchantTransactions) ([]*salemongo.Transaction, *int, bool)
	SetCacheAllMerchantTransactions(ctx context.Context, req *requests.FindAllMerchantTransactions, data []*salemongo.Transaction, total *int)

	GetCacheMerchantTransactions(ctx context.Context, req *requests.FindAllMerchantTransactionsById) ([]*salemongo.Transaction, *int, bool)
	SetCacheMerchantTransactions(ctx context.Context, req *requests.FindAllMerchantTransactionsById, data []*salemongo.Transaction, total *int)

	GetCacheMerchantTransactionApikey(ctx context.Context, req *requests.FindAllMerchantTransactionsByApiKey) ([]*salemongo.Transaction, *int, bool)
	SetCacheMerchantTransactionApikey(ctx context.Context, req *requests.FindAllMerchantTransactionsByApiKey, data []*salemongo.Transaction, total *int)
}
