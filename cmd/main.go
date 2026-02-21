// Package main is the entry point for the checkout bot CLI.
package main

import (
	"log"

	"zeng_bot/internal/models"
	"zeng_bot/internal/orchestrator"
	"zeng_bot/internal/task"
)

func main() {
	log.Println("zeng_bot starting...")

	// Placeholder config — will be replaced with CSV/JSON parsing.
	cfg := orchestrator.Config{
		WorkerCount: 3,
		Profile: models.Profile{
			Name:  "Test User",
			Email: "test@example.com",
		},
		Products: []models.TargetProduct{
			{DPCI: "000-00-0000", TCIN: "00000000", StoreID: "1234"},
		},
	}

	// Use the NoOpClient factory for skeleton testing.
	orch := orchestrator.New(cfg, func() task.CheckoutClient {
		return &task.NoOpClient{}
	})

	// Simulate a stock event to exercise the pipeline.
	events := make(chan models.StockEvent, 1)
	events <- models.StockEvent{
		Product:    cfg.Products[0],
		OfferID:    "test-offer",
		LocationID: "test-location",
	}
	close(events)

	orch.Run(events)
	log.Println("zeng_bot finished.")
}
