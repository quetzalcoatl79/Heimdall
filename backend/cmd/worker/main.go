package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nxo/engine/internal/config"
	"github.com/nxo/engine/internal/workers"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create worker manager
	manager, err := workers.NewManager(cfg)
	if err != nil {
		log.Fatalf("Failed to create worker manager: %v", err)
	}

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down workers...")
		manager.Shutdown()
	}()

	log.Printf("🔧 Worker started with concurrency: %d", cfg.Worker.Concurrency)
	manager.Start()
}
