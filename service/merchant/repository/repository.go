package repository

import (
	pb_user "github.com/MamangRust/monolith-payment-gateway-pb/user"
	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repositories interface {
	MerchantQueryRepository
	MerchantCommandRepository
	MerchantDocumentQueryRepository
	MerchantDocumentCommandRepository
	MerchantTransactionRepository
	UserRepository
	UserAdapter() adapter.UserAdapter
}

type repositories struct {
	MerchantQueryRepository
	MerchantCommandRepository
	MerchantDocumentQueryRepository
	MerchantDocumentCommandRepository
	MerchantTransactionRepository
	UserRepository
	userAdapter adapter.UserAdapter
}

// GuardOptions collects the dependency guards for each remote dependency. The
// options are applied to the adapter built inside NewRepositories, so every
// outbound gRPC call runs under a per-call deadline, bulkhead and circuit
// breaker.
type GuardOptions struct {
	User []adapter.GuardOption
}

// NewRepositories assembles the merchant repositories: Mongo-backed merchant
// and document repos plus the user gRPC adapter, constructed here under its
// guard.
func NewRepositories(db *mongo.Database, userClient pb_user.UserQueryServiceClient, guards ...GuardOptions) Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	merchantColl := db.Collection("merchants")
	docColl := db.Collection("merchant_documents")
	txColl := db.Collection("transactions")
	userAdapter := adapter.NewUserAdapter(userClient, g.User...)

	return &repositories{
		NewMerchantQueryRepository(merchantColl),
		NewMerchantCommandRepository(db, merchantColl),
		NewMerchantDocumentQueryRepository(docColl),
		NewMerchantDocumentCommandRepository(db, docColl),
		NewMerchantTransactionRepository(txColl),
		userAdapter,
		userAdapter,
	}
}

// UserAdapter returns the guarded user gRPC adapter built by NewRepositories
// so the service wiring and the repository share one breaker.
func (r *repositories) UserAdapter() adapter.UserAdapter {
	return r.userAdapter
}
