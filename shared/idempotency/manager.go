package idempotency

import (
	"context"
	"errors"

	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
)

// ErrKeyInUse is returned by a Store implementation's Claim method when the key
// is already claimed by a non-expired record owned by another request.
var ErrKeyInUse = errors.New("idempotency: key already claimed by another request")

// Outcome captures what a successful command execution produced so it can be
// replayed to retrying clients without re-running the mutation.
type Outcome struct {
	Status       string
	ResponseJSON []byte
	ResourceID   *int32
}

// Store is the persistence contract each command service's repository layer
// implements on top of its own idempotency_records table.
type Store interface {
	// Claim atomically claims scope+key for requestHash. It returns ErrKeyInUse
	// when a non-expired record already owns the key.
	Claim(ctx context.Context, scope, key, requestHash string) (*Record, error)
	// Get loads the current record for scope+key, if it has not expired.
	Get(ctx context.Context, scope, key string) (*Record, error)
	// Complete finalizes a successfully executed request so retries can replay it.
	Complete(ctx context.Context, scope, key, requestHash string, outcome Outcome) error
	// Release deletes the claim for a request that failed, allowing a clean retry.
	Release(ctx context.Context, scope, key, requestHash string) error
}

// ConflictError is the error a command service surfaces when the same
// Idempotency-Key is retried with a different request payload.
func ConflictError() error {
	return sharedErrors.NewConflictError("Idempotency-Key was already used with a different request payload")
}

// ProcessingError is the error surfaced while the original request carrying the
// same Idempotency-Key is still being processed.
func ProcessingError() error {
	return sharedErrors.NewConflictError("a request with this Idempotency-Key is still being processed; retry shortly")
}
