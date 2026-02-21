package task

import (
	"fmt"
	"log"

	"zeng_bot/internal/models"
)

// Worker represents a single checkout worker that listens for stock events
// and executes the checkout state machine.
type Worker struct {
	ID      int
	Profile models.Profile
	Client  CheckoutClient
	State   State
}

// NewWorker creates a worker with the given ID, profile, and client.
func NewWorker(id int, profile models.Profile, client CheckoutClient) *Worker {
	return &Worker{
		ID:      id,
		Profile: profile,
		Client:  client,
		State:   StateIdle,
	}
}

// Run processes a single stock event through the checkout state machine.
// It transitions through states sequentially: ATC -> Payment -> Success/Failed.
func (w *Worker) Run(event models.StockEvent) error {
	log.Printf("[worker %d] received stock event for DPCI %s", w.ID, event.Product.DPCI)

	// ATC
	w.State = StateAddingToCart
	log.Printf("[worker %d] state -> %s", w.ID, w.State)

	cartID, err := w.Client.AddToCart(event)
	if err != nil {
		w.State = StateFailed
		return fmt.Errorf("worker %d: add to cart failed: %w", w.ID, err)
	}

	// Payment
	w.State = StateSubmittingPayment
	log.Printf("[worker %d] state -> %s", w.ID, w.State)

	orderID, err := w.Client.SubmitPayment(cartID, w.Profile)
	if err != nil {
		w.State = StateFailed
		return fmt.Errorf("worker %d: payment failed: %w", w.ID, err)
	}

	w.State = StateSuccess
	log.Printf("[worker %d] state -> %s | order: %s", w.ID, w.State, orderID)
	return nil
}
