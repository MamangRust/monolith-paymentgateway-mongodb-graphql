package repository

import (
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
)

type TransferAtomicResult struct {
	Row      *salemongo.TransferAllFieldsRow
	Replayed bool
}
