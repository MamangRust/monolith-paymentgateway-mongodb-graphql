package repository

import (
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
)

type TransactionAtomicResult struct {
	Row      *salemongo.TransactionAllFieldsRow
	Replayed bool
}

type AuthorizeTransactionResult struct {
	Row      *salemongo.TransactionAllFieldsRow
	Replayed bool
}
