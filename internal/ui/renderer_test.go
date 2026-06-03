package ui

import (
	"errors"
	"strings"
	"testing"
	"time"

	"home-inventory-system/internal/domain"
	"home-inventory-system/internal/ui/components/itemlist"
)

func TestRenderLoading(t *testing.T) {
	output := renderLoading()

	if !strings.Contains(output, "Loading") {
		t.Errorf("Expected 'Loading' in output, got: %s", output)
	}
}

func TestRenderView_Loading(t *testing.T) {
	m := Model{state: StateLoading}

	output := renderView(m)

	if !strings.Contains(output, "Loading") {
		t.Errorf("Expected loading view, got: %s", output)
	}
}

func TestRenderView_BrowsingTypes(t *testing.T) {
	m := Model{
		state:    StateBrowsingTypes,
		itemList: itemlist.New(),
	}

	output := renderView(m)

	if !strings.Contains(output, "Home") {
		t.Error("Expected breadcrumb showing 'Home'")
	}
}

func TestRenderView_ViewingItems(t *testing.T) {
	m := Model{
		state:    StateViewingItems,
		itemList: itemlist.New(),
	}

	output := renderView(m)

	if !strings.Contains(output, "Home") {
		t.Error("Expected breadcrumb showing 'Home'")
	}
}

func TestRenderItemDetails(t *testing.T) {
	m := Model{
		state: StateItemSelected,
		selectedItem: &domain.Item{
			ID:         1,
			Name:       "Test Laptop",
			ItemTypeID: 2,
			CreatedAt:  time.Now().Format(time.RFC3339),
		},
		breadcrumb: []domain.ItemType{
			{ID: 1, Name: "Electronics"},
			{ID: 2, Name: "Computers"},
		},
		itemList: itemlist.New(),
	}

	output := renderItemDetails(m)

	requiredStrings := []string{
		"Home",
		"Electronics",
		"Computers",
		"Item Details",
		"Test Laptop",
		"Created:",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(output, required) {
			t.Errorf("Expected %q in item details view, got: %s", required, output)
		}
	}
}

func TestRenderItemDetails_NilItem(t *testing.T) {
	m := Model{
		state:        StateItemSelected,
		selectedItem: nil,
		itemList:     itemlist.New(),
	}

	output := renderItemDetails(m)

	// Should fall back to main view
	if !strings.Contains(output, "Home") {
		t.Error("Expected fallback to main view when item is nil")
	}
}

func TestRenderError(t *testing.T) {
	m := Model{
		state: StateError,
		err:   errors.New("database connection failed"),
	}

	output := renderError(m)

	if !strings.Contains(output, "error occurred") {
		t.Error("Expected error message")
	}

	if !strings.Contains(output, "database connection failed") {
		t.Error("Expected specific error text")
	}
}

func TestRenderHelp_BrowsingTypes(t *testing.T) {
	output := renderHelp(StateBrowsingTypes)

	if !strings.Contains(output, "drill down") {
		t.Error("Expected 'drill down' in browsing types help")
	}
}

func TestRenderHelp_ViewingItems(t *testing.T) {
	output := renderHelp(StateViewingItems)

	if !strings.Contains(output, "select") {
		t.Error("Expected 'select' in viewing items help")
	}
}

func TestRenderMainView(t *testing.T) {
	m := Model{
		state:      StateBrowsingTypes,
		breadcrumb: []domain.ItemType{{ID: 1, Name: "Electronics"}},
		itemList:   itemlist.New(),
	}

	output := renderMainView(m)

	// Should contain breadcrumb, content, and help
	if !strings.Contains(output, "Home") {
		t.Error("Expected 'Home' in breadcrumb")
	}

	if !strings.Contains(output, "Electronics") {
		t.Error("Expected 'Electronics' in breadcrumb")
	}

	if !strings.Contains(output, "navigate") {
		t.Error("Expected help text with 'navigate'")
	}
}

func TestRenderSplitPane_Wide(t *testing.T) {
	m := Model{
		state:    StateBrowsingTypes,
		width:    100,
		height:   30,
		itemList: itemlist.New(),
	}
	m.itemList.SetSize(40, 26)

	output := renderView(m)

	if !strings.Contains(output, "│") {
		t.Error("Expected vertical separator │ in wide split-pane layout")
	}
}

func TestRenderSplitPane_Narrow(t *testing.T) {
	m := Model{
		state:    StateBrowsingTypes,
		width:    60,
		height:   24,
		itemList: itemlist.New(),
	}
	m.itemList.SetSize(60, 20)

	output := renderView(m)

	if strings.Contains(output, "│") {
		t.Error("Expected no vertical separator │ in narrow layout")
	}
}

func TestRenderSplitPane_PaneWidths(t *testing.T) {
	const termWidth = 100
	leftWidth, rightWidth, paneHeight := splitPaneDimensions(termWidth, 30)

	if leftWidth <= 0 {
		t.Errorf("Expected positive left pane width, got %d", leftWidth)
	}
	if rightWidth <= 0 {
		t.Errorf("Expected positive right pane width, got %d", rightWidth)
	}
	if leftWidth+rightWidth+1 != termWidth {
		t.Errorf("Pane widths should sum to terminal width: %d + %d + 1 = %d, want %d",
			leftWidth, rightWidth, leftWidth+rightWidth+1, termWidth)
	}
	if paneHeight <= 0 {
		t.Errorf("Expected positive pane height, got %d", paneHeight)
	}
}

func TestRenderView_UnknownState(t *testing.T) {
	m := Model{state: AppState(999)} // Invalid state

	output := renderView(m)

	if output != "" {
		t.Errorf("Expected empty string for unknown state, got: %s", output)
	}
}
