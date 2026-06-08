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

func TestHandleDeleteCategoryWithLiftCommand_SubcategoryCase(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()
	rootResult := handler.HandleCommand(ctx, CreateRootTypeCommand{Name: "Root"}).(CreateRootTypeResult)
	kitchenResult := handler.HandleCommand(ctx, CreateChildTypeCommand{ParentID: rootResult.Type.ID, Name: "Kitchen"}).(CreateChildTypeResult)
	handler.HandleCommand(ctx, CreateChildTypeCommand{ParentID: kitchenResult.Type.ID, Name: "Pantry"})

	result := handler.HandleCommand(ctx, DeleteCategoryWithLiftCommand{ID: kitchenResult.Type.ID}).(DeleteCategoryWithLiftResult)

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}

	// Pantry should now be under Root
	children := handler.HandleQuery(ctx, ListChildTypesQuery{ParentID: rootResult.Type.ID}).(ListChildTypesResult)
	if len(children.Types) != 1 || children.Types[0].Name != "Pantry" {
		t.Errorf("expected Pantry under Root, got %+v", children.Types)
	}
}

func TestHandleDeleteCategoryWithLiftCommand_ItemCase(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()
	rootResult := handler.HandleCommand(ctx, CreateRootTypeCommand{Name: "Root"}).(CreateRootTypeResult)
	pantryResult := handler.HandleCommand(ctx, CreateChildTypeCommand{ParentID: rootResult.Type.ID, Name: "Pantry"}).(CreateChildTypeResult)
	handler.HandleCommand(ctx, CreateItemCommand{Name: "Rice", TypeID: pantryResult.Type.ID, Quantity: 1.0, UnitType: domain.UnitTypeCount})

	result := handler.HandleCommand(ctx, DeleteCategoryWithLiftCommand{ID: pantryResult.Type.ID}).(DeleteCategoryWithLiftResult)

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}

	// Rice should now be in Root
	items := handler.HandleQuery(ctx, ListItemsByTypeQuery{TypeID: rootResult.Type.ID}).(ListItemsByTypeResult)
	if len(items.Items) != 1 || items.Items[0].Name != "Rice" {
		t.Errorf("expected Rice in Root, got %+v", items.Items)
	}
}

func TestHandleDeleteCategoryWithLiftCommand_RootRejection(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()
	kitchenResult := handler.HandleCommand(ctx, CreateRootTypeCommand{Name: "Kitchen"}).(CreateRootTypeResult)
	handler.HandleCommand(ctx, CreateItemCommand{Name: "Spoon", TypeID: kitchenResult.Type.ID, Quantity: 1.0, UnitType: domain.UnitTypeCount})

	result := handler.HandleCommand(ctx, DeleteCategoryWithLiftCommand{ID: kitchenResult.Type.ID}).(DeleteCategoryWithLiftResult)

	if result.Err == nil {
		t.Error("expected error for root category with items")
	}
}

func TestHandleGetCategoryChildSummaryQuery(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()
	kitchen := handler.HandleCommand(ctx, CreateRootTypeCommand{Name: "Kitchen"}).(CreateRootTypeResult)
	handler.HandleCommand(ctx, CreateChildTypeCommand{ParentID: kitchen.Type.ID, Name: "Pantry"})
	handler.HandleCommand(ctx, CreateChildTypeCommand{ParentID: kitchen.Type.ID, Name: "Appliances"})

	result := handler.HandleQuery(ctx, GetCategoryChildSummaryQuery{ID: kitchen.Type.ID}).(GetCategoryChildSummaryResult)

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.ChildTypeCount != 2 {
		t.Errorf("expected 2 children, got %d", result.ChildTypeCount)
	}
	if result.ItemCount != 0 {
		t.Errorf("expected 0 items, got %d", result.ItemCount)
	}
	if result.ParentID != nil {
		t.Errorf("expected nil parent for root type, got %v", result.ParentID)
	}
}

func TestHandleUpdateItemCommand_Success(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()
	catResult := handler.HandleCommand(ctx, CreateRootTypeCommand{Name: "Kitchen"}).(CreateRootTypeResult)
	itemResult := handler.HandleCommand(ctx, CreateItemCommand{
		Name: "Spoon", TypeID: catResult.Type.ID, Quantity: 1.0, UnitType: domain.UnitTypeCount,
	}).(CreateItemResult)

	result := handler.HandleCommand(ctx, UpdateItemCommand{
		ID: itemResult.Item.ID, Name: "Big Spoon", TypeID: catResult.Type.ID,
		Quantity: 5.0, UnitType: domain.UnitTypeGrams,
	}).(UpdateItemResult)

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.Item.Name != "Big Spoon" {
		t.Errorf("expected name 'Big Spoon', got '%s'", result.Item.Name)
	}
	if result.Item.Quantity != 5.0 {
		t.Errorf("expected quantity 5.0, got %f", result.Item.Quantity)
	}
}

func TestHandleUpdateItemCommand_EmptyName(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()
	catResult := handler.HandleCommand(ctx, CreateRootTypeCommand{Name: "Kitchen"}).(CreateRootTypeResult)
	itemResult := handler.HandleCommand(ctx, CreateItemCommand{
		Name: "Spoon", TypeID: catResult.Type.ID, Quantity: 1.0, UnitType: domain.UnitTypeCount,
	}).(CreateItemResult)

	result := handler.HandleCommand(ctx, UpdateItemCommand{
		ID: itemResult.Item.ID, Name: "", TypeID: catResult.Type.ID, Quantity: 1.0, UnitType: domain.UnitTypeCount,
	}).(UpdateItemResult)

	if result.ValidationError == nil {
		t.Error("expected validation error for empty name")
	}
}

func TestHandleUpdateItemCommand_InvalidQuantity(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()
	catResult := handler.HandleCommand(ctx, CreateRootTypeCommand{Name: "Kitchen"}).(CreateRootTypeResult)
	itemResult := handler.HandleCommand(ctx, CreateItemCommand{
		Name: "Spoon", TypeID: catResult.Type.ID, Quantity: 1.0, UnitType: domain.UnitTypeCount,
	}).(CreateItemResult)

	result := handler.HandleCommand(ctx, UpdateItemCommand{
		ID: itemResult.Item.ID, Name: "Spoon", TypeID: catResult.Type.ID, Quantity: -1.0, UnitType: domain.UnitTypeCount,
	}).(UpdateItemResult)

	if result.ValidationError == nil {
		t.Error("expected validation error for negative quantity")
	}
}

func TestHandleUpdateCategoryCommand_Success(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()
	createResult := handler.HandleCommand(ctx, CreateRootTypeCommand{Name: "Kitchen"}).(CreateRootTypeResult)
	if createResult.Err != nil {
		t.Fatalf("failed to create category: %v", createResult.Err)
	}

	result := handler.HandleCommand(ctx, UpdateCategoryCommand{
		ID:          createResult.Type.ID,
		Name:        "Kitchen Updated",
		Description: "New desc",
	}).(UpdateCategoryResult)

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.Type == nil {
		t.Fatal("expected updated type to be returned")
	}
	if result.Type.Name != "Kitchen Updated" {
		t.Errorf("expected name 'Kitchen Updated', got '%s'", result.Type.Name)
	}
}

func TestHandleUpdateCategoryCommand_EmptyName(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()
	createResult := handler.HandleCommand(ctx, CreateRootTypeCommand{Name: "Kitchen"}).(CreateRootTypeResult)

	result := handler.HandleCommand(ctx, UpdateCategoryCommand{
		ID:   createResult.Type.ID,
		Name: "",
	}).(UpdateCategoryResult)

	if result.ValidationError == nil {
		t.Error("expected validation error for empty name")
	}
}

func TestHandleUpdateCategoryCommand_DuplicateName(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()
	handler.HandleCommand(ctx, CreateRootTypeCommand{Name: "Garage"})
	kitchenResult := handler.HandleCommand(ctx, CreateRootTypeCommand{Name: "Kitchen"}).(CreateRootTypeResult)

	// Try to rename Kitchen to Garage (duplicate root name)
	result := handler.HandleCommand(ctx, UpdateCategoryCommand{
		ID:   kitchenResult.Type.ID,
		Name: "Garage",
	}).(UpdateCategoryResult)

	if result.Err == nil {
		t.Error("expected error for duplicate name")
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

// --- Location tests ---

func TestHandleCreateRootLocationCommand(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	result := handler.HandleCommand(ctx, CreateRootLocationCommand{
		Name:        "House",
		Description: "My home",
	}).(CreateRootLocationResult)

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.Location == nil {
		t.Fatal("expected non-nil location")
	}
	if result.Location.Name != "House" {
		t.Errorf("expected name 'House', got '%s'", result.Location.Name)
	}
	if result.Location.Depth != 0 {
		t.Errorf("expected depth 0, got %d", result.Location.Depth)
	}
	if result.Location.ParentID != nil {
		t.Error("expected nil parent_id for root location")
	}
}

func TestHandleCreateRootLocationCommand_EmptyName(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	result := handler.HandleCommand(ctx, CreateRootLocationCommand{Name: ""}).(CreateRootLocationResult)

	if result.ValidationError == nil {
		t.Error("expected validation error for empty name")
	}
	if result.Location != nil {
		t.Error("expected nil location on validation error")
	}
}

func TestHandleCreateChildLocationCommand(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	rootResult := handler.HandleCommand(ctx, CreateRootLocationCommand{Name: "House"}).(CreateRootLocationResult)
	if rootResult.Err != nil {
		t.Fatalf("failed to create root location: %v", rootResult.Err)
	}

	childResult := handler.HandleCommand(ctx, CreateChildLocationCommand{
		ParentID: rootResult.Location.ID,
		Name:     "Bedroom",
	}).(CreateChildLocationResult)

	if childResult.Err != nil {
		t.Fatalf("unexpected error: %v", childResult.Err)
	}
	if childResult.Location.Depth != 1 {
		t.Errorf("expected depth 1, got %d", childResult.Location.Depth)
	}
	if childResult.Location.ParentID == nil || *childResult.Location.ParentID != rootResult.Location.ID {
		t.Errorf("expected parent_id %d", rootResult.Location.ID)
	}
}

func TestHandleListChildLocationsQuery(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	rootResult := handler.HandleCommand(ctx, CreateRootLocationCommand{Name: "House"}).(CreateRootLocationResult)
	handler.HandleCommand(ctx, CreateChildLocationCommand{ParentID: rootResult.Location.ID, Name: "Bedroom"})
	handler.HandleCommand(ctx, CreateChildLocationCommand{ParentID: rootResult.Location.ID, Name: "Kitchen"})

	result := handler.HandleQuery(ctx, ListChildLocationsQuery{ParentID: rootResult.Location.ID}).(ListChildLocationsResult)

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if len(result.Locations) != 2 {
		t.Errorf("expected 2 child locations, got %d", len(result.Locations))
	}
}

func TestHandleGetLocationPathQuery(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	house := handler.HandleCommand(ctx, CreateRootLocationCommand{Name: "House"}).(CreateRootLocationResult).Location
	bedroom := handler.HandleCommand(ctx, CreateChildLocationCommand{ParentID: house.ID, Name: "Bedroom"}).(CreateChildLocationResult).Location
	closet := handler.HandleCommand(ctx, CreateChildLocationCommand{ParentID: bedroom.ID, Name: "Closet"}).(CreateChildLocationResult).Location

	result := handler.HandleQuery(ctx, GetLocationPathQuery{LocationID: closet.ID}).(GetLocationPathResult)

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if len(result.Path) != 3 {
		t.Fatalf("expected path length 3, got %d", len(result.Path))
	}
	if result.Path[0].Name != "House" {
		t.Errorf("expected 'House' at path[0], got '%s'", result.Path[0].Name)
	}
}

func TestCreateItemWithLocation(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	catResult := handler.HandleCommand(ctx, CreateRootTypeCommand{Name: "Kitchen"}).(CreateRootTypeResult)
	locResult := handler.HandleCommand(ctx, CreateRootLocationCommand{Name: "Shelf"}).(CreateRootLocationResult)

	result := handler.HandleCommand(ctx, CreateItemCommand{
		Name:       "Jar",
		TypeID:     catResult.Type.ID,
		LocationID: &locResult.Location.ID,
		Quantity:   1.0,
		UnitType:   domain.UnitTypeCount,
	}).(CreateItemResult)

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.Item.LocationID == nil {
		t.Fatal("expected non-nil location_id on created item")
	}
	if *result.Item.LocationID != locResult.Location.ID {
		t.Errorf("expected location_id %d, got %d", locResult.Location.ID, *result.Item.LocationID)
	}
}

func TestCreateItemWithoutLocation(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	catResult := handler.HandleCommand(ctx, CreateRootTypeCommand{Name: "Kitchen"}).(CreateRootTypeResult)

	result := handler.HandleCommand(ctx, CreateItemCommand{
		Name:     "Spoon",
		TypeID:   catResult.Type.ID,
		Quantity: 1.0,
		UnitType: domain.UnitTypeCount,
	}).(CreateItemResult)

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.Item.LocationID != nil {
		t.Errorf("expected nil location_id, got %v", result.Item.LocationID)
	}
}

func TestHandleListItemsByLocationQuery(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	catResult := handler.HandleCommand(ctx, CreateRootTypeCommand{Name: "Kitchen"}).(CreateRootTypeResult)
	shelfResult := handler.HandleCommand(ctx, CreateRootLocationCommand{Name: "Shelf"}).(CreateRootLocationResult)
	drawerResult := handler.HandleCommand(ctx, CreateRootLocationCommand{Name: "Drawer"}).(CreateRootLocationResult)

	handler.HandleCommand(ctx, CreateItemCommand{Name: "Book", TypeID: catResult.Type.ID, LocationID: &shelfResult.Location.ID, Quantity: 1.0, UnitType: domain.UnitTypeCount})
	handler.HandleCommand(ctx, CreateItemCommand{Name: "Magazine", TypeID: catResult.Type.ID, LocationID: &shelfResult.Location.ID, Quantity: 1.0, UnitType: domain.UnitTypeCount})
	handler.HandleCommand(ctx, CreateItemCommand{Name: "Pen", TypeID: catResult.Type.ID, LocationID: &drawerResult.Location.ID, Quantity: 1.0, UnitType: domain.UnitTypeCount})

	result := handler.HandleQuery(ctx, ListItemsByLocationQuery{LocationID: shelfResult.Location.ID}).(ListItemsByLocationResult)

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if len(result.Items) != 2 {
		t.Errorf("expected 2 items at shelf, got %d", len(result.Items))
	}
}
