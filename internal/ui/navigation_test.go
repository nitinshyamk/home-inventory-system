package ui

import (
	"testing"

	"home-inventory-system/internal/domain"
	"home-inventory-system/internal/ui/components/itemlist"
)

func TestNavigateUpFromItems_HasParent(t *testing.T) {
	handler := setupTestHandler(t)

	m := Model{
		handler: handler,
		state:   StateViewingItems,
		breadcrumb: []domain.ItemType{
			{ID: 1, Name: "Parent"},
			{ID: 2, Name: "Current"},
		},
	}

	m, cmd := navigateUpFromItems(m)

	if cmd == nil {
		t.Error("Expected command to load parent types")
	}

	if m.currentTypeID == nil || *m.currentTypeID != 1 {
		t.Errorf("Expected currentTypeID=1, got %v", m.currentTypeID)
	}
}

func TestNavigateUpFromItems_AtRoot(t *testing.T) {
	handler := setupTestHandler(t)

	m := Model{
		handler:    handler,
		state:      StateViewingItems,
		breadcrumb: []domain.ItemType{{ID: 1, Name: "Root"}},
	}

	m, cmd := navigateUpFromItems(m)

	if cmd == nil {
		t.Error("Expected command to load root types")
	}

	if m.currentTypeID != nil {
		t.Errorf("Expected currentTypeID=nil for root, got %v", m.currentTypeID)
	}
}

func TestNavigateUpFromTypes_AlreadyAtRoot(t *testing.T) {
	m := Model{
		state:         StateBrowsingTypes,
		currentTypeID: nil,
	}

	m, cmd := navigateUpFromTypes(m)

	if cmd != nil {
		t.Error("Expected no command when already at root")
	}
}

func TestNavigateUpFromTypes_DeepNesting(t *testing.T) {
	handler := setupTestHandler(t)

	parentID := int64(1)
	m := Model{
		handler:       handler,
		state:         StateBrowsingTypes,
		currentTypeID: ptrInt64(3),
		breadcrumb: []domain.ItemType{
			{ID: 1, Name: "Root", ParentID: nil},
			{ID: 2, Name: "Parent", ParentID: &parentID},
			{ID: 3, Name: "Current"},
		},
	}

	m, cmd := navigateUpFromTypes(m)

	if cmd == nil {
		t.Error("Expected command to load grandparent types")
	}

	// Should navigate to grandparent (ID=1)
	if m.currentTypeID == nil || *m.currentTypeID != 1 {
		t.Errorf("Expected currentTypeID=1, got %v", m.currentTypeID)
	}
}

func TestSelectType(t *testing.T) {
	handler := setupTestHandler(t)

	// Create itemlist with a selected type
	list := itemlist.New()
	list.SetTypes([]domain.ItemType{
		{ID: 1, Name: "Electronics"},
	})

	m := Model{
		handler:  handler,
		state:    StateBrowsingTypes,
		itemList: list,
	}

	m, cmd := selectType(m)

	if cmd == nil {
		t.Error("Expected command to check leaf type")
	}
}

func TestSelectItem_NoOpForRegularItems(t *testing.T) {
	// Item detail is now shown live in the right pane; Enter on a regular item is a no-op.
	list := itemlist.New()
	list.SetLeafItems([]domain.Item{
		{ID: 1, Name: "Laptop", ItemTypeID: 2},
	})

	m := Model{
		state:    StateViewingItems,
		itemList: list,
	}

	result, cmd := selectItem(m)

	if result.state != StateViewingItems {
		t.Errorf("Expected state=%v (unchanged), got %v", StateViewingItems, result.state)
	}
	if cmd != nil {
		t.Error("Expected no command for regular item selection")
	}
}

func TestGoToRoot(t *testing.T) {
	handler := setupTestHandler(t)

	m := Model{
		handler:       handler,
		currentTypeID: ptrInt64(5),
		breadcrumb:    []domain.ItemType{{ID: 5, Name: "Deep"}},
	}

	m, cmd := goToRoot(m)

	if m.currentTypeID != nil {
		t.Errorf("Expected currentTypeID=nil, got %v", m.currentTypeID)
	}

	if m.breadcrumb != nil {
		t.Error("Expected breadcrumb to be cleared")
	}

	if cmd == nil {
		t.Error("Expected command to load root types")
	}
}

func TestGoToParent(t *testing.T) {
	handler := setupTestHandler(t)

	m := Model{handler: handler}

	m, cmd := goToParent(m, 3)

	if m.currentTypeID == nil || *m.currentTypeID != 3 {
		t.Errorf("Expected currentTypeID=3, got %v", m.currentTypeID)
	}

	if cmd == nil {
		t.Error("Expected command to load child types")
	}
}

func TestNavigateUp_FromDifferentStates(t *testing.T) {
	handler := setupTestHandler(t)

	tests := []struct {
		name         string
		initialState AppState
		hasCmd       bool
	}{
		{
			name:         "from ViewingItems",
			initialState: StateViewingItems,
			hasCmd:       true, // Loads parent/root
		},
		{
			name:         "from BrowsingTypes at root",
			initialState: StateBrowsingTypes,
			hasCmd:       false, // Already at root
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				handler:       handler,
				state:         tt.initialState,
				currentTypeID: nil,
			}

			_, cmd := navigateUp(m)

			hasCmd := cmd != nil
			if hasCmd != tt.hasCmd {
				t.Errorf("Expected hasCmd=%v, got %v", tt.hasCmd, hasCmd)
			}
		})
	}
}

func TestSelectCurrent_FromDifferentStates(t *testing.T) {
	handler := setupTestHandler(t)

	list := itemlist.New()
	list.SetTypes([]domain.ItemType{{ID: 1, Name: "Test"}})

	tests := []struct {
		name         string
		initialState AppState
		list         itemlist.Model
		wantCmd      bool
	}{
		{
			name:         "from BrowsingTypes",
			initialState: StateBrowsingTypes,
			list:         list,
			wantCmd:      true, // Triggers leaf check
		},
		{
			name:         "from Loading (invalid)",
			initialState: StateLoading,
			list:         itemlist.New(),
			wantCmd:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				handler:  handler,
				state:    tt.initialState,
				itemList: tt.list,
			}

			_, cmd := selectCurrent(m)

			hasCmd := cmd != nil
			if hasCmd != tt.wantCmd {
				t.Errorf("Expected hasCmd=%v, got %v", tt.wantCmd, hasCmd)
			}
		})
	}
}

// Helper for creating int64 pointers in tests
func ptrInt64(v int64) *int64 {
	return &v
}
