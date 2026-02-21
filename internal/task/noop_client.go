package task

import (
	"log"

	"zeng_bot/internal/models"
)

// NoOpClient is a skeleton CheckoutClient that satisfies the interface
// but only logs operations. Use this for testing the worker pool and
// state machine without making real API calls.
type NoOpClient struct{}

// AddToCart logs the request and returns a fake cart ID.
func (c *NoOpClient) AddToCart(event models.StockEvent) (string, error) {
	log.Printf("[noop] AddToCart called for DPCI %s at store %s", event.Product.DPCI, event.Product.StoreID)
	return "fake-cart-id-001", nil
}

// SubmitPayment logs the request and returns a fake order ID.
func (c *NoOpClient) SubmitPayment(cartID string, profile models.Profile) (string, error) {
	log.Printf("[noop] SubmitPayment called for cart %s, profile %s", cartID, profile.Name)
	return "fake-order-id-001", nil
}

// compile-time check: NoOpClient must satisfy CheckoutClient.
var _ CheckoutClient = (*NoOpClient)(nil)
