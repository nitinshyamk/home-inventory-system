package ui

import (
	"context"
	"testing"

	"home-inventory-system/internal/db"
	"home-inventory-system/internal/repository"
	"home-inventory-system/internal/service"
)

// setupTestHandler creates an in-memory SQLite DB with test data for UI tests.
// The database is seeded with a sample hierarchy via the service layer.
func setupTestHandler(t *testing.T) *service.Handler {
	t.Helper()

	// Create in-memory database
	cfg := db.Config{
		Path: ":memory:",
	}

	database, err := db.Open(cfg)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	// Run migrations
	if err := db.RunMigrations(database); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create handler
	repo := repository.New(database)
	handler := service.NewHandler(repo)

	ctx := context.Background()

	// Seed test data via service layer
	// Create root type: Electronics
	electronicsResult := handler.HandleCommand(ctx, service.CreateRootTypeCommand{
		Name:        "Electronics",
		Description: "Electronic devices",
	})
	if result, ok := electronicsResult.(service.CreateRootTypeResult); ok && result.Err != nil {
		t.Fatalf("Failed to create Electronics type: %v", result.Err)
	}

	// Create child type: Computers (child of Electronics)
	computersResult := handler.HandleCommand(ctx, service.CreateChildTypeCommand{
		ParentID:    1, // Electronics ID
		Name:        "Computers",
		Description: "Computing devices",
	})
	if result, ok := computersResult.(service.CreateChildTypeResult); ok && result.Err != nil {
		t.Fatalf("Failed to create Computers type: %v", result.Err)
	}

	// Create item under Computers (makes it a leaf)
	itemResult := handler.HandleCommand(ctx, service.CreateItemCommand{
		Name:   "Laptop",
		TypeID: 2, // Computers ID
	})
	if result, ok := itemResult.(service.CreateItemResult); ok && result.Err != nil {
		t.Fatalf("Failed to create Laptop item: %v", result.Err)
	}

	return handler
}
