package resilience

import (
	"context"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"go.uber.org/zap"
)

// DependencyGuard wraps outbound dependency calls (gRPC clients) with a
// per-call deadline, a bulkhead (max concurrent in-flight calls) and a circuit
// breaker. A nil *DependencyGuard is valid and behaves as a plain passthrough,
// which is what makes guards optional for adapters (see pkg/adapter/guard.go).
type DependencyGuard struct {
	name     string
	breaker  *CircuitBreaker
	limiter  *RequestLimiter
	timeout  time.Duration
	logger   logger.LoggerInterface
	inFlight int64
}

// NewDependencyGuard builds a guard for one remote dependency.
//
//	threshold     failures (while closed) before the breaker opens
//	timeoutSecs   seconds the breaker stays open before going half-open
//	maxConcurrent bulkhead: max in-flight calls; excess calls fail fast
//	callTimeout   deadline applied to each wrapped call
func NewDependencyGuard(name string, threshold uint64, timeoutSecs uint64,
	maxConcurrent int64, callTimeout time.Duration, log logger.LoggerInterface) *DependencyGuard {
	return &DependencyGuard{
		name:    name,
		breaker: NewCircuitBreaker(threshold, timeoutSecs, log),
		limiter: NewRequestLimiter(maxConcurrent, log),
		timeout: callTimeout,
		logger:  log,
	}
}

// Call runs fn under the guard. A nil receiver is a passthrough so adapters can
// be constructed without a guard.
func (g *DependencyGuard) Call(ctx context.Context, fn func(ctx context.Context) error) error {
	if g == nil {
		return fn(ctx)
	}

	if g.breaker.IsOpen() {
		err := ErrCircuitOpen(g.name)
		g.logger.Error("dependency guard rejected call", zap.String("dependency", g.name), zap.Error(err))
		return err
	}

	if !g.limiter.TryAcquire() {
		err := ErrBulkheadFull(g.name)
		g.logger.Error("dependency guard bulkhead full", zap.String("dependency", g.name), zap.Error(err))
		g.breaker.RecordFailure()
		return err
	}
	defer g.limiter.Release()

	callCtx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	callErr := fn(callCtx)
	if callErr != nil {
		g.breaker.RecordFailure()
		return callErr
	}

	g.breaker.RecordSuccess()
	return nil
}

// GuardError marks rejections made by the guard itself (open breaker, full
// bulkhead) as opposed to errors surfaced by the wrapped call.
type GuardError struct {
	Dependency string
	Reason     string
}

func (e *GuardError) Error() string {
	return "dependency " + e.Dependency + ": " + e.Reason
}

// ErrCircuitOpen reports that the guard's circuit breaker is open.
func ErrCircuitOpen(dependency string) error {
	return &GuardError{Dependency: dependency, Reason: "circuit breaker is open"}
}

// ErrBulkheadFull reports that the guard's bulkhead has no free permits.
func ErrBulkheadFull(dependency string) error {
	return &GuardError{Dependency: dependency, Reason: "bulkhead is full"}
}
