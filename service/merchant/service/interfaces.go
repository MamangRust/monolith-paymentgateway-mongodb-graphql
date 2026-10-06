package service

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type MerchantQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllMerchants) ([]*salemongo.Merchant, *int, error)
	FindById(ctx context.Context, merchant_id int) (*salemongo.Merchant, error)
	FindByActive(ctx context.Context, req *requests.FindAllMerchants) ([]*salemongo.Merchant, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllMerchants) ([]*salemongo.Merchant, *int, error)
	FindByApiKey(ctx context.Context, api_key string) (*salemongo.Merchant, error)
	FindByMerchantUserId(ctx context.Context, user_id int) ([]*salemongo.Merchant, error)
}

type MerchantDocumentQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*salemongo.MerchantDocument, *int, error)
	FindByActive(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*salemongo.MerchantDocument, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*salemongo.MerchantDocument, *int, error)
	FindById(ctx context.Context, document_id int) (*salemongo.MerchantDocument, error)
}

type MerchantTransactionService interface {
	FindAllTransactions(ctx context.Context, req *requests.FindAllMerchantTransactions) ([]*salemongo.Transaction, *int, error)
	FindAllTransactionsByApikey(ctx context.Context, req *requests.FindAllMerchantTransactionsByApiKey) ([]*salemongo.Transaction, *int, error)
	FindAllTransactionsByMerchant(ctx context.Context, req *requests.FindAllMerchantTransactionsById) ([]*salemongo.Transaction, *int, error)
}

type MerchantCommandService interface {
	CreateMerchant(ctx context.Context, request *requests.CreateMerchantRequest) (*salemongo.Merchant, error)
	UpdateMerchant(ctx context.Context, request *requests.UpdateMerchantRequest) (*salemongo.Merchant, error)
	UpdateMerchantStatus(ctx context.Context, request *requests.UpdateMerchantStatusRequest) (*salemongo.Merchant, error)
	TrashedMerchant(ctx context.Context, merchant_id int) (*salemongo.Merchant, error)
	RestoreMerchant(ctx context.Context, merchant_id int) (*salemongo.Merchant, error)
	DeleteMerchantPermanent(ctx context.Context, merchant_id int) (bool, error)
	RestoreAllMerchant(ctx context.Context) (bool, error)
	DeleteAllMerchantPermanent(ctx context.Context) (bool, error)
}

type MerchantDocumentCommandService interface {
	CreateMerchantDocument(ctx context.Context, request *requests.CreateMerchantDocumentRequest) (*salemongo.MerchantDocument, error)
	UpdateMerchantDocument(ctx context.Context, request *requests.UpdateMerchantDocumentRequest) (*salemongo.MerchantDocument, error)
	UpdateMerchantDocumentStatus(ctx context.Context, request *requests.UpdateMerchantDocumentStatusRequest) (*salemongo.MerchantDocument, error)
	TrashedMerchantDocument(ctx context.Context, document_id int) (*salemongo.MerchantDocument, error)
	RestoreMerchantDocument(ctx context.Context, document_id int) (*salemongo.MerchantDocument, error)
	DeleteMerchantDocumentPermanent(ctx context.Context, document_id int) (bool, error)
	RestoreAllMerchantDocument(ctx context.Context) (bool, error)
	DeleteAllMerchantDocumentPermanent(ctx context.Context) (bool, error)
}
