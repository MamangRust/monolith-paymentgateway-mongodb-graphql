package idempotency

import "time"

const (
	StatusProcessing = "processing"
	StatusSuccess    = "success"
	StatusFailed     = "failed"
	StatusUnknown    = "unknown"
)

// Record is the service-neutral representation of a persisted command key.
// ResourceID is a pointer so callers can distinguish "not set" (NULL) from 0.
type Record struct {
	ID           int64
	Key          string
	RequestHash  string
	OperationID  string
	Status       string
	ResponseJSON []byte
	ResourceID   *int32
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ExpiresAt    time.Time
}
