// Package validation contains shared business validation for financial
// operations (topup, transaction, transfer, withdraw). Keeping these checks in
// one place makes the money-moving guards consistent across services.
package validation

import (
	"errors"

	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
)

// MaxAmount is the largest amount accepted by any financial command. It guards
// against overflow when converting between int32/int64 amounts and caps the
// blast radius of a single operation.
const MaxAmount = 1_000_000_000_000 // 1e12 in minor units

// ErrInvalidAmount is returned when an amount is not a positive value within
// the allowed business range.
var ErrInvalidAmount = sharedErrors.ErrBadRequest.WithMessage("amount must be greater than 0")

// ErrAmountTooLarge is returned when an amount exceeds MaxAmount.
var ErrAmountTooLarge = sharedErrors.ErrBadRequest.WithMessage("amount exceeds the maximum allowed value")

// ErrSameCardTransfer is returned when a transfer sender and receiver are the
// same card, which is a no-op that should never move money.
var ErrSameCardTransfer = sharedErrors.ErrBadRequest.WithMessage("sender and receiver card must be different")

// ValidateAmount rejects non-positive amounts and amounts above MaxAmount.
// Every financial command must call this before any balance mutation.
func ValidateAmount(amount int) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if amount > MaxAmount {
		return ErrAmountTooLarge
	}
	return nil
}

// ValidateAmountInt64 is the int64 variant used by request DTOs that carry
// int64 amounts.
func ValidateAmountInt64(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if amount > MaxAmount {
		return ErrAmountTooLarge
	}
	return nil
}

// ValidateTransferParties rejects a transfer where sender and receiver are the
// same card. Both sides are resolved independently, so the check runs on the
// resolved card numbers.
func ValidateTransferParties(senderCard, receiverCard string) error {
	if senderCard != "" && senderCard == receiverCard {
		return ErrSameCardTransfer
	}
	return nil
}

// ValidationError is a convenience alias so callers do not need to import both
// this package and shared/errors for the concrete type.
type ValidationError = sharedErrors.ValidationError

// IsValidationError reports whether err is one of the validation sentinels.
func IsValidationError(err error) bool {
	return errors.Is(err, ErrInvalidAmount) ||
		errors.Is(err, ErrAmountTooLarge) ||
		errors.Is(err, ErrSameCardTransfer)
}
