// Package async provides bounded helpers for fire-and-forget side effects
// (e.g. marking an operation as failed) that must not depend on the request
// context, which may already be cancelled when the caller returns.
package async

import (
	"context"
	"time"
)

// RunWithTimeout executes fn in a goroutine with a detached background context
// carrying only a fixed timeout. The caller's request context is deliberately
// not propagated: a client timeout or cancelled request must not prevent the
// side effect from completing.
//
// Callers must treat fn as best-effort: it should not block on unbounded work
// and should tolerate a short timeout.
func RunWithTimeout(timeout time.Duration, fn func(ctx context.Context)) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		fn(ctx)
	}()
}
