package ui

import (
	"testing"

	"home-inventory-system/internal/ui/messages"
)

func TestLoadRootTypes(t *testing.T) {
	handler := setupTestHandler(t)

	cmd := loadRootTypes(handler)
	msg := cmd()

	typesMsg, ok := msg.(messages.TypesLoadedMsg)
	if !ok {
		t.Fatalf("Expected TypesLoadedMsg, got %T", msg)
	}

	if typesMsg.Err != nil {
		t.Fatalf("loadRootTypes returned error: %v", typesMsg.Err)
	}

	if len(typesMsg.Types) == 0 {
		t.Error("Expected at least one root type")
	}

	if typesMsg.ParentID != nil {
		t.Errorf("Root types should have nil ParentID, got %v", typesMsg.ParentID)
	}

	// Verify we got Electronics
	found := false
	for _, typ := range typesMsg.Types {
		if typ.Name == "Electronics" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find Electronics in root types")
	}
}

func TestLoadChildTypes(t *testing.T) {
	handler := setupTestHandler(t)

	cmd := loadChildTypes(handler, 1) // Electronics has Computers as child
	msg := cmd()

	typesMsg, ok := msg.(messages.TypesLoadedMsg)
	if !ok {
		t.Fatalf("Expected TypesLoadedMsg, got %T", msg)
	}

	if typesMsg.Err != nil {
		t.Fatalf("loadChildTypes returned error: %v", typesMsg.Err)
	}

	if typesMsg.ParentID == nil || *typesMsg.ParentID != 1 {
		t.Errorf("Expected ParentID=1, got %v", typesMsg.ParentID)
	}

	// Verify we got Computers
	found := false
	for _, typ := range typesMsg.Types {
		if typ.Name == "Computers" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find Computers in child types")
	}
}

func TestLoadItemsForType(t *testing.T) {
	handler := setupTestHandler(t)

	cmd := loadItemsForType(handler, 2) // Computers type
	msg := cmd()

	itemsMsg, ok := msg.(messages.LeafItemsLoadedMsg)
	if !ok {
		t.Fatalf("Expected LeafItemsLoadedMsg, got %T", msg)
	}

	if itemsMsg.Err != nil {
		t.Fatalf("loadItemsForType returned error: %v", itemsMsg.Err)
	}

	if itemsMsg.TypeID != 2 {
		t.Errorf("Expected TypeID=2, got %d", itemsMsg.TypeID)
	}

	if len(itemsMsg.Items) == 0 {
		t.Error("Expected at least one item")
	}

	// Verify we got Laptop
	if itemsMsg.Items[0].Name != "Laptop" {
		t.Errorf("Expected item name 'Laptop', got %q", itemsMsg.Items[0].Name)
	}
}

func TestLoadBreadcrumb(t *testing.T) {
	handler := setupTestHandler(t)

	cmd := loadBreadcrumb(handler, 2) // Computers type
	msg := cmd()

	breadcrumbMsg, ok := msg.(messages.BreadcrumbLoadedMsg)
	if !ok {
		t.Fatalf("Expected BreadcrumbLoadedMsg, got %T", msg)
	}

	if breadcrumbMsg.Err != nil {
		t.Fatalf("loadBreadcrumb returned error: %v", breadcrumbMsg.Err)
	}

	// Should have path: Electronics -> Computers
	if len(breadcrumbMsg.Path) < 2 {
		t.Errorf("Expected at least 2 items in path, got %d", len(breadcrumbMsg.Path))
	}
}

func TestCheckLeafType(t *testing.T) {
	handler := setupTestHandler(t)

	tests := []struct {
		name     string
		typeID   int64
		wantLeaf bool
	}{
		{
			name:     "Electronics (not leaf - has Computers child)",
			typeID:   1,
			wantLeaf: false,
		},
		{
			name:     "Computers (leaf - has items)",
			typeID:   2,
			wantLeaf: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := checkLeafType(handler, tt.typeID)
			msg := cmd()

			leafMsg, ok := msg.(messages.LeafCheckCompleteMsg)
			if !ok {
				t.Fatalf("Expected LeafCheckCompleteMsg, got %T", msg)
			}

			if leafMsg.Err != nil {
				t.Fatalf("checkLeafType returned error: %v", leafMsg.Err)
			}

			if leafMsg.TypeID != tt.typeID {
				t.Errorf("Expected TypeID=%d, got %d", tt.typeID, leafMsg.TypeID)
			}

			if leafMsg.IsLeaf != tt.wantLeaf {
				t.Errorf("Expected IsLeaf=%v, got %v", tt.wantLeaf, leafMsg.IsLeaf)
			}
		})
	}
}
