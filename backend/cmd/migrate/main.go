package main

import (
	"flag"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/nxo/engine/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		log.Fatal("Usage: migrate <up|down|create> [name]")
	}

	m, err := migrate.New(
		"file://migrations",
		cfg.Database.DSN(),
	)
	if err != nil {
		log.Fatalf("Failed to create migrator: %v", err)
	}
	defer m.Close()

	switch args[0] {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration up failed: %v", err)
		}
		log.Println("✅ Migrations applied successfully")

	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration down failed: %v", err)
		}
		log.Println("✅ Migrations reverted successfully")

	case "create":
		if len(args) < 2 {
			log.Fatal("Usage: migrate create <name>")
		}
		createMigration(args[1])

	default:
		log.Fatalf("Unknown command: %s", args[0])
	}
}

func createMigration(name string) {
	// Create migration files
	upFile := "migrations/" + name + ".up.sql"
	downFile := "migrations/" + name + ".down.sql"

	if err := os.WriteFile(upFile, []byte("-- Migration UP\n"), 0644); err != nil {
		log.Fatalf("Failed to create up migration: %v", err)
	}
	if err := os.WriteFile(downFile, []byte("-- Migration DOWN\n"), 0644); err != nil {
		log.Fatalf("Failed to create down migration: %v", err)
	}

	log.Printf("✅ Created migrations: %s, %s", upFile, downFile)
}
