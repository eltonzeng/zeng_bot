// Package task implements the state machine for a single checkout attempt.
package task

// State represents the current phase of a checkout task's lifecycle.
type State int

const (
	// StateIdle means the worker is waiting for a stock event.
	StateIdle State = iota
	// StateAddingToCart means the worker is sending the ATC request.
	StateAddingToCart
	// StateSubmittingPayment means the worker is submitting payment details.
	StateSubmittingPayment
	// StateSuccess means the order was placed successfully.
	StateSuccess
	// StateFailed means the task encountered a terminal error.
	StateFailed
)

// String returns a human-readable label for the state.
func (s State) String() string {
	switch s {
	case StateIdle:
		return "Idle"
	case StateAddingToCart:
		return "Adding to Cart"
	case StateSubmittingPayment:
		return "Submitting Payment"
	case StateSuccess:
		return "Success"
	case StateFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}
