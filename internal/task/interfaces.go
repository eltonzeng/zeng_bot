package task

import "zeng_bot/internal/models"

// CheckoutClient defines the HTTP operations required for a checkout flow.
// Implementations must use a TLS-spoofing client (e.g. bogdanfinn/tls-client).
type CheckoutClient interface {
	// AddToCart sends an add-to-cart request for the given product.
	AddToCart(event models.StockEvent) (cartID string, err error)

	// SubmitPayment finalizes the order with the given cart and profile.
	SubmitPayment(cartID string, profile models.Profile) (orderID string, err error)
}
