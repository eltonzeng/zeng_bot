package models

// TargetProduct represents a product to monitor and purchase on Target.
type TargetProduct struct {
	DPCI    string `json:"dpci"`
	TCIN    string `json:"tcin"`
	Name    string `json:"name,omitempty"`
	StoreID string `json:"store_id"`
}

// StockEvent is broadcast by the Monitor when a product becomes available.
// Workers receive this to begin the checkout flow.
type StockEvent struct {
	Product    TargetProduct `json:"product"`
	OfferID    string        `json:"offer_id"`
	LocationID string        `json:"location_id"`
}
