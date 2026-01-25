package ui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"home-inventory-system/internal/domain"
	"home-inventory-system/internal/ui/components/itemlist"
	"home-inventory-system/internal/ui/messages"
)

func TestHandleWindowResize(t *testing.T) {
	m := Model{
		width:    80,
		height:   24,
		itemList: itemlist.New(),
	}

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	m, _ = handleWindowResize(m, msg)

	if m.width != 120 {
		t.Errorf("Expected width=120, got %d", m.width)
	}
	if m.height != 40 {
		t.Errorf("Expected height=40, got %d", m.height)
	}
}

func TestHandleTypesLoaded_Success(t *testing.T) {
	m := Model{
		state:    StateLoading,
		itemList: itemlist.New(),
	}

	msg := messages.TypesLoadedMsg{
		Types: []domain.ItemType{
			{ID: 1, Name: "Electronics"},
			{ID: 2, Name: "Books"},
		},
		ParentID: nil,
		Err:      nil,
	}

	m, _ = handleTypesLoaded(m, msg)

	if m.state != StateBrowsingTypes {
		t.Errorf("Expected state=%v, got %v", StateBrowsingTypes, m.state)
	}

	if m.currentTypeID != nil {
		t.Errorf("Expected currentTypeID=nil for root, got %v", m.currentTypeID)
	}
}

func TestHandleTypesLoaded_WithParent(t *testing.T) {
	handler := setupTestHandler(t)

	parentID := int64(1)
	m := Model{
		state:    StateLoading,
		itemList: itemlist.New(),
		handler:  handler,
	}

	msg := messages.TypesLoadedMsg{
		Types: []domain.ItemType{
			{ID: 2, Name: "Computers", ParentID: &parentID},
		},
		ParentID: &parentID,
		Err:      nil,
	}

	m, cmd := handleTypesLoaded(m, msg)

	if m.state != StateBrowsingTypes {
		t.Errorf("Expected state=%v, got %v", StateBrowsingTypes, m.state)
	}

	if m.currentTypeID == nil || *m.currentTypeID != 1 {
		t.Errorf("Expected currentTypeID=1, got %v", m.currentTypeID)
	}

	if cmd == nil {
		t.Error("Expected command to load breadcrumb")
	}
}

func TestHandleTypesLoaded_Error(t *testing.T) {
	m := Model{state: StateLoading, itemList: itemlist.New()}

	msg := messages.TypesLoadedMsg{
		Types: nil,
		Err:   errors.New("database error"),
	}

	m, _ = handleTypesLoaded(m, msg)

	if m.state != StateError {
		t.Errorf("Expected state=Error, got %v", m.state)
	}

	if m.err == nil {
		t.Error("Expected error to be set")
	}
}

func TestHandleLeafItemsLoaded_Success(t *testing.T) {
	m := Model{
		state:    StateBrowsingTypes,
		itemList: itemlist.New(),
	}

	msg := messages.LeafItemsLoadedMsg{
		TypeID: 5,
		Items: []domain.Item{
			{ID: 1, Name: "Item 1", ItemTypeID: 5},
			{ID: 2, Name: "Item 2", ItemTypeID: 5},
		},
		Err: nil,
	}

	m, _ = handleLeafItemsLoaded(m, msg)

	if m.state != StateViewingItems {
		t.Errorf("Expected state=%v, got %v", StateViewingItems, m.state)
	}

	if len(m.items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(m.items))
	}
}

func TestHandleLeafItemsLoaded_Error(t *testing.T) {
	m := Model{state: StateBrowsingTypes, itemList: itemlist.New()}

	msg := messages.LeafItemsLoadedMsg{
		TypeID: 5,
		Items:  nil,
		Err:    errors.New("query failed"),
	}

	m, _ = handleLeafItemsLoaded(m, msg)

	if m.state != StateError {
		t.Errorf("Expected state=Error, got %v", m.state)
	}
}

func TestHandleBreadcrumbLoaded_Success(t *testing.T) {
	m := Model{breadcrumb: nil}

	path := []domain.ItemType{
		{ID: 1, Name: "Parent"},
		{ID: 2, Name: "Child"},
	}

	msg := messages.BreadcrumbLoadedMsg{
		Path: path,
		Err:  nil,
	}

	m, _ = handleBreadcrumbLoaded(m, msg)

	if len(m.breadcrumb) != 2 {
		t.Errorf("Expected breadcrumb length=2, got %d", len(m.breadcrumb))
	}
}

func TestHandleBreadcrumbLoaded_ErrorNonFatal(t *testing.T) {
	m := Model{breadcrumb: nil}

	msg := messages.BreadcrumbLoadedMsg{
		Path: nil,
		Err:  errors.New("failed to load path"),
	}

	m, _ = handleBreadcrumbLoaded(m, msg)

	// Should not transition to error state (non-fatal)
	if m.state == StateError {
		t.Error("Breadcrumb error should be non-fatal")
	}

	// Breadcrumb should remain unchanged
	if m.breadcrumb != nil {
		t.Error("Breadcrumb should remain nil on error")
	}
}

func TestHandleTypeSelected(t *testing.T) {
	handler := setupTestHandler(t)

	m := Model{
		handler: handler,
		state:   StateBrowsingTypes,
	}

	msg := messages.TypeSelectedMsg{
		Type: domain.ItemType{ID: 1, Name: "Electronics"},
	}

	m, cmd := handleTypeSelected(m, msg)

	if cmd == nil {
		t.Error("Expected command to check leaf type")
	}

	// Execute the command to verify it returns LeafCheckCompleteMsg
	resultMsg := cmd()
	if _, ok := resultMsg.(messages.LeafCheckCompleteMsg); !ok {
		t.Errorf("Expected LeafCheckCompleteMsg, got %T", resultMsg)
	}
}

func TestHandleLeafCheckComplete_Leaf(t *testing.T) {
	handler := setupTestHandler(t)

	m := Model{
		handler: handler,
		state:   StateBrowsingTypes,
	}

	msg := messages.LeafCheckCompleteMsg{
		TypeID: 2, // Computers (leaf)
		IsLeaf: true,
		Err:    nil,
	}

	m, cmd := handleLeafCheckComplete(m, msg)

	if m.currentTypeID == nil || *m.currentTypeID != 2 {
		t.Errorf("Expected currentTypeID=2, got %v", m.currentTypeID)
	}

	if cmd == nil {
		t.Error("Expected batch command to load items and breadcrumb")
	}
}

func TestHandleLeafCheckComplete_NotLeaf(t *testing.T) {
	handler := setupTestHandler(t)

	m := Model{
		handler: handler,
		state:   StateBrowsingTypes,
	}

	msg := messages.LeafCheckCompleteMsg{
		TypeID: 1, // Electronics (not leaf)
		IsLeaf: false,
		Err:    nil,
	}

	m, cmd := handleLeafCheckComplete(m, msg)

	if m.currentTypeID == nil || *m.currentTypeID != 1 {
		t.Errorf("Expected currentTypeID=1, got %v", m.currentTypeID)
	}

	if cmd == nil {
		t.Error("Expected command to load child types")
	}
}

func TestHandleLeafCheckComplete_Error(t *testing.T) {
	m := Model{state: StateBrowsingTypes}

	msg := messages.LeafCheckCompleteMsg{
		TypeID: 1,
		IsLeaf: false,
		Err:    errors.New("query failed"),
	}

	m, _ = handleLeafCheckComplete(m, msg)

	if m.state != StateError {
		t.Errorf("Expected state=Error, got %v", m.state)
	}
}

func TestSetError(t *testing.T) {
	m := Model{state: StateLoading}

	testErr := errors.New("test error")
	m, _ = setError(m, testErr)

	if m.state != StateError {
		t.Errorf("Expected state=Error, got %v", m.state)
	}

	if m.err != testErr {
		t.Errorf("Expected error=%v, got %v", testErr, m.err)
	}
}

func TestDelegateToItemList(t *testing.T) {
	m := Model{
		itemList: itemlist.New(),
	}

	// Send a key message
	msg := tea.KeyMsg{Type: tea.KeyDown}
	m, cmd := delegateToItemList(m, msg)

	// Command may be nil or non-nil depending on itemlist implementation
	_ = cmd

	// Just verify no panic occurred
	if m.itemList.Mode() != itemlist.ModeTypes {
		// itemlist should be initialized in Types mode
	}
}
