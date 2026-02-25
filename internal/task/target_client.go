package task

import (
	"encoding/json"
	"fmt"
	"log"

	"zeng_bot/internal/models"
	"zeng_bot/internal/session"
)

const (
	targetAPIKey   = "9f36aeafbe60771e321a7cc95a78140772ab3e96"
	targetCartURL  = "https://carts.target.com/web_checkouts/v1/cart_items?field_groups=CART%2CCART_ITEMS%2CSUMMARY&key=" + targetAPIKey
	targetOrderURL = "https://carts.target.com/web_checkouts/v1/checkout?key=" + targetAPIKey
)

// TargetClient is the production CheckoutClient that makes real
// API calls to Target using a TLS-spoofed Session.
type TargetClient struct {
	session   session.Session
	visitorID string
}

// NewTargetClient creates a TargetClient backed by the given Session.
// It runs WarmUp to populate cookies and visitorId, then logs in with
// the account credentials from the provided Profile.
func NewTargetClient(sess *session.TargetSession, profile models.Profile) (*TargetClient, error) {
	if err := sess.WarmUp(); err != nil {
		return nil, fmt.Errorf("failed to warm up session: %w", err)
	}

	if err := sess.Login(profile.Email, profile.Password); err != nil {
		return nil, fmt.Errorf("failed to log in: %w", err)
	}

	return &TargetClient{
		session:   sess,
		visitorID: sess.VisitorID,
	}, nil
}

// AddToCart sends a POST to Target's cart API for the given stock event.
func (c *TargetClient) AddToCart(event models.StockEvent) (string, error) {
	payload := models.ATCRequest{
		CartItem: models.ATCCartItem{
			TCIN:          event.Product.TCIN,
			Quantity:      1,
			ItemChannelID: "10",
		},
		CartType:        "REGULAR",
		ChannelID:       "10",
		ShoppingContext: "DIGITAL",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ATC payload: %w", err)
	}

	headers := models.DefaultHeaders()
	headers["Content-Type"] = "application/json"
	headers["Accept"] = "application/json"
	headers["Origin"] = "https://www.target.com"
	headers["Referer"] = "https://www.target.com/"
	headers["x-application-name"] = "web"
	if c.visitorID != "" {
		headers["x-visitor-id"] = c.visitorID
	}

	log.Printf("[target-client] ATC request for TCIN %s", event.Product.TCIN)
	log.Printf("[target-client] ATC body: %s", string(body))

	status, respBody, err := c.session.Do("POST", targetCartURL, headers, body)
	if err != nil {
		return "", fmt.Errorf("ATC request failed: %w", err)
	}

	log.Printf("[target-client] ATC response status=%d body=%s", status, string(respBody))

	if status < 200 || status >= 300 {
		return "", fmt.Errorf("ATC returned status %d: %s", status, string(respBody))
	}

	var atcResp models.ATCResponse
	if err := json.Unmarshal(respBody, &atcResp); err != nil {
		return "", fmt.Errorf("failed to parse ATC response: %w", err)
	}

	if atcResp.CartID == "" {
		return "", fmt.Errorf("ATC response missing cart_id: %s", string(respBody))
	}

	log.Printf("[target-client] ATC success, cartID=%s", atcResp.CartID)
	return atcResp.CartID, nil
}

// SubmitPayment sends a POST to Target's checkout API to finalize the order.
func (c *TargetClient) SubmitPayment(cartID string, profile models.Profile) (string, error) {
	payload := models.OrderRequest{
		CartID:          cartID,
		ShippingAddress: profile.Shipping,
		BillingAddress:  profile.Billing,
		PaymentInfo: models.OrderPayment{
			CardNumber: profile.Payment.CardNumber,
			ExpMonth:   profile.Payment.ExpMonth,
			ExpYear:    profile.Payment.ExpYear,
			CVV:        profile.Payment.CVV,
			CardType:   "VISA",
		},
		ContactInfo: models.OrderContact{
			Name:  profile.Name,
			Email: profile.Email,
			Phone: profile.Phone,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal order payload: %w", err)
	}

	headers := models.DefaultHeaders()
	headers["Content-Type"] = "application/json"
	headers["Accept"] = "application/json"
	headers["Origin"] = "https://www.target.com"
	headers["Referer"] = "https://www.target.com/"
	headers["x-application-name"] = "web"
	headers["x-api-key"] = targetAPIKey
	if c.visitorID != "" {
		headers["x-visitor-id"] = c.visitorID
	}

	log.Printf("[target-client] submitting payment for cart %s", cartID)

	status, respBody, err := c.session.Do("POST", targetOrderURL, headers, body)
	if err != nil {
		return "", fmt.Errorf("order request failed: %w", err)
	}

	log.Printf("[target-client] order response status=%d body=%s", status, string(respBody))

	if status < 200 || status >= 300 {
		return "", fmt.Errorf("order returned status %d: %s", status, string(respBody))
	}

	var orderResp models.OrderResponse
	if err := json.Unmarshal(respBody, &orderResp); err != nil {
		return "", fmt.Errorf("failed to parse order response: %w", err)
	}

	if orderResp.OrderID == "" {
		return "", fmt.Errorf("order response missing order_id: %s", string(respBody))
	}

	log.Printf("[target-client] order success, orderID=%s", orderResp.OrderID)
	return orderResp.OrderID, nil
}

// compile-time check: TargetClient must satisfy CheckoutClient.
var _ CheckoutClient = (*TargetClient)(nil)
