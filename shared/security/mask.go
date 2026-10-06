// Package security provides helpers to keep sensitive identifiers (card
// numbers, API keys) out of logs, traces, and error responses.
package security

import "strings"

// MaskCardNumber masks all but the first 6 and last 4 digits of a card number.
// Short or malformed values are masked entirely so a raw PAN never leaks.
//
// Examples:
//
//	4532123456789012 -> 453212******9012
//	1234            -> ****
//	""              -> ""
func MaskCardNumber(cardNumber string) string {
	if cardNumber == "" {
		return ""
	}

	const (
		visiblePrefix = 6
		visibleSuffix = 4
	)

	trimmed := strings.TrimSpace(cardNumber)
	if len(trimmed) <= visiblePrefix+visibleSuffix {
		return strings.Repeat("*", len(trimmed))
	}

	return trimmed[:visiblePrefix] + strings.Repeat("*", len(trimmed)-visiblePrefix-visibleSuffix) + trimmed[len(trimmed)-visibleSuffix:]
}

// MaskAPIKey masks all but the last 4 characters of an API key. Short keys are
// masked entirely.
//
// Examples:
//
//	sk_live_9f2c1d8e -> ************8e
//	abc             -> ***
//	""              -> ""
func MaskAPIKey(apiKey string) string {
	if apiKey == "" {
		return ""
	}

	const visibleSuffix = 4

	trimmed := strings.TrimSpace(apiKey)
	if len(trimmed) <= visibleSuffix {
		return strings.Repeat("*", len(trimmed))
	}

	return strings.Repeat("*", len(trimmed)-visibleSuffix) + trimmed[len(trimmed)-visibleSuffix:]
}

// MaskEmail obfuscates the local part of an email address while keeping the
// domain readable for support/debugging.
//
//	john.doe@example.com -> j***e@example.com
func MaskEmail(email string) string {
	at := strings.LastIndex(email, "@")
	if at <= 1 {
		return strings.Repeat("*", len(email))
	}

	local := email[:at]
	domain := email[at:]

	if len(local) == 1 {
		return "*" + domain
	}

	return local[:1] + strings.Repeat("*", len(local)-2) + local[len(local)-1:] + domain
}
