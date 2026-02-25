// Package models defines the core data structures used throughout the bot.
package models

// Profile represents a user's account credentials, billing and shipping
// information for checkout. Password is required — Target does not allow
// guest checkout.
type Profile struct {
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Phone    string  `json:"phone"`
	Billing  Address `json:"billing"`
	Shipping Address `json:"shipping"`
	Payment  Payment `json:"payment"`
}

// Address holds a street address used for billing or shipping.
type Address struct {
	Line1   string `json:"line1"`
	Line2   string `json:"line2,omitempty"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zip_code"`
	Country string `json:"country"`
}

// Payment holds credit card details for checkout submission.
type Payment struct {
	CardNumber string `json:"card_number"`
	ExpMonth   string `json:"exp_month"`
	ExpYear    string `json:"exp_year"`
	CVV        string `json:"cvv"`
}
