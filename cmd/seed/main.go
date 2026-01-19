package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"home-inventory-system/internal/db"
	"home-inventory-system/internal/repository"
	"home-inventory-system/scripts/seed"
)

func main() {
	dbPath := flag.String("db-path", "", "Path to SQLite database file (default: ./inventory.db or INVENTORY_DB_PATH env var)")
	minimal := flag.Bool("minimal", false, "Use minimal seed data for quick testing")
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

	// Initialize repository
	repo := repository.New(database)

	// Seed the database
	ctx := context.Background()
	if *minimal {
		fmt.Println("Seeding database with minimal data...")
		if err := seed.SeedMinimal(ctx, repo); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to seed database: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("Seeding database with full data...")
		if err := seed.Seed(ctx, repo); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to seed database: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("Database seeded successfully!")
}
