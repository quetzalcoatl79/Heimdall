package main

import (
	"log"
	"os"

	"github.com/nxo/engine/internal/config"
	"github.com/nxo/engine/internal/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create and start server
	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	port := cfg.App.Port
	if port == "" {
		port = "8080"
	}
	_ = os.Getenv("APP_PORT")

	log.Printf("🚀 Engine API starting on port %s", port)
	if err := srv.Start(":" + port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
