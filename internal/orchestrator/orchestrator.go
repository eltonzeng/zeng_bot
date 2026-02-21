// Package orchestrator manages the application lifecycle, initializes the
// monitor, and runs the worker pool for concurrent checkout tasks.
package orchestrator

import (
	"log"
	"sync"

	"zeng_bot/internal/models"
	"zeng_bot/internal/task"
)

// Config holds the settings for the orchestrator.
type Config struct {
	WorkerCount int              `json:"worker_count"`
	Profile     models.Profile   `json:"profile"`
	Products    []models.TargetProduct `json:"products"`
	Proxies     []models.Proxy   `json:"proxies"`
}

// Orchestrator coordinates the monitor and worker pool.
type Orchestrator struct {
	cfg     Config
	workers []*task.Worker
}

// New creates an Orchestrator with the given config and a client factory.
// clientFactory is called once per worker to create its CheckoutClient.
func New(cfg Config, clientFactory func() task.CheckoutClient) *Orchestrator {
	workers := make([]*task.Worker, cfg.WorkerCount)
	for i := range workers {
		workers[i] = task.NewWorker(i, cfg.Profile, clientFactory())
	}
	return &Orchestrator{cfg: cfg, workers: workers}
}

// Run starts workers listening on the stock event channel. It blocks until
// the channel is closed (monitor stopped) and all workers finish.
func (o *Orchestrator) Run(events <-chan models.StockEvent) {
	var wg sync.WaitGroup

	for _, w := range o.workers {
		wg.Add(1)
		go func(w *task.Worker) {
			defer wg.Done()
			for event := range events {
				if err := w.Run(event); err != nil {
					log.Printf("[orchestrator] worker %d error: %v", w.ID, err)
				}
			}
		}(w)
	}

	log.Printf("[orchestrator] %d workers started, waiting for events...", len(o.workers))
	wg.Wait()
	log.Println("[orchestrator] all workers finished")
}
