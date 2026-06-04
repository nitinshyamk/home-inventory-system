package testing

import (
	"context"
	"strings"
	"testing"

	"home-inventory-system/internal/service"
	"home-inventory-system/internal/ui"
)

func TestCreateCategoryAtRoot(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	// At root with empty list, we have:
	// - "+ Add New Category" (index 0)
	// - "+ Add New Item" (index 1)
	// List starts with first item selected, so just press Enter
	sim.SendKeys(KeyEnter) // Open form (should be "+ Add New Category")

	// Verify form opened
	sim.AssertState(t, ui.StateCreatingCategory)
	sim.AssertListContains(t, "Create New Category")

	// Fill in form
	sim.SendKeys(Type("Kitchen"))
	sim.SendKeys(KeyTab) // Move to description
	sim.SendKeys(Type("Cooking and food items"))
	sim.SendKeys(KeyEnter) // Submit

	// Verify category was created
	sim.AssertState(t, ui.StateBrowsingTypes)
	sim.AssertCategoryExists(t, "Kitchen", "Cooking and food items")
	sim.AssertListContains(t, "Kitchen")
	sim.AssertNoError(t)
}

func TestCreateChildCategory(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	// Create parent category first
	ctx := context.Background()
	parentResult := sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{
		Name: "Kitchen",
	}).(service.CreateRootTypeResult)

	sim.Reload()

	// Navigate to Kitchen
	sim.SendKeys(KeyEnter)

	// Should show "+ Add New Category" in empty category
	sim.AssertListContains(t, "+ Add New Category")

	// Open form (first item should be "+ Add New Category")
	sim.SendKeys(KeyEnter)

	// Verify form opened
	sim.AssertState(t, ui.StateCreatingCategory)

	// Fill in form
	sim.SendKeys(Type("Pantry"))
	sim.SendKeys(KeyTab)
	sim.SendKeys(Type("Dry goods storage"))
	sim.SendKeys(KeyEnter)

	// Verify child category was created
	sim.AssertState(t, ui.StateBrowsingTypes)
	sim.AssertCategoryExistsUnder(t, "Kitchen", "Pantry")
	sim.AssertListContains(t, "Pantry")
	sim.AssertNoError(t)

	// Verify parent ID is set correctly
	childTypes := sim.handler.HandleQuery(ctx, service.ListChildTypesQuery{
		ParentID: parentResult.Type.ID,
	}).(service.ListChildTypesResult)

	if len(childTypes.Types) != 1 {
		t.Fatalf("expected 1 child category, got %d", len(childTypes.Types))
	}
	if childTypes.Types[0].Name != "Pantry" {
		t.Errorf("expected child name 'Pantry', got %q", childTypes.Types[0].Name)
	}
}

func TestCreateItemInEmptyCategory(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	// Create parent category
	ctx := context.Background()
	categoryResult := sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{
		Name: "Pantry",
	}).(service.CreateRootTypeResult)

	sim.Reload()

	// Navigate to Pantry
	sim.SendKeys(KeyEnter)

	// Should show both "+ Add New Category" and "+ Add New Item"
	sim.AssertListContains(t, "+ Add New Category")
	sim.AssertListContains(t, "+ Add New Item")

	// Cursor starts at "+ Add New Category", move to "+ Add New Item"
	sim.SendKeys(KeyDown)  // Move to "+ Add New Item"
	sim.SendKeys(KeyEnter) // Open item form

	// Verify item form opened
	sim.AssertState(t, ui.StateCreatingItem)
	sim.AssertListContains(t, "Add New Item")

	// Fill in form
	sim.SendKeys(Type("Rice Bag"))
	sim.SendKeys(KeyTab) // Move to quantity field (has default "1.0")
	// Clear the default value (3 characters: "1.0")
	sim.SendKeys(KeyBackspace, KeyBackspace, KeyBackspace)
	sim.SendKeys(Type("5.0")) // Type new value
	sim.SendKeys(KeyTab)      // Move to unit type selector

	// Debug: check form state before submitting
	t.Logf("Before Enter - State: %s", sim.Model().State())

	// Now at unit selector (focus index 2), default is Count - just submit
	sim.SendKeys(KeyEnter) // Submit form (Enter on unit selector submits)

	// Debug: check form state after submitting
	t.Logf("After Enter - State: %s", sim.Model().State())

	// Verify item was created
	sim.AssertState(t, ui.StateViewingItems)
	sim.AssertItemExists(t, "Rice Bag", "Pantry")
	sim.AssertListContains(t, "Rice Bag")
	sim.AssertNoError(t)

	// Verify item details
	items := sim.handler.HandleQuery(ctx, service.ListItemsByTypeQuery{
		TypeID: categoryResult.Type.ID,
	}).(service.ListItemsByTypeResult)

	if len(items.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items.Items))
	}

	item := items.Items[0]
	if item.Name != "Rice Bag" {
		t.Errorf("expected item name 'Rice Bag', got %q", item.Name)
	}
	if item.Quantity != 5.0 {
		t.Errorf("expected quantity 5.0, got %f", item.Quantity)
	}
	if item.UnitType != "Count" {
		t.Errorf("expected unit type 'Count', got %q", item.UnitType)
	}
}

func TestCancelCategoryForm(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	// Open category creation form (cursor starts on "+ Add New Category")
	sim.SendKeys(KeyEnter)

	// Verify form opened
	sim.AssertState(t, ui.StateCreatingCategory)

	// Type some data
	sim.SendKeys(Type("Kitchen"))

	// Cancel with Esc
	sim.SendKeys(KeyEsc)

	// Should be back at browsing state
	sim.AssertState(t, ui.StateBrowsingTypes)
	sim.AssertNoError(t)

	// Category should not have been created
	categories := sim.GetCategories(t)
	if len(categories) != 0 {
		t.Errorf("expected 0 categories after cancel, got %d", len(categories))
	}
}

func TestCancelItemForm(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	// Create category
	ctx := context.Background()
	sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{
		Name: "Pantry",
	})

	sim.Reload()

	// Navigate to category and open item form
	sim.SendKeys(KeyEnter)
	sim.SendKeys(KeyDown) // Move to "+ Add New Item"
	sim.SendKeys(KeyEnter)

	// Verify form opened
	sim.AssertState(t, ui.StateCreatingItem)

	// Type some data
	sim.SendKeys(Type("Rice"))

	// Cancel
	sim.SendKeys(KeyEsc)

	// Note: Currently returns to ViewingItems even though category is empty
	// This is a known issue that will be fixed in the form refactoring
	sim.AssertState(t, ui.StateViewingItems)
	sim.AssertNoError(t)
}

func TestFormValidationEmptyName(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	// Open category form (cursor starts on "+ Add New Category")
	sim.SendKeys(KeyEnter)

	// Try to submit without entering name
	// Form has focus on name field, pressing Enter moves to description
	sim.SendKeys(KeyEnter) // Move to description field
	sim.SendKeys(KeyEnter) // Try to submit (on last field)

	// Should still be in form state with error
	sim.AssertState(t, ui.StateCreatingCategory)
	sim.AssertListContains(t, "Name is required")
	sim.AssertNoError(t) // Form errors are not app errors
}

func TestFormNavigationWithTab(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	// Open category form (cursor starts on "+ Add New Category")
	sim.SendKeys(KeyEnter)

	// Tab through fields
	sim.SendKeys(Type("Kitchen"))
	sim.SendKeys(KeyTab)
	sim.SendKeys(Type("Description"))

	// Submit
	sim.SendKeys(KeyEnter)

	// Verify both fields were filled
	sim.AssertCategoryExists(t, "Kitchen", "Description")
}

func TestDeleteCategory_Empty(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	ctx := context.Background()
	sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{Name: "Empty"})
	sim.Reload()

	sim.SendKeys(Type("d"))
	sim.AssertState(t, ui.StateDeletingCategory)

	// Confirm deletion of empty category
	sim.SendKeys(KeyEnter)
	sim.AssertState(t, ui.StateBrowsingTypes)
	sim.AssertListNotContains(t, "Empty")
	sim.AssertNoError(t)
}

func TestDeleteCategory_Cancel(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	ctx := context.Background()
	sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{Name: "Keep"})
	sim.Reload()

	sim.SendKeys(Type("d"))
	sim.AssertState(t, ui.StateDeletingCategory)

	sim.SendKeys(KeyTab, KeyEnter) // Tab to Cancel, Enter
	sim.AssertState(t, ui.StateBrowsingTypes)
	sim.AssertListContains(t, "Keep")
	sim.AssertNoError(t)
}

func TestDeleteCategory_ShowsSubcategoryImpact(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	ctx := context.Background()
	kitchenResult := sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{Name: "Kitchen"}).(service.CreateRootTypeResult)
	sim.handler.HandleCommand(ctx, service.CreateChildTypeCommand{ParentID: kitchenResult.Type.ID, Name: "Pantry"})
	sim.Reload()

	sim.SendKeys(Type("d"))
	sim.AssertState(t, ui.StateDeletingCategory)

	// Right pane should show impact info
	view := sim.View()
	if !strings.Contains(view, "subcategor") {
		t.Errorf("expected subcategory impact text in confirmation view, got:\n%s", view)
	}

	// Cancel
	sim.SendKeys(KeyEsc)
	sim.AssertState(t, ui.StateBrowsingTypes)
}

func TestDeleteCategory_RootWithItemsRejection(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	ctx := context.Background()
	catResult := sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{Name: "RootWithItems"}).(service.CreateRootTypeResult)
	sim.handler.HandleCommand(ctx, service.CreateItemCommand{
		Name: "Spoon", TypeID: catResult.Type.ID, Quantity: 1.0, UnitType: "Count",
	})
	sim.Reload()

	sim.SendKeys(Type("d"))
	sim.AssertState(t, ui.StateDeletingCategory)

	// Right pane should show rejection (no Confirm button, just Cancel)
	view := sim.View()
	if !strings.Contains(view, "cannot") && !strings.Contains(view, "Cannot") {
		t.Errorf("expected rejection message in confirmation view, got:\n%s", view)
	}

	// Only Cancel is available — Esc should return to browsing
	sim.SendKeys(KeyEsc)
	sim.AssertState(t, ui.StateBrowsingTypes)
	// Category should still exist
	sim.AssertListContains(t, "RootWithItems")
}

func TestDeleteItem_Confirm(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	ctx := context.Background()
	catResult := sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{Name: "Pantry"}).(service.CreateRootTypeResult)
	sim.handler.HandleCommand(ctx, service.CreateItemCommand{
		Name: "Rice", TypeID: catResult.Type.ID, Quantity: 1.0, UnitType: "Count",
	})
	sim.Reload()

	sim.SendKeys(KeyEnter) // navigate into Pantry
	sim.AssertState(t, ui.StateViewingItems)
	sim.AssertListContains(t, "Rice")

	// Press 'd' to open delete confirmation
	sim.SendKeys(Type("d"))
	sim.AssertState(t, ui.StateDeletingItem)

	// Focus starts at Confirm (0). Press Enter to confirm.
	sim.SendKeys(KeyEnter)

	sim.AssertState(t, ui.StateViewingItems)
	sim.AssertListNotContains(t, "Rice")
	sim.AssertNoError(t)
}

func TestDeleteItem_Cancel(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	ctx := context.Background()
	catResult := sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{Name: "Pantry"}).(service.CreateRootTypeResult)
	sim.handler.HandleCommand(ctx, service.CreateItemCommand{
		Name: "Rice", TypeID: catResult.Type.ID, Quantity: 1.0, UnitType: "Count",
	})
	sim.Reload()

	sim.SendKeys(KeyEnter)
	sim.SendKeys(Type("d"))
	sim.AssertState(t, ui.StateDeletingItem)

	// Tab to Cancel and press Enter
	sim.SendKeys(KeyTab, KeyEnter)

	sim.AssertState(t, ui.StateViewingItems)
	sim.AssertListContains(t, "Rice") // still there
	sim.AssertNoError(t)
}

func TestEditItem_SaveChanges(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	ctx := context.Background()
	catResult := sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{Name: "Pantry"}).(service.CreateRootTypeResult)
	sim.handler.HandleCommand(ctx, service.CreateItemCommand{
		Name: "Rice", TypeID: catResult.Type.ID, Quantity: 1.0, UnitType: "Count",
	})
	sim.Reload()

	sim.SendKeys(KeyEnter) // navigate into Pantry

	sim.SendKeys(Type("e"))
	sim.AssertState(t, ui.StateEditingItem)

	// Append to name field (cursor is at end of "Rice" after SetValue)
	sim.SendKeys(Type(" Bag")) // "Rice Bag"
	// Tab past Qty(1), Unit(2), ChangeCategory(3) to Save(4)
	sim.SendKeys(KeyTab, KeyTab, KeyTab, KeyTab)
	sim.SendKeys(KeyEnter)

	sim.AssertState(t, ui.StateViewingItems)
	sim.AssertListContains(t, "Rice Bag")
	sim.AssertNoError(t)
}

func TestEditItem_Cancel(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	ctx := context.Background()
	catResult := sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{Name: "Pantry"}).(service.CreateRootTypeResult)
	sim.handler.HandleCommand(ctx, service.CreateItemCommand{
		Name: "Rice", TypeID: catResult.Type.ID, Quantity: 1.0, UnitType: "Count",
	})
	sim.Reload()

	sim.SendKeys(KeyEnter) // navigate into Pantry
	sim.SendKeys(Type("e"))
	sim.AssertState(t, ui.StateEditingItem)

	sim.SendKeys(KeyEsc)
	sim.AssertState(t, ui.StateViewingItems)
	sim.AssertListContains(t, "Rice")
	sim.AssertNoError(t)
}

func TestEditItem_UnitTypeSelection(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	ctx := context.Background()
	catResult := sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{Name: "Pantry"}).(service.CreateRootTypeResult)
	sim.handler.HandleCommand(ctx, service.CreateItemCommand{
		Name: "Milk", TypeID: catResult.Type.ID, Quantity: 1.0, UnitType: "Count",
	})
	sim.Reload()

	sim.SendKeys(KeyEnter) // into Pantry
	sim.SendKeys(Type("e"))

	// Tab to Unit selector: Name(0)→Qty(1)→Unit(2) = 2 Tabs
	sim.SendKeys(KeyTab, KeyTab)
	sim.SendKeys(KeyDown, KeyDown) // Count → Grams → Liters
	// Tab to Save: Unit(2)→ChangeCategory(3)→Save(4) = 2 Tabs
	sim.SendKeys(KeyTab, KeyTab, KeyEnter)

	sim.AssertState(t, ui.StateViewingItems)
	items := sim.handler.HandleQuery(ctx, service.ListItemsByTypeQuery{TypeID: catResult.Type.ID}).(service.ListItemsByTypeResult)
	if len(items.Items) != 1 || items.Items[0].UnitType != "Liters" {
		t.Errorf("expected item with unit Liters, got %+v", items.Items)
	}
}

func TestEditItem_DisabledSaveWhenUnchanged(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	ctx := context.Background()
	catResult := sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{Name: "Pantry"}).(service.CreateRootTypeResult)
	sim.handler.HandleCommand(ctx, service.CreateItemCommand{
		Name: "Rice", TypeID: catResult.Type.ID, Quantity: 1.0, UnitType: "Count",
	})
	sim.Reload()

	sim.SendKeys(KeyEnter)
	sim.SendKeys(Type("e"))
	sim.AssertState(t, ui.StateEditingItem)

	// Tab to Save without any changes: Name(0)→Qty(1)→Unit(2)→ChangeCategory(3)→Save(4) = 4 Tabs
	sim.SendKeys(KeyTab, KeyTab, KeyTab, KeyTab, KeyEnter)
	// Save is disabled when unchanged
	sim.AssertState(t, ui.StateEditingItem)
}

func TestEditItem_ChangeCategoryViaInlinePicker(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	ctx := context.Background()
	// Freezer (F) comes before Pantry (P) alphabetically.
	// Peas live in Pantry; we'll move them to Freezer via the picker.
	pantryResult := sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{Name: "Pantry"}).(service.CreateRootTypeResult)
	freezerResult := sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{Name: "Freezer"}).(service.CreateRootTypeResult)
	sim.handler.HandleCommand(ctx, service.CreateItemCommand{
		Name: "Peas", TypeID: pantryResult.Type.ID, Quantity: 1.0, UnitType: "Count",
	})
	sim.Reload()

	// Alphabetical list: Freezer(0), Pantry(1). Navigate Down to Pantry, then Enter.
	sim.SendKeys(KeyDown)  // move to Pantry
	sim.SendKeys(KeyEnter) // drill into Pantry
	sim.SendKeys(Type("e"))
	sim.AssertState(t, ui.StateEditingItem)

	// Tab to Change Category: Name(0)→Qty(1)→Unit(2)→ChangeCategory(3) = 3 Tabs
	sim.SendKeys(KeyTab, KeyTab, KeyTab)
	sim.SendKeys(KeyEnter) // open picker
	sim.AssertState(t, ui.StatePickingCategory)

	// Picker shows [Freezer(0), Pantry(1)] alphabetically.
	// Cursor starts at 0 (Freezer). Just press Enter to select Freezer.
	sim.SendKeys(KeyEnter)
	sim.AssertState(t, ui.StateEditingItem)

	// Tab to Save: focusChangeCategory(3)→Save(4), then Enter
	sim.SendKeys(KeyTab, KeyEnter)

	// Verify item moved to Freezer
	sim.AssertState(t, ui.StateViewingItems)
	freezerItems := sim.handler.HandleQuery(ctx, service.ListItemsByTypeQuery{TypeID: freezerResult.Type.ID}).(service.ListItemsByTypeResult)
	if len(freezerItems.Items) != 1 || freezerItems.Items[0].Name != "Peas" {
		t.Errorf("expected Peas in Freezer, got %+v", freezerItems.Items)
	}
}

func TestEditItem_ChangeCategoryCancel(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	ctx := context.Background()
	pantryResult := sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{Name: "Pantry"}).(service.CreateRootTypeResult)
	sim.handler.HandleCommand(ctx, service.CreateItemCommand{
		Name: "Rice", TypeID: pantryResult.Type.ID, Quantity: 1.0, UnitType: "Count",
	})
	sim.Reload()

	sim.SendKeys(KeyEnter)         // navigate into Pantry
	sim.SendKeys(Type("e"))        // open edit form
	// Tab to ChangeCategory(3) = 3 Tabs, then Enter to open picker
	sim.SendKeys(KeyTab, KeyTab, KeyTab, KeyEnter)
	sim.AssertState(t, ui.StatePickingCategory)

	sim.SendKeys(KeyEsc) // cancel
	sim.AssertState(t, ui.StateEditingItem)
}

func TestHandleUpdateItemCommand_ChangesCategory(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	ctx := context.Background()
	pantryResult := sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{Name: "Pantry"}).(service.CreateRootTypeResult)
	freezerResult := sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{Name: "Freezer"}).(service.CreateRootTypeResult)
	itemResult := sim.handler.HandleCommand(ctx, service.CreateItemCommand{
		Name: "Peas", TypeID: pantryResult.Type.ID, Quantity: 1.0, UnitType: "Count",
	}).(service.CreateItemResult)

	result := sim.handler.HandleCommand(ctx, service.UpdateItemCommand{
		ID: itemResult.Item.ID, Name: "Peas", TypeID: freezerResult.Type.ID,
		Quantity: 1.0, UnitType: "Count",
	}).(service.UpdateItemResult)

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.Item.ItemTypeID != freezerResult.Type.ID {
		t.Errorf("expected item in Freezer (id=%d), got typeID=%d", freezerResult.Type.ID, result.Item.ItemTypeID)
	}
}

func TestEditCategory_SaveChanges(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	ctx := context.Background()
	sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{
		Name:        "Kitchen",
		Description: "Old",
	})
	sim.Reload()

	// Press 'e' to open edit form
	sim.SendKeys(Type("e"))
	sim.AssertState(t, ui.StateEditingCategory)

	// Append to name to make a change, then Tab to Save and submit
	sim.SendKeys(Type(" Updated")) // Name is now "Kitchen Updated"
	sim.SendKeys(KeyTab, KeyTab)   // Tab past Description to Save
	sim.SendKeys(KeyEnter)         // submit

	sim.AssertState(t, ui.StateBrowsingTypes)
	sim.AssertListContains(t, "Kitchen Updated")
	sim.AssertNoError(t)
}

func TestEditCategory_Cancel(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	ctx := context.Background()
	sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{Name: "Kitchen"})
	sim.Reload()

	sim.SendKeys(Type("e"))
	sim.AssertState(t, ui.StateEditingCategory)

	sim.SendKeys(KeyEsc)
	sim.AssertState(t, ui.StateBrowsingTypes)
	sim.AssertListContains(t, "Kitchen")
	sim.AssertNoError(t)
}

func TestEditCategory_ValidationEmptyName(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	ctx := context.Background()
	sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{Name: "Kit"})
	sim.Reload()

	sim.SendKeys(Type("e"))
	sim.AssertState(t, ui.StateEditingCategory)

	// Clear name field ("Kit" = 3 chars, cursor is at end after SetValue)
	sim.SendKeys(KeyBackspace, KeyBackspace, KeyBackspace)
	// Tab to Save and try to submit with empty name
	sim.SendKeys(KeyTab, KeyTab, KeyEnter)

	// Still in edit state because submit() sets error without transitioning
	sim.AssertState(t, ui.StateEditingCategory)
	sim.AssertNoError(t) // app-level error should not be set
}

func TestEditCategory_DisabledSaveWhenUnchanged(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	ctx := context.Background()
	sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{Name: "Kitchen"})
	sim.Reload()

	sim.SendKeys(Type("e"))
	sim.AssertState(t, ui.StateEditingCategory)

	// Tab straight to Save without changing anything
	sim.SendKeys(KeyTab, KeyTab, KeyEnter)

	// Save is disabled when no changes — should still be in edit state
	sim.AssertState(t, ui.StateEditingCategory)
}

func TestItemFormUnitTypeSelection(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	// Create category
	ctx := context.Background()
	categoryResult := sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{
		Name: "Liquids",
	}).(service.CreateRootTypeResult)

	sim.Reload()

	// Navigate and open item form
	sim.SendKeys(KeyEnter)
	sim.SendKeys(KeyDown) // "+ Add New Item"
	sim.SendKeys(KeyEnter)

	// Fill name and quantity
	sim.SendKeys(Type("Milk"))
	sim.SendKeys(KeyTab)
	// Clear default "1.0" and type new value
	sim.SendKeys(KeyBackspace, KeyBackspace, KeyBackspace)
	sim.SendKeys(Type("2.0"))
	sim.SendKeys(KeyTab) // Move to unit selector

	// Cycle through units: Count -> Grams -> Liters
	sim.SendKeys(KeyDown)  // Grams
	sim.SendKeys(KeyDown)  // Liters
	sim.SendKeys(KeyEnter) // Submit

	// Verify item was created with Liters
	items := sim.handler.HandleQuery(ctx, service.ListItemsByTypeQuery{
		TypeID: categoryResult.Type.ID,
	}).(service.ListItemsByTypeResult)

	if len(items.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items.Items))
	}

	if items.Items[0].UnitType != "Liters" {
		t.Errorf("expected unit type 'Liters', got %q", items.Items[0].UnitType)
	}
}
