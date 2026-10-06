package repository

import (
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
)

type WithdrawAtomicResult struct {
	Row      *salemongo.WithdrawAllFieldsRow
	Replayed bool
}
