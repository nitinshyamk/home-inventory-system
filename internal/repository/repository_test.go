package repository

import (
	"context"
	"database/sql"
	"testing"

	"home-inventory-system/internal/db"
	"home-inventory-system/internal/domain"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) (*Repository, func()) {
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

	repo := New(sqlDB)
	cleanup := func() {
		sqlDB.Close()
	}

	return repo, cleanup
}

func TestCreateRootType(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	itemType, err := repo.CreateRootType(ctx, "Electronics", "Electronic devices")
	if err != nil {
		t.Fatalf("failed to create root type: %v", err)
	}

	if itemType.Name != "Electronics" {
		t.Errorf("expected name 'Electronics', got '%s'", itemType.Name)
	}
	if itemType.Depth != 0 {
		t.Errorf("expected depth 0, got %d", itemType.Depth)
	}
	if itemType.ParentID != nil {
		t.Error("expected parent_id to be nil for root type")
	}
}

func TestCreateChildType(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create root type
	root, err := repo.CreateRootType(ctx, "Kitchen", "")
	if err != nil {
		t.Fatalf("failed to create root type: %v", err)
	}

	// Create child type
	child, err := repo.CreateChildType(ctx, root.ID, "Pantry Items", "")
	if err != nil {
		t.Fatalf("failed to create child type: %v", err)
	}

	if child.Depth != 1 {
		t.Errorf("expected depth 1, got %d", child.Depth)
	}
	if child.ParentID == nil || *child.ParentID != root.ID {
		t.Errorf("expected parent_id %d, got %v", root.ID, child.ParentID)
	}

	// Create grandchild
	grandchild, err := repo.CreateChildType(ctx, child.ID, "Spices", "")
	if err != nil {
		t.Fatalf("failed to create grandchild type: %v", err)
	}

	if grandchild.Depth != 2 {
		t.Errorf("expected depth 2, got %d", grandchild.Depth)
	}
}

func TestGetRootTypes(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create multiple root types
	_, err := repo.CreateRootType(ctx, "Kitchen", "")
	if err != nil {
		t.Fatalf("failed to create root type: %v", err)
	}
	root2, err := repo.CreateRootType(ctx, "Garage", "")
	if err != nil {
		t.Fatalf("failed to create root type: %v", err)
	}

	// Create a child under root2
	_, err = repo.CreateChildType(ctx, root2.ID, "Tools", "")
	if err != nil {
		t.Fatalf("failed to create child type: %v", err)
	}

	// Get root types
	roots, err := repo.GetRootTypes(ctx)
	if err != nil {
		t.Fatalf("failed to get root types: %v", err)
	}

	if len(roots) != 2 {
		t.Errorf("expected 2 root types, got %d", len(roots))
	}
}

func TestGetChildTypes(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	root, err := repo.CreateRootType(ctx, "Kitchen", "")
	if err != nil {
		t.Fatalf("failed to create root type: %v", err)
	}

	_, err = repo.CreateChildType(ctx, root.ID, "Pantry Items", "")
	if err != nil {
		t.Fatalf("failed to create child type: %v", err)
	}
	_, err = repo.CreateChildType(ctx, root.ID, "Appliances", "")
	if err != nil {
		t.Fatalf("failed to create child type: %v", err)
	}

	children, err := repo.GetChildTypes(ctx, root.ID)
	if err != nil {
		t.Fatalf("failed to get child types: %v", err)
	}

	if len(children) != 2 {
		t.Errorf("expected 2 children, got %d", len(children))
	}
}

func TestGetTypePath(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create hierarchy: Kitchen > Pantry Items > Spices
	kitchen, err := repo.CreateRootType(ctx, "Kitchen", "")
	if err != nil {
		t.Fatalf("failed to create kitchen: %v", err)
	}

	pantry, err := repo.CreateChildType(ctx, kitchen.ID, "Pantry Items", "")
	if err != nil {
		t.Fatalf("failed to create pantry: %v", err)
	}

	spices, err := repo.CreateChildType(ctx, pantry.ID, "Spices", "")
	if err != nil {
		t.Fatalf("failed to create spices: %v", err)
	}

	// Get path from spices
	path, err := repo.GetTypePath(ctx, spices.ID)
	if err != nil {
		t.Fatalf("failed to get type path: %v", err)
	}

	if len(path) != 3 {
		t.Fatalf("expected path length 3, got %d", len(path))
	}

	// Path should be ordered by depth (root first)
	if path[0].Name != "Kitchen" {
		t.Errorf("expected first element 'Kitchen', got '%s'", path[0].Name)
	}
	if path[1].Name != "Pantry Items" {
		t.Errorf("expected second element 'Pantry Items', got '%s'", path[1].Name)
	}
	if path[2].Name != "Spices" {
		t.Errorf("expected third element 'Spices', got '%s'", path[2].Name)
	}
}

func TestIsLeafType(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	root, err := repo.CreateRootType(ctx, "Kitchen", "")
	if err != nil {
		t.Fatalf("failed to create root: %v", err)
	}

	// Root should NOT be a leaf initially (no items)
	isLeaf, err := repo.IsLeafType(ctx, root.ID)
	if err != nil {
		t.Fatalf("failed to check leaf status: %v", err)
	}
	if isLeaf {
		t.Error("expected root to not be a leaf when it has no items")
	}

	// Add a child category
	child, err := repo.CreateChildType(ctx, root.ID, "Pantry", "")
	if err != nil {
		t.Fatalf("failed to create child: %v", err)
	}

	// Root should still not be a leaf (still no items)
	isLeaf, err = repo.IsLeafType(ctx, root.ID)
	if err != nil {
		t.Fatalf("failed to check leaf status: %v", err)
	}
	if isLeaf {
		t.Error("expected root to not be a leaf when it has no items")
	}

	// Child should also not be a leaf (no items yet)
	isLeaf, err = repo.IsLeafType(ctx, child.ID)
	if err != nil {
		t.Fatalf("failed to check leaf status: %v", err)
	}
	if isLeaf {
		t.Error("expected child to not be a leaf when it has no items")
	}

	// Add an item to the child
	_, err = repo.CreateItem(ctx, "Rice", child.ID, nil, 5.0, "Grams")
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Now child SHOULD be a leaf (has items)
	isLeaf, err = repo.IsLeafType(ctx, child.ID)
	if err != nil {
		t.Fatalf("failed to check leaf status: %v", err)
	}
	if !isLeaf {
		t.Error("expected child to be a leaf when it has items")
	}
}

func TestItemOnlyOnLeafNodes(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create hierarchy: Kitchen > Pantry Items
	kitchen, err := repo.CreateRootType(ctx, "Kitchen", "")
	if err != nil {
		t.Fatalf("failed to create kitchen: %v", err)
	}

	pantry, err := repo.CreateChildType(ctx, kitchen.ID, "Pantry Items", "")
	if err != nil {
		t.Fatalf("failed to create pantry: %v", err)
	}

	// Should be able to create item on leaf node (pantry)
	item, err := repo.CreateItem(ctx, "Rice", pantry.ID, nil, 1.0, domain.UnitTypeCount)
	if err != nil {
		t.Fatalf("failed to create item on leaf node: %v", err)
	}
	if item.Name != "Rice" {
		t.Errorf("expected item name 'Rice', got '%s'", item.Name)
	}

	// Should NOT be able to create item on non-leaf node (kitchen)
	_, err = repo.CreateItem(ctx, "Invalid Item", kitchen.ID, nil, 1.0, domain.UnitTypeCount)
	if err == nil {
		t.Error("expected error when creating item on non-leaf node")
	}
}

func TestCannotAddChildToNodeWithItems(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a root type
	kitchen, err := repo.CreateRootType(ctx, "Kitchen", "")
	if err != nil {
		t.Fatalf("failed to create kitchen: %v", err)
	}

	// Add an item to it (it's a leaf node)
	_, err = repo.CreateItem(ctx, "Spoon", kitchen.ID, nil, 1.0, domain.UnitTypeCount)
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Now try to add a child - should fail because there are items
	_, err = repo.CreateChildType(ctx, kitchen.ID, "Pantry Items", "")
	if err == nil {
		t.Error("expected error when adding child to node with items")
	}
}

func TestListLeafTypes(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create hierarchy: Kitchen > Pantry Items > Spices
	//                         > Appliances
	kitchen, _ := repo.CreateRootType(ctx, "Kitchen", "")
	pantry, _ := repo.CreateChildType(ctx, kitchen.ID, "Pantry Items", "")
	_, _ = repo.CreateChildType(ctx, pantry.ID, "Spices", "")
	_, _ = repo.CreateChildType(ctx, kitchen.ID, "Appliances", "")

	leaves, err := repo.ListLeafTypes(ctx)
	if err != nil {
		t.Fatalf("failed to list leaf types: %v", err)
	}

	// Spices and Appliances should be leaves
	if len(leaves) != 2 {
		t.Errorf("expected 2 leaf types, got %d", len(leaves))
	}
}

func TestGetTypeDescendants(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create hierarchy
	kitchen, _ := repo.CreateRootType(ctx, "Kitchen", "")
	pantry, _ := repo.CreateChildType(ctx, kitchen.ID, "Pantry Items", "")
	_, _ = repo.CreateChildType(ctx, pantry.ID, "Spices", "")
	_, _ = repo.CreateChildType(ctx, pantry.ID, "Grains", "")

	// Get all descendants of kitchen
	descendants, err := repo.GetTypeDescendants(ctx, kitchen.ID)
	if err != nil {
		t.Fatalf("failed to get descendants: %v", err)
	}

	// Should include Kitchen, Pantry Items, Spices, Grains
	if len(descendants) != 4 {
		t.Errorf("expected 4 descendants, got %d", len(descendants))
	}

	// Get descendants of pantry
	descendants, err = repo.GetTypeDescendants(ctx, pantry.ID)
	if err != nil {
		t.Fatalf("failed to get descendants: %v", err)
	}

	// Should include Pantry Items, Spices, Grains
	if len(descendants) != 3 {
		t.Errorf("expected 3 descendants of pantry, got %d", len(descendants))
	}
}

func TestDeleteItemTypeWithChildren(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create hierarchy
	kitchen, _ := repo.CreateRootType(ctx, "Kitchen", "")
	_, _ = repo.CreateChildType(ctx, kitchen.ID, "Pantry Items", "")

	// Try to delete kitchen - should fail because it has children (RESTRICT)
	err := repo.DeleteItemType(ctx, kitchen.ID)
	if err == nil {
		t.Error("expected error when deleting type with children")
	}
}

func TestDeleteItemTypeWithItems(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a leaf type with an item
	kitchen, _ := repo.CreateRootType(ctx, "Kitchen", "")
	_, _ = repo.CreateItem(ctx, "Spoon", kitchen.ID, nil, 1.0, domain.UnitTypeCount)

	// Try to delete kitchen - should fail because it has items (RESTRICT)
	err := repo.DeleteItemType(ctx, kitchen.ID)
	if err == nil {
		t.Error("expected error when deleting type with items")
	}
}

func TestCountItemsByType(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	kitchen, _ := repo.CreateRootType(ctx, "Kitchen", "")
	_, _ = repo.CreateItem(ctx, "Spoon", kitchen.ID, nil, 1.0, domain.UnitTypeCount)
	_, _ = repo.CreateItem(ctx, "Fork", kitchen.ID, nil, 2.0, domain.UnitTypeCount)
	_, _ = repo.CreateItem(ctx, "Knife", kitchen.ID, nil, 3.0, domain.UnitTypeCount)

	count, err := repo.CountItemsByType(ctx, kitchen.ID)
	if err != nil {
		t.Fatalf("failed to count items: %v", err)
	}

	if count != 3 {
		t.Errorf("expected 3 items, got %d", count)
	}
}

func TestListItemsByType(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	kitchen, _ := repo.CreateRootType(ctx, "Kitchen", "")
	_, _ = repo.CreateItem(ctx, "Spoon", kitchen.ID, nil, 1.0, domain.UnitTypeCount)
	_, _ = repo.CreateItem(ctx, "Fork", kitchen.ID, nil, 1.0, domain.UnitTypeCount)

	items, err := repo.ListItemsByType(ctx, kitchen.ID)
	if err != nil {
		t.Fatalf("failed to list items: %v", err)
	}

	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestLiftAndDeleteCategory_SubcategoryLift(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Root > Kitchen > Pantry, Appliances
	root, _ := repo.CreateRootType(ctx, "Root", "")
	kitchen, _ := repo.CreateChildType(ctx, root.ID, "Kitchen", "")
	pantry, _ := repo.CreateChildType(ctx, kitchen.ID, "Pantry", "")
	_, _ = repo.CreateChildType(ctx, kitchen.ID, "Appliances", "")

	if err := repo.LiftAndDeleteItemType(ctx, kitchen.ID); err != nil {
		t.Fatalf("lift-and-delete failed: %v", err)
	}

	// Pantry and Appliances should now be direct children of Root
	children, err := repo.GetChildTypes(ctx, root.ID)
	if err != nil {
		t.Fatalf("failed to get children: %v", err)
	}
	if len(children) != 2 {
		t.Errorf("expected 2 children of Root, got %d", len(children))
	}

	// Pantry's depth should be 1 (was 2)
	pantryUpdated, _ := repo.GetItemType(ctx, pantry.ID)
	if pantryUpdated.Depth != 1 {
		t.Errorf("expected Pantry depth=1, got %d", pantryUpdated.Depth)
	}
}

func TestLiftAndDeleteCategory_ItemLift(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Root > Pantry (leaf with items)
	root, _ := repo.CreateRootType(ctx, "Root", "")
	pantry, _ := repo.CreateChildType(ctx, root.ID, "Pantry", "")
	_, _ = repo.CreateItem(ctx, "Rice", pantry.ID, nil, 1.0, domain.UnitTypeCount)

	if err := repo.LiftAndDeleteItemType(ctx, pantry.ID); err != nil {
		t.Fatalf("lift-and-delete failed: %v", err)
	}

	// Rice should now be in Root
	items, err := repo.ListItemsByType(ctx, root.ID)
	if err != nil {
		t.Fatalf("failed to list items: %v", err)
	}
	if len(items) != 1 || items[0].Name != "Rice" {
		t.Errorf("expected Rice in Root, got %+v", items)
	}
}

func TestLiftAndDeleteCategory_RootWithItemsRejected(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	kitchen, _ := repo.CreateRootType(ctx, "Kitchen", "")
	_, _ = repo.CreateItem(ctx, "Spoon", kitchen.ID, nil, 1.0, domain.UnitTypeCount)

	err := repo.LiftAndDeleteItemType(ctx, kitchen.ID)
	if err == nil {
		t.Error("expected error for root category with items")
	}
}

func TestLiftAndDeleteCategory_DepthsCorrectAfterLift(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// A > B > C > D (3 levels deep)
	a, _ := repo.CreateRootType(ctx, "A", "")
	b, _ := repo.CreateChildType(ctx, a.ID, "B", "")
	c, _ := repo.CreateChildType(ctx, b.ID, "C", "")
	d, _ := repo.CreateChildType(ctx, c.ID, "D", "")

	// Delete B — C and D should be lifted to A
	if err := repo.LiftAndDeleteItemType(ctx, b.ID); err != nil {
		t.Fatalf("lift-and-delete failed: %v", err)
	}

	cUpdated, _ := repo.GetItemType(ctx, c.ID)
	dUpdated, _ := repo.GetItemType(ctx, d.ID)

	if cUpdated.Depth != 1 {
		t.Errorf("expected C depth=1, got %d", cUpdated.Depth)
	}
	if dUpdated.Depth != 2 {
		t.Errorf("expected D depth=2, got %d", dUpdated.Depth)
	}
}

func TestUpdateItem(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	kitchen, _ := repo.CreateRootType(ctx, "Kitchen", "")
	item, err := repo.CreateItem(ctx, "Spoon", kitchen.ID, nil, 1.0, domain.UnitTypeCount)
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	if err := repo.UpdateItem(ctx, item.ID, "Large Spoon", kitchen.ID, nil, 3.0, domain.UnitTypeCount); err != nil {
		t.Fatalf("failed to update item: %v", err)
	}

	updated, err := repo.GetItem(ctx, item.ID)
	if err != nil {
		t.Fatalf("failed to get updated item: %v", err)
	}
	if updated.Name != "Large Spoon" {
		t.Errorf("expected name 'Large Spoon', got '%s'", updated.Name)
	}
	if updated.Quantity != 3.0 {
		t.Errorf("expected quantity 3.0, got %f", updated.Quantity)
	}
}

func TestUpdateItemType(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	root, err := repo.CreateRootType(ctx, "Kitchen", "Old description")
	if err != nil {
		t.Fatalf("failed to create root type: %v", err)
	}

	if err := repo.UpdateItemType(ctx, root.ID, "Kitchen Updated", "New description"); err != nil {
		t.Fatalf("failed to update item type: %v", err)
	}

	updated, err := repo.GetItemType(ctx, root.ID)
	if err != nil {
		t.Fatalf("failed to get updated type: %v", err)
	}
	if updated.Name != "Kitchen Updated" {
		t.Errorf("expected name 'Kitchen Updated', got '%s'", updated.Name)
	}
	if updated.Description != "New description" {
		t.Errorf("expected description 'New description', got '%s'", updated.Description)
	}
}

func TestUniqueNamePerParent(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create two root types with same name - should fail
	_, err := repo.CreateRootType(ctx, "Kitchen", "")
	if err != nil {
		t.Fatalf("failed to create first kitchen: %v", err)
	}

	_, err = repo.CreateRootType(ctx, "Kitchen", "")
	if err == nil {
		t.Error("expected error when creating duplicate root type name")
	}

	// Create parent and two children with same name - should fail
	parent, _ := repo.CreateRootType(ctx, "Home", "")
	_, err = repo.CreateChildType(ctx, parent.ID, "Room", "")
	if err != nil {
		t.Fatalf("failed to create first room: %v", err)
	}

	_, err = repo.CreateChildType(ctx, parent.ID, "Room", "")
	if err == nil {
		t.Error("expected error when creating duplicate child type name under same parent")
	}

	// Same name under different parents should be OK
	parent2, _ := repo.CreateRootType(ctx, "Office", "")
	_, err = repo.CreateChildType(ctx, parent2.ID, "Room", "")
	if err != nil {
		t.Errorf("should allow same name under different parents: %v", err)
	}
}

// --- Location tests ---

func TestCreateRootLocation(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	loc, err := repo.CreateRootLocation(ctx, "House", "My house")
	if err != nil {
		t.Fatalf("failed to create root location: %v", err)
	}

	if loc.Name != "House" {
		t.Errorf("expected name 'House', got '%s'", loc.Name)
	}
	if loc.Depth != 0 {
		t.Errorf("expected depth 0, got %d", loc.Depth)
	}
	if loc.ParentID != nil {
		t.Error("expected parent_id to be nil for root location")
	}
	if loc.Description != "My house" {
		t.Errorf("expected description 'My house', got '%s'", loc.Description)
	}
}

func TestCreateChildLocation(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	house, err := repo.CreateRootLocation(ctx, "House", "")
	if err != nil {
		t.Fatalf("failed to create root location: %v", err)
	}

	bedroom, err := repo.CreateChildLocation(ctx, house.ID, "Bedroom", "")
	if err != nil {
		t.Fatalf("failed to create child location: %v", err)
	}

	if bedroom.Depth != 1 {
		t.Errorf("expected depth 1, got %d", bedroom.Depth)
	}
	if bedroom.ParentID == nil || *bedroom.ParentID != house.ID {
		t.Errorf("expected parent_id %d, got %v", house.ID, bedroom.ParentID)
	}

	closet, err := repo.CreateChildLocation(ctx, bedroom.ID, "Closet", "")
	if err != nil {
		t.Fatalf("failed to create grandchild location: %v", err)
	}

	if closet.Depth != 2 {
		t.Errorf("expected depth 2, got %d", closet.Depth)
	}
}

func TestGetRootLocations(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, _ = repo.CreateRootLocation(ctx, "House", "")
	_, _ = repo.CreateRootLocation(ctx, "Storage Unit", "")

	roots, err := repo.GetRootLocations(ctx)
	if err != nil {
		t.Fatalf("failed to get root locations: %v", err)
	}

	if len(roots) != 2 {
		t.Errorf("expected 2 root locations, got %d", len(roots))
	}
}

func TestGetLocationPath(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	house, _ := repo.CreateRootLocation(ctx, "House", "")
	bedroom, _ := repo.CreateChildLocation(ctx, house.ID, "Bedroom", "")
	closet, _ := repo.CreateChildLocation(ctx, bedroom.ID, "Closet", "")

	path, err := repo.GetLocationPath(ctx, closet.ID)
	if err != nil {
		t.Fatalf("failed to get location path: %v", err)
	}

	if len(path) != 3 {
		t.Fatalf("expected path length 3, got %d", len(path))
	}

	if path[0].Name != "House" {
		t.Errorf("expected first 'House', got '%s'", path[0].Name)
	}
	if path[1].Name != "Bedroom" {
		t.Errorf("expected second 'Bedroom', got '%s'", path[1].Name)
	}
	if path[2].Name != "Closet" {
		t.Errorf("expected third 'Closet', got '%s'", path[2].Name)
	}
}

func TestLocationLeafNodeEnforcement(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	kitchen, _ := repo.CreateRootType(ctx, "Kitchen", "")
	house, _ := repo.CreateRootLocation(ctx, "House", "")
	bedroom, _ := repo.CreateChildLocation(ctx, house.ID, "Bedroom", "")

	// Can assign item to leaf location (bedroom has no children)
	item, err := repo.CreateItem(ctx, "Lamp", kitchen.ID, &bedroom.ID, 1.0, domain.UnitTypeCount)
	if err != nil {
		t.Fatalf("failed to create item at leaf location: %v", err)
	}
	if item.LocationID == nil || *item.LocationID != bedroom.ID {
		t.Errorf("expected location_id %d, got %v", bedroom.ID, item.LocationID)
	}

	// Cannot assign item to non-leaf location (house has children)
	_, err = repo.CreateItem(ctx, "Couch", kitchen.ID, &house.ID, 1.0, domain.UnitTypeCount)
	if err == nil {
		t.Error("expected error when assigning item to non-leaf location")
	}
}

func TestLocationCannotAddChildUnderLocationWithItems(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	kitchen, _ := repo.CreateRootType(ctx, "Kitchen", "")
	shelf, _ := repo.CreateRootLocation(ctx, "Shelf", "")

	// Assign item to shelf (it's a leaf)
	_, err := repo.CreateItem(ctx, "Book", kitchen.ID, &shelf.ID, 1.0, domain.UnitTypeCount)
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Now try to add a sublocation under shelf — should fail
	_, err = repo.CreateChildLocation(ctx, shelf.ID, "Top Shelf", "")
	if err == nil {
		t.Error("expected error when creating sublocation under location with items")
	}
}

func TestListItemsByLocation(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	kitchen, _ := repo.CreateRootType(ctx, "Kitchen", "")
	shelf, _ := repo.CreateRootLocation(ctx, "Shelf", "")
	other, _ := repo.CreateRootLocation(ctx, "Drawer", "")

	_, _ = repo.CreateItem(ctx, "Book", kitchen.ID, &shelf.ID, 1.0, domain.UnitTypeCount)
	_, _ = repo.CreateItem(ctx, "Magazine", kitchen.ID, &shelf.ID, 1.0, domain.UnitTypeCount)
	_, _ = repo.CreateItem(ctx, "Pen", kitchen.ID, &other.ID, 1.0, domain.UnitTypeCount)

	items, err := repo.ListItemsByLocation(ctx, shelf.ID)
	if err != nil {
		t.Fatalf("failed to list items by location: %v", err)
	}

	if len(items) != 2 {
		t.Errorf("expected 2 items at shelf, got %d", len(items))
	}
}

func TestItemWithoutLocation(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	kitchen, _ := repo.CreateRootType(ctx, "Kitchen", "")

	item, err := repo.CreateItem(ctx, "Spoon", kitchen.ID, nil, 1.0, domain.UnitTypeCount)
	if err != nil {
		t.Fatalf("failed to create item without location: %v", err)
	}

	if item.LocationID != nil {
		t.Errorf("expected location_id nil, got %v", item.LocationID)
	}

	fetched, err := repo.GetItem(ctx, item.ID)
	if err != nil {
		t.Fatalf("failed to get item: %v", err)
	}
	if fetched.LocationID != nil {
		t.Errorf("expected location_id nil on fetch, got %v", fetched.LocationID)
	}
}

func TestDeleteLocationWithItems(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	kitchen, _ := repo.CreateRootType(ctx, "Kitchen", "")
	shelf, _ := repo.CreateRootLocation(ctx, "Shelf", "")
	_, _ = repo.CreateItem(ctx, "Book", kitchen.ID, &shelf.ID, 1.0, domain.UnitTypeCount)

	err := repo.DeleteLocation(ctx, shelf.ID)
	if err == nil {
		t.Error("expected error when deleting location that has items")
	}
}

func TestLocationUniqueNamePerParent(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := repo.CreateRootLocation(ctx, "House", "")
	if err != nil {
		t.Fatalf("failed to create first root location: %v", err)
	}

	_, err = repo.CreateRootLocation(ctx, "House", "")
	if err == nil {
		t.Error("expected error when creating duplicate root location name")
	}

	parent, _ := repo.CreateRootLocation(ctx, "Storage", "")
	_, err = repo.CreateChildLocation(ctx, parent.ID, "Shelf", "")
	if err != nil {
		t.Fatalf("failed to create first shelf: %v", err)
	}
	_, err = repo.CreateChildLocation(ctx, parent.ID, "Shelf", "")
	if err == nil {
		t.Error("expected error when creating duplicate child location name")
	}
}
