package service

import (
	"context"
	"database/sql"
	"testing"

	"home-inventory-system/internal/db"
	"home-inventory-system/internal/domain"
	"home-inventory-system/internal/repository"
	"home-inventory-system/scripts/seed"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestHandler(t *testing.T) (*Handler, func()) {
	t.Helper()

	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	// Enable foreign keys
	if _, err := sqlDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		sqlDB.Close()
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	// Run migrations
	if err := db.RunMigrations(sqlDB); err != nil {
		sqlDB.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	repo := repository.New(sqlDB)
	handler := NewHandler(repo)
	cleanup := func() {
		sqlDB.Close()
	}

	return handler, cleanup
}

func TestHandleListRootTypesQuery(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Seed minimal data
	if err := seed.SeedMinimal(ctx, handler.Repository()); err != nil {
		t.Fatalf("failed to seed: %v", err)
	}

	result := handler.HandleQuery(ctx, ListRootTypesQuery{})
	rootsResult, ok := result.(ListRootTypesResult)
	if !ok {
		t.Fatalf("expected ListRootTypesResult, got %T", result)
	}

	if rootsResult.Err != nil {
		t.Fatalf("unexpected error: %v", rootsResult.Err)
	}

	// SeedMinimal creates Kitchen and Garage roots
	if len(rootsResult.Types) != 2 {
		t.Errorf("expected 2 root types, got %d", len(rootsResult.Types))
	}
}

func TestHandleListChildTypesQuery(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create hierarchy
	createResult := handler.HandleCommand(ctx, CreateRootTypeCommand{Name: "Kitchen"})
	kitchenResult := createResult.(CreateRootTypeResult)
	if kitchenResult.Err != nil {
		t.Fatalf("failed to create Kitchen: %v", kitchenResult.Err)
	}

	handler.HandleCommand(ctx, CreateChildTypeCommand{ParentID: kitchenResult.Type.ID, Name: "Pantry"})
	handler.HandleCommand(ctx, CreateChildTypeCommand{ParentID: kitchenResult.Type.ID, Name: "Appliances"})

	result := handler.HandleQuery(ctx, ListChildTypesQuery{ParentID: kitchenResult.Type.ID})
	childResult, ok := result.(ListChildTypesResult)
	if !ok {
		t.Fatalf("expected ListChildTypesResult, got %T", result)
	}

	if childResult.Err != nil {
		t.Fatalf("unexpected error: %v", childResult.Err)
	}

	if len(childResult.Types) != 2 {
		t.Errorf("expected 2 children, got %d", len(childResult.Types))
	}
}

func TestHandleGetTypePathQuery(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create hierarchy: Kitchen > Pantry > Spices
	kitchenResult := handler.HandleCommand(ctx, CreateRootTypeCommand{Name: "Kitchen"}).(CreateRootTypeResult)
	pantryResult := handler.HandleCommand(ctx, CreateChildTypeCommand{ParentID: kitchenResult.Type.ID, Name: "Pantry"}).(CreateChildTypeResult)
	spicesResult := handler.HandleCommand(ctx, CreateChildTypeCommand{ParentID: pantryResult.Type.ID, Name: "Spices"}).(CreateChildTypeResult)

	result := handler.HandleQuery(ctx, GetTypePathQuery{TypeID: spicesResult.Type.ID})
	pathResult, ok := result.(GetTypePathResult)
	if !ok {
		t.Fatalf("expected GetTypePathResult, got %T", result)
	}

	if pathResult.Err != nil {
		t.Fatalf("unexpected error: %v", pathResult.Err)
	}

	if len(pathResult.Path) != 3 {
		t.Fatalf("expected path length 3, got %d", len(pathResult.Path))
	}

	// Verify order: root to leaf
	if pathResult.Path[0].Name != "Kitchen" {
		t.Errorf("expected first path element 'Kitchen', got '%s'", pathResult.Path[0].Name)
	}
	if pathResult.Path[2].Name != "Spices" {
		t.Errorf("expected last path element 'Spices', got '%s'", pathResult.Path[2].Name)
	}
}

func TestHandleIsLeafTypeQuery(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	kitchenResult := handler.HandleCommand(ctx, CreateRootTypeCommand{Name: "Kitchen"}).(CreateRootTypeResult)
	pantryResult := handler.HandleCommand(ctx, CreateChildTypeCommand{ParentID: kitchenResult.Type.ID, Name: "Pantry"}).(CreateChildTypeResult)

	// Kitchen should not be a leaf (has no items)
	result := handler.HandleQuery(ctx, IsLeafTypeQuery{TypeID: kitchenResult.Type.ID})
	leafResult := result.(IsLeafTypeResult)
	if leafResult.IsLeaf {
		t.Error("expected Kitchen to not be a leaf (no items)")
	}

	// Pantry should also not be a leaf (no items yet)
	result = handler.HandleQuery(ctx, IsLeafTypeQuery{TypeID: pantryResult.Type.ID})
	leafResult = result.(IsLeafTypeResult)
	if leafResult.IsLeaf {
		t.Error("expected Pantry to not be a leaf (no items)")
	}

	// Add an item to Pantry
	handler.HandleCommand(ctx, CreateItemCommand{
		Name:     "Rice",
		TypeID:   pantryResult.Type.ID,
		Quantity: 5.0,
		UnitType: "Grams",
	})

	// Now Pantry should be a leaf (has items)
	result = handler.HandleQuery(ctx, IsLeafTypeQuery{TypeID: pantryResult.Type.ID})
	leafResult = result.(IsLeafTypeResult)
	if !leafResult.IsLeaf {
		t.Error("expected Pantry to be a leaf (has items)")
	}
}

func TestHandleCreateItemCommand(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create a leaf type
	kitchenResult := handler.HandleCommand(ctx, CreateRootTypeCommand{Name: "Kitchen"}).(CreateRootTypeResult)

	// Create an item
	result := handler.HandleCommand(ctx, CreateItemCommand{Name: "Spoon", TypeID: kitchenResult.Type.ID, Quantity: 1.0, UnitType: domain.UnitTypeCount})
	itemResult := result.(CreateItemResult)
	if itemResult.Err != nil {
		t.Fatalf("unexpected error: %v", itemResult.Err)
	}

	if itemResult.Item.Name != "Spoon" {
		t.Errorf("expected item name 'Spoon', got '%s'", itemResult.Item.Name)
	}
}

func TestHandleListItemsByTypeQuery(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	kitchenResult := handler.HandleCommand(ctx, CreateRootTypeCommand{Name: "Kitchen"}).(CreateRootTypeResult)
	handler.HandleCommand(ctx, CreateItemCommand{Name: "Spoon", TypeID: kitchenResult.Type.ID, Quantity: 1.0, UnitType: domain.UnitTypeCount})
	handler.HandleCommand(ctx, CreateItemCommand{Name: "Fork", TypeID: kitchenResult.Type.ID, Quantity: 2.0, UnitType: domain.UnitTypeCount})

	result := handler.HandleQuery(ctx, ListItemsByTypeQuery{TypeID: kitchenResult.Type.ID})
	itemsResult := result.(ListItemsByTypeResult)
	if itemsResult.Err != nil {
		t.Fatalf("unexpected error: %v", itemsResult.Err)
	}

	if len(itemsResult.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(itemsResult.Items))
	}
}

func TestHandleCountItemsByTypeQuery(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	kitchenResult := handler.HandleCommand(ctx, CreateRootTypeCommand{Name: "Kitchen"}).(CreateRootTypeResult)
	handler.HandleCommand(ctx, CreateItemCommand{Name: "Spoon", TypeID: kitchenResult.Type.ID, Quantity: 1.0, UnitType: domain.UnitTypeCount})
	handler.HandleCommand(ctx, CreateItemCommand{Name: "Fork", TypeID: kitchenResult.Type.ID, Quantity: 2.0, UnitType: domain.UnitTypeCount})
	handler.HandleCommand(ctx, CreateItemCommand{Name: "Knife", TypeID: kitchenResult.Type.ID, Quantity: 3.0, UnitType: domain.UnitTypeCount})

	result := handler.HandleQuery(ctx, CountItemsByTypeQuery{TypeID: kitchenResult.Type.ID})
	countResult := result.(CountItemsByTypeResult)
	if countResult.Err != nil {
		t.Fatalf("unexpected error: %v", countResult.Err)
	}

	if countResult.Count != 3 {
		t.Errorf("expected count 3, got %d", countResult.Count)
	}
}

func TestHandleListLeafTypesQuery(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create hierarchy: Kitchen > Pantry > Spices
	//                         > Appliances
	kitchenResult := handler.HandleCommand(ctx, CreateRootTypeCommand{Name: "Kitchen"}).(CreateRootTypeResult)
	pantryResult := handler.HandleCommand(ctx, CreateChildTypeCommand{ParentID: kitchenResult.Type.ID, Name: "Pantry"}).(CreateChildTypeResult)
	handler.HandleCommand(ctx, CreateChildTypeCommand{ParentID: pantryResult.Type.ID, Name: "Spices"})
	handler.HandleCommand(ctx, CreateChildTypeCommand{ParentID: kitchenResult.Type.ID, Name: "Appliances"})

	result := handler.HandleQuery(ctx, ListLeafTypesQuery{})
	leafResult := result.(ListLeafTypesResult)
	if leafResult.Err != nil {
		t.Fatalf("unexpected error: %v", leafResult.Err)
	}

	// Spices and Appliances are leaves
	if len(leafResult.Types) != 2 {
		t.Errorf("expected 2 leaf types, got %d", len(leafResult.Types))
	}
}

func TestSeedIntegration(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Seed full data
	if err := seed.Seed(ctx, handler.Repository()); err != nil {
		t.Fatalf("failed to seed: %v", err)
	}

	// Verify roots were created
	result := handler.HandleQuery(ctx, ListRootTypesQuery{})
	rootsResult := result.(ListRootTypesResult)
	if rootsResult.Err != nil {
		t.Fatalf("unexpected error: %v", rootsResult.Err)
	}

	// Should have Kitchen, Electronics, Garage
	if len(rootsResult.Types) != 3 {
		t.Errorf("expected 3 root types, got %d", len(rootsResult.Types))
	}

	// Verify items were created by counting items at all leaf types
	leafResult := handler.HandleQuery(ctx, ListLeafTypesQuery{}).(ListLeafTypesResult)
	if leafResult.Err != nil {
		t.Fatalf("unexpected error: %v", leafResult.Err)
	}

	var totalItems int64
	for _, leafType := range leafResult.Types {
		countResult := handler.HandleQuery(ctx, CountItemsByTypeQuery{TypeID: leafType.ID}).(CountItemsByTypeResult)
		if countResult.Err != nil {
			t.Fatalf("unexpected error counting items for type %d: %v", leafType.ID, countResult.Err)
		}
		totalItems += countResult.Count
	}

	// Should have many items from the seed
	if totalItems < 30 {
		t.Errorf("expected at least 30 items, got %d", totalItems)
	}
}
