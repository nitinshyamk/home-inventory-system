package repository

import (
	"context"
	"database/sql"
	"testing"

	"home-inventory-system/internal/db"

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

	itemType, err := repo.CreateRootType(ctx, "Electronics", sql.NullString{String: "Electronic devices", Valid: true})
	if err != nil {
		t.Fatalf("failed to create root type: %v", err)
	}

	if itemType.Name != "Electronics" {
		t.Errorf("expected name 'Electronics', got '%s'", itemType.Name)
	}
	if itemType.Depth != 0 {
		t.Errorf("expected depth 0, got %d", itemType.Depth)
	}
	if itemType.ParentID.Valid {
		t.Error("expected parent_id to be NULL for root type")
	}
}

func TestCreateChildType(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create root type
	root, err := repo.CreateRootType(ctx, "Kitchen", sql.NullString{})
	if err != nil {
		t.Fatalf("failed to create root type: %v", err)
	}

	// Create child type
	child, err := repo.CreateChildType(ctx, root.ID, "Pantry Items", sql.NullString{})
	if err != nil {
		t.Fatalf("failed to create child type: %v", err)
	}

	if child.Depth != 1 {
		t.Errorf("expected depth 1, got %d", child.Depth)
	}
	if !child.ParentID.Valid || child.ParentID.Int64 != root.ID {
		t.Errorf("expected parent_id %d, got %v", root.ID, child.ParentID)
	}

	// Create grandchild
	grandchild, err := repo.CreateChildType(ctx, child.ID, "Spices", sql.NullString{})
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
	_, err := repo.CreateRootType(ctx, "Kitchen", sql.NullString{})
	if err != nil {
		t.Fatalf("failed to create root type: %v", err)
	}
	root2, err := repo.CreateRootType(ctx, "Garage", sql.NullString{})
	if err != nil {
		t.Fatalf("failed to create root type: %v", err)
	}

	// Create a child under root2
	_, err = repo.CreateChildType(ctx, root2.ID, "Tools", sql.NullString{})
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

	root, err := repo.CreateRootType(ctx, "Kitchen", sql.NullString{})
	if err != nil {
		t.Fatalf("failed to create root type: %v", err)
	}

	_, err = repo.CreateChildType(ctx, root.ID, "Pantry Items", sql.NullString{})
	if err != nil {
		t.Fatalf("failed to create child type: %v", err)
	}
	_, err = repo.CreateChildType(ctx, root.ID, "Appliances", sql.NullString{})
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
	kitchen, err := repo.CreateRootType(ctx, "Kitchen", sql.NullString{})
	if err != nil {
		t.Fatalf("failed to create kitchen: %v", err)
	}

	pantry, err := repo.CreateChildType(ctx, kitchen.ID, "Pantry Items", sql.NullString{})
	if err != nil {
		t.Fatalf("failed to create pantry: %v", err)
	}

	spices, err := repo.CreateChildType(ctx, pantry.ID, "Spices", sql.NullString{})
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

	root, err := repo.CreateRootType(ctx, "Kitchen", sql.NullString{})
	if err != nil {
		t.Fatalf("failed to create root: %v", err)
	}

	// Root should be a leaf initially
	isLeaf, err := repo.IsLeafType(ctx, root.ID)
	if err != nil {
		t.Fatalf("failed to check leaf status: %v", err)
	}
	if !isLeaf {
		t.Error("expected root to be a leaf when it has no children")
	}

	// Add a child
	child, err := repo.CreateChildType(ctx, root.ID, "Pantry", sql.NullString{})
	if err != nil {
		t.Fatalf("failed to create child: %v", err)
	}

	// Root should no longer be a leaf
	isLeaf, err = repo.IsLeafType(ctx, root.ID)
	if err != nil {
		t.Fatalf("failed to check leaf status: %v", err)
	}
	if isLeaf {
		t.Error("expected root to not be a leaf when it has children")
	}

	// Child should be a leaf
	isLeaf, err = repo.IsLeafType(ctx, child.ID)
	if err != nil {
		t.Fatalf("failed to check leaf status: %v", err)
	}
	if !isLeaf {
		t.Error("expected child to be a leaf")
	}
}

func TestItemOnlyOnLeafNodes(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create hierarchy: Kitchen > Pantry Items
	kitchen, err := repo.CreateRootType(ctx, "Kitchen", sql.NullString{})
	if err != nil {
		t.Fatalf("failed to create kitchen: %v", err)
	}

	pantry, err := repo.CreateChildType(ctx, kitchen.ID, "Pantry Items", sql.NullString{})
	if err != nil {
		t.Fatalf("failed to create pantry: %v", err)
	}

	// Should be able to create item on leaf node (pantry)
	item, err := repo.CreateItem(ctx, "Rice", pantry.ID)
	if err != nil {
		t.Fatalf("failed to create item on leaf node: %v", err)
	}
	if item.Name != "Rice" {
		t.Errorf("expected item name 'Rice', got '%s'", item.Name)
	}

	// Should NOT be able to create item on non-leaf node (kitchen)
	_, err = repo.CreateItem(ctx, "Invalid Item", kitchen.ID)
	if err == nil {
		t.Error("expected error when creating item on non-leaf node")
	}
}

func TestCannotAddChildToNodeWithItems(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a root type
	kitchen, err := repo.CreateRootType(ctx, "Kitchen", sql.NullString{})
	if err != nil {
		t.Fatalf("failed to create kitchen: %v", err)
	}

	// Add an item to it (it's a leaf node)
	_, err = repo.CreateItem(ctx, "Spoon", kitchen.ID)
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Now try to add a child - should fail because there are items
	_, err = repo.CreateChildType(ctx, kitchen.ID, "Pantry Items", sql.NullString{})
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
	kitchen, _ := repo.CreateRootType(ctx, "Kitchen", sql.NullString{})
	pantry, _ := repo.CreateChildType(ctx, kitchen.ID, "Pantry Items", sql.NullString{})
	_, _ = repo.CreateChildType(ctx, pantry.ID, "Spices", sql.NullString{})
	_, _ = repo.CreateChildType(ctx, kitchen.ID, "Appliances", sql.NullString{})

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
	kitchen, _ := repo.CreateRootType(ctx, "Kitchen", sql.NullString{})
	pantry, _ := repo.CreateChildType(ctx, kitchen.ID, "Pantry Items", sql.NullString{})
	_, _ = repo.CreateChildType(ctx, pantry.ID, "Spices", sql.NullString{})
	_, _ = repo.CreateChildType(ctx, pantry.ID, "Grains", sql.NullString{})

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
	kitchen, _ := repo.CreateRootType(ctx, "Kitchen", sql.NullString{})
	_, _ = repo.CreateChildType(ctx, kitchen.ID, "Pantry Items", sql.NullString{})

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
	kitchen, _ := repo.CreateRootType(ctx, "Kitchen", sql.NullString{})
	_, _ = repo.CreateItem(ctx, "Spoon", kitchen.ID)

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

	kitchen, _ := repo.CreateRootType(ctx, "Kitchen", sql.NullString{})
	_, _ = repo.CreateItem(ctx, "Spoon", kitchen.ID)
	_, _ = repo.CreateItem(ctx, "Fork", kitchen.ID)
	_, _ = repo.CreateItem(ctx, "Knife", kitchen.ID)

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

	kitchen, _ := repo.CreateRootType(ctx, "Kitchen", sql.NullString{})
	_, _ = repo.CreateItem(ctx, "Spoon", kitchen.ID)
	_, _ = repo.CreateItem(ctx, "Fork", kitchen.ID)

	items, err := repo.ListItemsByType(ctx, kitchen.ID)
	if err != nil {
		t.Fatalf("failed to list items: %v", err)
	}

	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestUniqueNamePerParent(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create two root types with same name - should fail
	_, err := repo.CreateRootType(ctx, "Kitchen", sql.NullString{})
	if err != nil {
		t.Fatalf("failed to create first kitchen: %v", err)
	}

	_, err = repo.CreateRootType(ctx, "Kitchen", sql.NullString{})
	if err == nil {
		t.Error("expected error when creating duplicate root type name")
	}

	// Create parent and two children with same name - should fail
	parent, _ := repo.CreateRootType(ctx, "Home", sql.NullString{})
	_, err = repo.CreateChildType(ctx, parent.ID, "Room", sql.NullString{})
	if err != nil {
		t.Fatalf("failed to create first room: %v", err)
	}

	_, err = repo.CreateChildType(ctx, parent.ID, "Room", sql.NullString{})
	if err == nil {
		t.Error("expected error when creating duplicate child type name under same parent")
	}

	// Same name under different parents should be OK
	parent2, _ := repo.CreateRootType(ctx, "Office", sql.NullString{})
	_, err = repo.CreateChildType(ctx, parent2.ID, "Room", sql.NullString{})
	if err != nil {
		t.Errorf("should allow same name under different parents: %v", err)
	}
}
