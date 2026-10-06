package repository

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type UserRepository interface {
	FindById(ctx context.Context, user_id int) (*salemongo.User, error)
}

type MerchantQueryRepository interface {
	FindAllMerchants(ctx context.Context, req *requests.FindAllMerchants) ([]*salemongo.Merchant, error)
	FindByActive(ctx context.Context, req *requests.FindAllMerchants) ([]*salemongo.Merchant, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllMerchants) ([]*salemongo.Merchant, error)
	FindByApiKey(ctx context.Context, api_key string) (*salemongo.Merchant, error)
	FindByMerchantId(ctx context.Context, merchant_id int) (*salemongo.Merchant, error)
	FindByName(ctx context.Context, name string) (*salemongo.Merchant, error)
	FindByMerchantUserId(ctx context.Context, user_id int) ([]*salemongo.Merchant, error)
	CountAllMerchants(ctx context.Context, search string) (int64, error)
	CountActiveMerchants(ctx context.Context, search string) (int64, error)
	CountTrashedMerchants(ctx context.Context, search string) (int64, error)
}

type MerchantDocumentQueryRepository interface {
	FindAllDocuments(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*salemongo.MerchantDocument, error)
	FindByIdDocument(ctx context.Context, id int) (*salemongo.MerchantDocument, error)
	FindByActiveDocuments(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*salemongo.MerchantDocument, error)
	FindByTrashedDocuments(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*salemongo.MerchantDocument, error)
	CountAllDocuments(ctx context.Context, search string) (int64, error)
	CountActiveDocuments(ctx context.Context, search string) (int64, error)
	CountTrashedDocuments(ctx context.Context, search string) (int64, error)
}

type MerchantTransactionRepository interface {
	FindAllTransactions(ctx context.Context, req *requests.FindAllMerchantTransactions) ([]*salemongo.Transaction, error)
	FindAllTransactionsByMerchant(ctx context.Context, req *requests.FindAllMerchantTransactionsById) ([]*salemongo.Transaction, error)
	FindAllTransactionsByApikey(ctx context.Context, req *requests.FindAllMerchantTransactionsByApiKey) ([]*salemongo.Transaction, error)
	CountAllTransactions(ctx context.Context, search string) (int64, error)
	CountTransactionsByMerchant(ctx context.Context, merchantID int, search string) (int64, error)
	CountTransactionsByApikey(ctx context.Context, apiKey, search string) (int64, error)
}

type MerchantCommandRepository interface {
	CreateMerchant(ctx context.Context, request *requests.CreateMerchantRequest) (*salemongo.Merchant, error)
	UpdateMerchant(ctx context.Context, request *requests.UpdateMerchantRequest) (*salemongo.Merchant, error)
	UpdateMerchantStatus(ctx context.Context, request *requests.UpdateMerchantStatusRequest) (*salemongo.Merchant, error)

	TrashedMerchant(ctx context.Context, merchantId int) (*salemongo.Merchant, error)
	RestoreMerchant(ctx context.Context, merchantId int) (*salemongo.Merchant, error)
	DeleteMerchantPermanent(ctx context.Context, merchantId int) (bool, error)

	RestoreAllMerchant(ctx context.Context) (bool, error)
	DeleteAllMerchantPermanent(ctx context.Context) (bool, error)
}

type MerchantDocumentCommandRepository interface {
	CreateMerchantDocument(ctx context.Context, request *requests.CreateMerchantDocumentRequest) (*salemongo.MerchantDocument, error)
	UpdateMerchantDocument(ctx context.Context, request *requests.UpdateMerchantDocumentRequest) (*salemongo.MerchantDocument, error)
	UpdateMerchantDocumentStatus(ctx context.Context, request *requests.UpdateMerchantDocumentStatusRequest) (*salemongo.MerchantDocument, error)
	TrashedMerchantDocument(ctx context.Context, merchant_document_id int) (*salemongo.MerchantDocument, error)
	RestoreMerchantDocument(ctx context.Context, merchant_document_id int) (*salemongo.MerchantDocument, error)
	DeleteMerchantDocumentPermanent(ctx context.Context, merchant_document_id int) (bool, error)
	RestoreAllMerchantDocument(ctx context.Context) (bool, error)
	DeleteAllMerchantDocumentPermanent(ctx context.Context) (bool, error)
}
