package idempotency

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"

	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
)

const MaxKeyLength = 255

// ValidateKey validates the client-provided idempotency key.
func ValidateKey(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return sharedErrors.NewBadRequestError("Idempotency-Key header is required")
	}
	if len(key) > MaxKeyLength {
		return sharedErrors.NewBadRequestError("Idempotency-Key must be at most 255 characters")
	}
	return nil
}

// HashRequest produces a stable SHA-256 hash of a request payload. Fields tagged
// json:"-" (the key and computed hash) are intentionally excluded.
func HashRequest(request any) string {
	payload, err := json.Marshal(request)
	if err != nil {
		// Requests used by the financial commands contain only JSON-marshalable
		// primitives. Keep a deterministic fallback if a future field violates that.
		payload = []byte("<unmarshalable-request>")
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}
