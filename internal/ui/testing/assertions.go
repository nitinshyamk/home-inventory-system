package testing

import (
	"context"
	"strings"
	"testing"

	"home-inventory-system/internal/domain"
	"home-inventory-system/internal/service"
	"home-inventory-system/internal/ui"
)

// AssertState checks that the UI is in the expected state
func (s *Simulator) AssertState(t *testing.T, expected ui.AppState) {
	t.Helper()
	actual := s.model.State()
	if actual != expected {
		t.Errorf("expected state %s, got %s", expected, actual)
	}
}

// AssertListContains checks that the current list contains an item with the given text
func (s *Simulator) AssertListContains(t *testing.T, text string) {
	t.Helper()
	view := s.model.View()
	if !strings.Contains(view, text) {
		t.Errorf("expected list to contain %q, but it was not found in:\n%s", text, view)
	}
}

// AssertListNotContains checks that the current list does not contain an item with the given text
func (s *Simulator) AssertListNotContains(t *testing.T, text string) {
	t.Helper()
	view := s.model.View()
	if strings.Contains(view, text) {
		t.Errorf("expected list to not contain %q, but it was found", text)
	}
}

// AssertCategoryExists checks that a category with the given name exists in the database
func (s *Simulator) AssertCategoryExists(t *testing.T, name string, description string) {
	t.Helper()

	ctx := context.Background()
	result := s.handler.HandleQuery(ctx, service.ListRootTypesQuery{})
	rootResult, ok := result.(service.ListRootTypesResult)
	if !ok || rootResult.Err != nil {
		t.Fatalf("failed to query root types: %v", rootResult.Err)
	}

	// Search in root types
	for _, itemType := range rootResult.Types {
		if itemType.Name == name {
			if description != "" && itemType.Description != description {
				t.Errorf("category %q exists but has description %q, expected %q",
					name, itemType.Description, description)
			}
			return
		}
	}

	t.Errorf("category %q not found in database", name)
}

// AssertCategoryExistsUnder checks that a child category exists under a parent
func (s *Simulator) AssertCategoryExistsUnder(t *testing.T, parentName, childName string) {
	t.Helper()

	ctx := context.Background()

	// Find parent
	result := s.handler.HandleQuery(ctx, service.ListRootTypesQuery{})
	rootResult := result.(service.ListRootTypesResult)

	var parentID int64
	found := false
	for _, itemType := range rootResult.Types {
		if itemType.Name == parentName {
			parentID = itemType.ID
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("parent category %q not found", parentName)
	}

	// Check children
	childResult := s.handler.HandleQuery(ctx, service.ListChildTypesQuery{ParentID: parentID})
	childTypesResult := childResult.(service.ListChildTypesResult)

	for _, itemType := range childTypesResult.Types {
		if itemType.Name == childName {
			return
		}
	}

	t.Errorf("child category %q not found under parent %q", childName, parentName)
}

// AssertItemExists checks that an item with the given name exists
func (s *Simulator) AssertItemExists(t *testing.T, name string, categoryName string) {
	t.Helper()

	ctx := context.Background()

	// Find category
	result := s.handler.HandleQuery(ctx, service.ListRootTypesQuery{})
	rootResult := result.(service.ListRootTypesResult)

	var categoryID int64
	found := false
	for _, itemType := range rootResult.Types {
		if itemType.Name == categoryName {
			categoryID = itemType.ID
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("category %q not found", categoryName)
	}

	// Check items
	itemResult := s.handler.HandleQuery(ctx, service.ListItemsByTypeQuery{TypeID: categoryID})
	itemsResult := itemResult.(service.ListItemsByTypeResult)

	for _, item := range itemsResult.Items {
		if item.Name == name {
			return
		}
	}

	t.Errorf("item %q not found in category %q", name, categoryName)
}

// AssertItemCount checks the total number of items in a category
func (s *Simulator) AssertItemCount(t *testing.T, categoryName string, expectedCount int) {
	t.Helper()

	ctx := context.Background()

	// Find category
	result := s.handler.HandleQuery(ctx, service.ListRootTypesQuery{})
	rootResult := result.(service.ListRootTypesResult)

	var categoryID int64
	found := false
	for _, itemType := range rootResult.Types {
		if itemType.Name == categoryName {
			categoryID = itemType.ID
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("category %q not found", categoryName)
	}

	// Count items
	itemResult := s.handler.HandleQuery(ctx, service.ListItemsByTypeQuery{TypeID: categoryID})
	itemsResult := itemResult.(service.ListItemsByTypeResult)

	actualCount := len(itemsResult.Items)
	if actualCount != expectedCount {
		t.Errorf("expected %d items in category %q, got %d", expectedCount, categoryName, actualCount)
	}
}

// AssertNoError checks that the UI is not in an error state
func (s *Simulator) AssertNoError(t *testing.T) {
	t.Helper()
	if s.model.State() == ui.StateError {
		t.Errorf("UI is in error state: %v", s.model.Error())
	}
}

// AssertError checks that the UI is in an error state
func (s *Simulator) AssertError(t *testing.T, expectedError string) {
	t.Helper()
	if s.model.State() != ui.StateError {
		t.Errorf("expected error state, got %s", s.model.State())
		return
	}

	actualError := s.model.Error()
	if actualError == nil {
		t.Errorf("expected error %q, but no error was set", expectedError)
		return
	}

	if !strings.Contains(actualError.Error(), expectedError) {
		t.Errorf("expected error containing %q, got %q", expectedError, actualError.Error())
	}
}

// DumpView prints the current view (for debugging)
func (s *Simulator) DumpView(t *testing.T) {
	t.Helper()
	view := s.model.View()
	t.Logf("Current view:\n%s\n", view)
}

// GetCategories returns all root categories
func (s *Simulator) GetCategories(t *testing.T) []domain.ItemType {
	t.Helper()

	ctx := context.Background()
	result := s.handler.HandleQuery(ctx, service.ListRootTypesQuery{})
	rootResult, ok := result.(service.ListRootTypesResult)
	if !ok || rootResult.Err != nil {
		t.Fatalf("failed to query root types: %v", rootResult.Err)
	}

	return rootResult.Types
}

// AssertBreadcrumbContains checks that the breadcrumb contains the expected text
func (s *Simulator) AssertBreadcrumbContains(t *testing.T, text string) {
	t.Helper()
	view := s.model.View()
	// Breadcrumb is at the top of the view
	lines := strings.Split(view, "\n")
	if len(lines) == 0 {
		t.Fatal("view is empty")
	}

	if !strings.Contains(lines[0], text) {
		t.Errorf("expected breadcrumb to contain %q, got: %s", text, lines[0])
	}
}
