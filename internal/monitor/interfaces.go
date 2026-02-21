// Package monitor is responsible for polling Target's inventory API.
package monitor

import "zeng_bot/internal/models"

// Monitor defines the behavior for a stock monitoring service.
// It polls the inventory API and broadcasts StockEvents to workers.
type Monitor interface {
	// Start begins polling for the given products. Stock events are sent
	// to the returned channel. The caller should cancel the context or
	// call Stop to shut down polling.
	Start(products []models.TargetProduct) (<-chan models.StockEvent, error)

	// Stop gracefully shuts down the monitor.
	Stop()
}
