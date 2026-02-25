// Package main is the entry point for the checkout bot CLI.
package main

import (
	"log"
	"os"

	"zeng_bot/internal/models"
	"zeng_bot/internal/orchestrator"
	"zeng_bot/internal/session"
	"zeng_bot/internal/task"
)

func main() {
	log.Println("zeng_bot starting...")

	// Placeholder config — will be replaced with CSV/JSON parsing.
	cfg := orchestrator.Config{
		WorkerCount: 3,
		Profile: models.Profile{
			Name:     "Test User",
			Email:    "test",
			Password: "test",
		},
		Products: []models.TargetProduct{
			{DPCI: "000-00-0000", TCIN: "89828965", StoreID: "1234"},
		},
	}

	// Set USE_REAL_CLIENT=1 to use the production TargetClient
	// backed by real HTTP sessions. Default is NoOp for safe testing.
	useReal := os.Getenv("USE_REAL_CLIENT") == "1"

	var clientFactory func() task.CheckoutClient
	if useReal {
		log.Println("[main] using real TargetClient")
		clientFactory = func() task.CheckoutClient {
			sess, err := session.NewTargetSession()
			if err != nil {
				log.Fatalf("[main] failed to create session: %v", err)
			}
			client, err := task.NewTargetClient(sess, cfg.Profile)
			if err != nil {
				log.Fatalf("[main] failed to create target client: %v", err)
			}
			return client
		}
	} else {
		log.Println("[main] using NoOpClient (set USE_REAL_CLIENT=1 for real requests)")
		clientFactory = func() task.CheckoutClient {
			return &task.NoOpClient{}
		}
	}

	orch := orchestrator.New(cfg, clientFactory)

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
