package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"home-inventory-system/internal/db"
	"home-inventory-system/internal/repository"
	"home-inventory-system/internal/service"
	"home-inventory-system/internal/ui"
)

func main() {
	// Parse command-line flags
	dbPath := flag.String("db-path", "", "Path to SQLite database file (default: ./inventory.db or INVENTORY_DB_PATH env var)")
	flag.Parse()

	// Create database configuration
	var cfg db.Config
	if *dbPath != "" {
		cfg = db.NewConfigWithPath(*dbPath)
	} else {
		cfg = db.DefaultConfig()
	}

	// Open database connection
	database, err := db.Open(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// Run migrations
	if err := db.RunMigrations(database); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	// Initialize layers
	repo := repository.New(database)
	handler := service.NewHandler(repo)

	// Run the application
	if err := ui.Run(handler); err != nil {
		log.Fatal(err)
	}
}
