package repository

import (
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
)

type TopupAtomicResult struct {
	Row      *salemongo.TopupAllFieldsRow
	Replayed bool
}
