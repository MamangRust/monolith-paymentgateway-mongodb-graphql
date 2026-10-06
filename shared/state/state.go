// Package state defines the financial command state machine shared by
// transaction and transfer services.
//
// The machine is intentionally small and backward compatible: every status is
// a plain string that fits the existing VARCHAR(20) status columns, and rows
// created before this package existed keep their legacy values (pending,
// success, failed).
//
// Transitions are guarded at the database layer (UPDATE ... WHERE status = <from>)
// so a transition is atomic and exactly-once: two concurrent actors cannot both
// move a row out of the same state.
package state

import "fmt"

// Command lifecycle statuses.
const (
	// Pending is the initial state after a command row is created.
	Pending = "pending"
	// Processing means settlement steps (debit/credit) are in flight.
	Processing = "processing"
	// Success means all settlement steps completed.
	Success = "success"
	// Failed means the command failed before any money moved.
	Failed = "failed"
	// Compensating means a partial settlement was detected and reversal is in flight.
	Compensating = "compensating"
	// Compensated means the reversal of a partial settlement completed.
	Compensated = "compensated"
	// Unknown means the outcome of a remote step cannot be determined and the
	// row needs reconciliation.
	Unknown = "unknown"
)

// allowedTransitions enumerates every legal (from -> to) pair.
var allowedTransitions = map[string]map[string]bool{
	Pending: {
		Processing: true,
		Failed:     true,
	},
	Processing: {
		Success:      true,
		Compensating: true,
		Failed:       true,
		Unknown:      true,
	},
	Compensating: {
		Compensated: true,
		Unknown:     true,
	},
	Unknown: {
		Compensating: true, // recovery worker may retry compensation
		Compensated:  true,
	},
}

// CanTransition reports whether moving from one status to another is legal.
func CanTransition(from, to string) bool {
	allowed, ok := allowedTransitions[from]
	if !ok {
		return false
	}
	return allowed[to]
}

// CheckTransition returns an error describing an illegal transition.
func CheckTransition(from, to string) error {
	if CanTransition(from, to) {
		return nil
	}
	return fmt.Errorf("illegal status transition: %q -> %q", from, to)
}

// IsTerminal reports whether a status is a final state that the command row
// should not leave (success, failed, compensated).
func IsTerminal(status string) bool {
	switch status {
	case Success, Failed, Compensated:
		return true
	}
	return false
}

// IsRecoverable reports whether a status is one the recovery worker should
// scan for (processing, compensating, unknown).
func IsRecoverable(status string) bool {
	switch status {
	case Processing, Compensating, Unknown:
		return true
	}
	return false
}

// IsValid reports whether a status string is a known state.
func IsValid(status string) bool {
	_, ok := allowedTransitions[status]
	return ok || status == Success || status == Failed || status == Compensated
}
