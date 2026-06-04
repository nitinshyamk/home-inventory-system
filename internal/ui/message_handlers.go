package ui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"home-inventory-system/internal/service"
	"home-inventory-system/internal/ui/components/categorypicker"
	"home-inventory-system/internal/ui/components/itemedit"
	"home-inventory-system/internal/ui/formcontroller"
	"home-inventory-system/internal/ui/messages"
)

// Message handlers transform incoming messages into model updates.
// All handlers follow the pattern: (Model, Msg) -> (Model, Cmd)

// handleWindowResize updates model dimensions when terminal is resized.
func handleWindowResize(m Model, msg tea.WindowSizeMsg) (Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	var listWidth, paneHeight int
	if msg.Width >= splitPaneMinWidth {
		listWidth, _, paneHeight = splitPaneDimensions(msg.Width, msg.Height)
	} else {
		listWidth = msg.Width
		paneHeight = msg.Height - 4
		if paneHeight < 1 {
			paneHeight = 1
		}
	}
	m.itemList.SetSize(listWidth, paneHeight)
	return m, nil
}

// handleTypesLoaded processes loaded item types and updates UI state.
func handleTypesLoaded(m Model, msg messages.TypesLoadedMsg) (Model, tea.Cmd) {
	if msg.Err != nil {
		return setError(m, msg.Err)
	}

	m.itemList.SetTypes(msg.Types)
	m = transitionTo(m, StateBrowsingTypes)
	m.currentTypeID = msg.ParentID

	// Seed right pane with first category detail if available
	m, rightCmd := refreshRightPane(m)

	if msg.ParentID != nil {
		return m, tea.Batch(loadBreadcrumb(m.handler, *msg.ParentID), rightCmd)
	}

	m = clearBreadcrumb(m)
	return m, rightCmd
}

// handleLeafItemsLoaded processes loaded items and switches to viewing state.
func handleLeafItemsLoaded(m Model, msg messages.LeafItemsLoadedMsg) (Model, tea.Cmd) {
	if msg.Err != nil {
		return setError(m, msg.Err)
	}

	m.items = msg.Items
	m.itemList.SetLeafItems(msg.Items)
	m = transitionTo(m, StateViewingItems)
	m, _ = refreshRightPane(m) // synchronous for items, no cmd needed
	return m, nil
}

// handleBreadcrumbLoaded updates breadcrumb trail (non-fatal if fails).
func handleBreadcrumbLoaded(m Model, msg messages.BreadcrumbLoadedMsg) (Model, tea.Cmd) {
	if msg.Err != nil {
		// Non-fatal, just don't show breadcrumb
		return m, nil
	}
	m = setBreadcrumb(m, msg.Path)
	return m, nil
}

// handleTypeSelected initiates async leaf check when user selects a type.
// This replaces the previous synchronous query that blocked the UI.
func handleTypeSelected(m Model, msg messages.TypeSelectedMsg) (Model, tea.Cmd) {
	// Start async leaf check instead of blocking UI
	return m, checkLeafType(m.handler, msg.Type.ID)
}

// handleLeafCheckComplete processes leaf check result and loads appropriate data.
// If leaf: load items. If not leaf: load child types.
func handleLeafCheckComplete(m Model, msg messages.LeafCheckCompleteMsg) (Model, tea.Cmd) {
	if msg.Err != nil {
		return setError(m, msg.Err)
	}

	// Update current position
	m.currentTypeID = &msg.TypeID

	if msg.IsLeaf {
		// Load items at this leaf node
		return m, tea.Batch(
			loadItemsForType(m.handler, msg.TypeID),
			loadBreadcrumb(m.handler, msg.TypeID),
		)
	}

	// Not a leaf, load child types
	return m, loadChildTypes(m.handler, msg.TypeID)
}

// setError transitions the model to error state.
func setError(m Model, err error) (Model, tea.Cmd) {
	m = transitionTo(m, StateError)
	m.err = err
	return m, nil
}

// delegateToItemList passes messages to the itemlist component and refreshes
// the right pane detail for the newly highlighted row.
func delegateToItemList(m Model, msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.itemList, cmd = m.itemList.Update(msg)
	m, detailCmd := refreshRightPane(m)
	return m, tea.Batch(cmd, detailCmd)
}

// refreshRightPane updates right pane state based on the current itemList selection.
// For items, this is synchronous. For categories, it fires an async count load.
func refreshRightPane(m Model) (Model, tea.Cmd) {
	if selectedType := m.itemList.SelectedType(); selectedType != nil {
		m.rightPane = rightPaneContent{category: selectedType}
		return m, loadCategoryDetail(m.handler, selectedType.ID)
	}
	if selectedItem := m.itemList.SelectedLeafItem(); selectedItem != nil {
		m.rightPane = rightPaneContent{item: selectedItem, itemTypePath: m.breadcrumb}
		return m, nil
	}
	m.rightPane = rightPaneContent{}
	return m, nil
}

// delegateToCategoryForm passes messages to the category form component.
func delegateToCategoryForm(m Model, msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.categoryForm, cmd = m.categoryForm.Update(msg)

	// Check if form was submitted or cancelled
	if m.categoryForm.Submitted() {
		ctx := formcontroller.FormContext{
			CurrentTypeID: m.currentTypeID,
			Width:         m.width,
			Height:        m.height,
		}
		return m, m.formController.SubmitCategoryForm(m.categoryForm.GetData(), ctx)
	}

	if m.categoryForm.Cancelled() {
		return m, func() tea.Msg {
			return messages.FormCancelledMsg{}
		}
	}

	return m, cmd
}

// delegateToItemForm passes messages to the item form component.
func delegateToItemForm(m Model, msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.itemForm, cmd = m.itemForm.Update(msg)

	// Check if form was submitted or cancelled
	if m.itemForm.Submitted() {
		ctx := formcontroller.FormContext{
			CurrentTypeID: m.currentTypeID,
			Width:         m.width,
			Height:        m.height,
		}
		return m, m.formController.SubmitItemForm(m.itemForm.GetData(), ctx)
	}

	if m.itemForm.Cancelled() {
		return m, func() tea.Msg {
			return messages.FormCancelledMsg{}
		}
	}

	return m, cmd
}

// handleFormCancelled returns to the previous state when form is cancelled.
func handleFormCancelled(m Model) (Model, tea.Cmd) {
	switch m.state {
	case StateCreatingCategory:
		// Return to browsing types
		m = transitionTo(m, StateBrowsingTypes)
	case StateCreatingItem:
		// Return to viewing items
		m = transitionTo(m, StateViewingItems)
	}
	return m, nil
}

// delegateToCategoryEditForm passes key events to the category edit form in the right pane.
func delegateToCategoryEditForm(m Model, msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.categoryEditForm, cmd = m.categoryEditForm.Update(msg)

	if m.categoryEditForm.Submitted() {
		return m, submitCategoryEdit(m.handler, m.rightPane.category.ID, m.categoryEditForm.Name(), m.categoryEditForm.Description())
	}
	if m.categoryEditForm.Cancelled() {
		m = transitionTo(m, StateBrowsingTypes)
	}
	return m, cmd
}

// submitCategoryEdit sends an UpdateCategoryCommand and emits CategoryUpdatedMsg.
func submitCategoryEdit(handler *service.Handler, id int64, name, description string) tea.Cmd {
	return func() tea.Msg {
		result, ok := handler.HandleCommand(context.Background(), service.UpdateCategoryCommand{
			ID:          id,
			Name:        name,
			Description: description,
		}).(service.UpdateCategoryResult)
		if !ok {
			return messages.CategoryUpdatedMsg{Err: fmt.Errorf("unexpected result type")}
		}
		if result.ValidationError != nil {
			return messages.CategoryUpdatedMsg{Err: result.ValidationError}
		}
		return messages.CategoryUpdatedMsg{Category: result.Type, Err: result.Err}
	}
}

// handleCategoryUpdated processes the result of a category update.
func handleCategoryUpdated(m Model, msg messages.CategoryUpdatedMsg) (Model, tea.Cmd) {
	if msg.Err != nil {
		// Show the error inline in the form (non-fatal)
		m.categoryEditForm.SetError(msg.Err.Error())
		return m, nil
	}
	// Return to browsing and refresh the list
	m = transitionTo(m, StateBrowsingTypes)
	if m.currentTypeID == nil {
		return m, loadRootTypes(m.handler)
	}
	return m, loadChildTypes(m.handler, *m.currentTypeID)
}

// delegateToItemEditForm passes key events to the item edit form in the right pane.
func delegateToItemEditForm(m Model, msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.itemEditForm, cmd = m.itemEditForm.Update(msg)

	if m.itemEditForm.ChangingCategory() {
		m.itemEditForm.ClearChangingCategory()
		m = transitionTo(m, StatePickingCategory)
		return m, loadLeafTypesWithPaths(m.handler)
	}
	if m.itemEditForm.Submitted() {
		return m, submitItemEdit(m.handler, m.rightPane.item.ID, m.itemEditForm)
	}
	if m.itemEditForm.Cancelled() {
		m = transitionTo(m, StateViewingItems)
	}
	return m, cmd
}

// delegateToCategoryPicker passes key events to the inline category picker.
func delegateToCategoryPicker(m Model, msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.categoryPicker, cmd = m.categoryPicker.Update(msg)

	if entry := m.categoryPicker.Selected(); entry != nil {
		// Apply selection to item edit form
		m.itemEditForm.SetCurrentType(entry.TypeID)
		m.itemEditForm.SetCategoryPath(entry.FullPath)
		m = transitionTo(m, StateEditingItem)
		return m, cmd
	}
	if m.categoryPicker.Cancelled() {
		m = transitionTo(m, StateEditingItem)
	}
	return m, cmd
}

// handleLeafTypesWithPathsLoaded creates the category picker from loaded data.
func handleLeafTypesWithPathsLoaded(m Model, msg messages.LeafTypesWithPathsLoadedMsg) (Model, tea.Cmd) {
	if msg.Err != nil {
		// Non-fatal: go back to item edit
		m = transitionTo(m, StateEditingItem)
		return m, nil
	}
	entries := categorypicker.BuildEntries(msg.LeafTypes, msg.Paths)
	m.categoryPicker = categorypicker.New(entries)
	return m, nil
}

// submitItemEdit sends an UpdateItemCommand and emits ItemUpdatedMsg.
func submitItemEdit(handler *service.Handler, id int64, form itemedit.Model) tea.Cmd {
	return func() tea.Msg {
		result, ok := handler.HandleCommand(context.Background(), service.UpdateItemCommand{
			ID:       id,
			Name:     form.Name(),
			TypeID:   form.TypeID(),
			Quantity: form.Quantity(),
			UnitType: form.UnitType(),
		}).(service.UpdateItemResult)
		if !ok {
			return messages.ItemUpdatedMsg{Err: fmt.Errorf("unexpected result type")}
		}
		if result.ValidationError != nil {
			return messages.ItemUpdatedMsg{Err: result.ValidationError}
		}
		return messages.ItemUpdatedMsg{Item: result.Item, Err: result.Err}
	}
}

// handleItemUpdated processes the result of an item update.
func handleItemUpdated(m Model, msg messages.ItemUpdatedMsg) (Model, tea.Cmd) {
	if msg.Err != nil {
		m.itemEditForm.SetError(msg.Err.Error())
		return m, nil
	}
	m = transitionTo(m, StateViewingItems)
	if m.currentTypeID == nil {
		return m, nil
	}
	return m, loadItemsForType(m.handler, *m.currentTypeID)
}

// handleDeleteItemConfirmKey handles key events in the item delete confirmation state.
func handleDeleteItemConfirmKey(m Model, msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch keyMsg.String() {
	case "esc":
		m = transitionTo(m, StateViewingItems)
	case "tab", "shift+tab":
		m.deleteConfirmFocus = 1 - m.deleteConfirmFocus // toggle 0↔1
	case "enter":
		if m.deleteConfirmFocus == 0 { // Confirm
			item := m.rightPane.item
			return m, deleteItem(m.handler, item.ID, item.ItemTypeID)
		}
		m = transitionTo(m, StateViewingItems)
	}
	return m, nil
}

// deleteItem sends a DeleteItemCommand and emits ItemDeletedMsg.
func deleteItem(handler *service.Handler, id int64, typeID int64) tea.Cmd {
	return func() tea.Msg {
		result, ok := handler.HandleCommand(context.Background(), service.DeleteItemCommand{ID: id}).(service.DeleteItemResult)
		if !ok {
			return messages.ItemDeletedMsg{Err: fmt.Errorf("unexpected result type")}
		}
		return messages.ItemDeletedMsg{DeletedID: id, TypeID: typeID, Err: result.Err}
	}
}

// handleCategoryChildSummaryLoaded stores the loaded summary for the delete confirmation.
func handleCategoryChildSummaryLoaded(m Model, msg messages.CategoryChildSummaryLoadedMsg) (Model, tea.Cmd) {
	if msg.Err != nil {
		return setError(m, msg.Err)
	}
	m.deleteSummary = &deleteCategorySummary{
		childTypeCount: msg.ChildTypeCount,
		itemCount:      msg.ItemCount,
		parentID:       msg.ParentID,
	}
	return m, nil
}

// handleDeleteCategoryConfirmKey handles key events in the category delete confirmation state.
func handleDeleteCategoryConfirmKey(m Model, msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	// If the category has root-level items, only Cancel is available
	isRootWithItems := m.deleteSummary != nil &&
		m.deleteSummary.parentID == nil &&
		m.deleteSummary.itemCount > 0

	switch keyMsg.String() {
	case "esc":
		m = transitionTo(m, StateBrowsingTypes)
	case "tab", "shift+tab":
		if !isRootWithItems {
			m.deleteConfirmFocus = 1 - m.deleteConfirmFocus
		}
	case "enter":
		if isRootWithItems || m.deleteConfirmFocus == 1 { // Cancel
			m = transitionTo(m, StateBrowsingTypes)
			return m, nil
		}
		// Confirm Delete (Task 7: only works for empty categories; lift handled in Task 8)
		cat := m.rightPane.category
		return m, deleteCategorySimple(m.handler, cat.ID, m.deleteSummary)
	}
	return m, nil
}

// deleteCategorySimple uses the existing DeleteItemTypeCommand (no lift).
// For Task 7, this handles the empty category case. Lift is added in Task 8.
func deleteCategorySimple(handler *service.Handler, id int64, summary *deleteCategorySummary) tea.Cmd {
	return func() tea.Msg {
		var parentID *int64
		if summary != nil {
			parentID = summary.parentID
		}
		result, ok := handler.HandleCommand(context.Background(), service.DeleteItemTypeCommand{ID: id}).(service.DeleteItemTypeResult)
		if !ok {
			return messages.CategoryDeletedMsg{Err: fmt.Errorf("unexpected result type")}
		}
		return messages.CategoryDeletedMsg{DeletedID: id, ParentID: parentID, Err: result.Err}
	}
}

// handleCategoryDeleted processes the result of a category delete.
func handleCategoryDeleted(m Model, msg messages.CategoryDeletedMsg) (Model, tea.Cmd) {
	if msg.Err != nil {
		// Show error inline in right pane (non-fatal: stay in DeletingCategory)
		// For now, surface it as a setError; Task 8 will improve this.
		return setError(m, msg.Err)
	}
	m = transitionTo(m, StateBrowsingTypes)
	m.rightPane = rightPaneContent{}
	m.deleteSummary = nil
	if msg.ParentID == nil {
		return m, loadRootTypes(m.handler)
	}
	return m, loadChildTypes(m.handler, *msg.ParentID)
}

// handleItemDeleted processes the result of an item delete.
func handleItemDeleted(m Model, msg messages.ItemDeletedMsg) (Model, tea.Cmd) {
	if msg.Err != nil {
		return setError(m, msg.Err)
	}
	m = transitionTo(m, StateViewingItems)
	m.rightPane = rightPaneContent{} // clear right pane
	if m.currentTypeID == nil {
		return m, nil
	}
	return m, loadItemsForType(m.handler, *m.currentTypeID)
}

// handleRightPaneDetail updates the right pane with loaded category count data.
func handleRightPaneDetail(m Model, msg messages.RightPaneDetailMsg) (Model, tea.Cmd) {
	if msg.Err != nil {
		return m, nil // non-fatal: keep existing right pane content
	}
	// Only apply if the right pane is still showing the same category
	if m.rightPane.category != nil && m.rightPane.category.ID == msg.CategoryID {
		m.rightPane.isLeaf = msg.IsLeaf
		m.rightPane.childCount = msg.ChildCount
		m.rightPane.itemCount = msg.ItemCount
	}
	return m, nil
}

// handleCategoryCreated processes category creation result and refreshes the list.
func handleCategoryCreated(m Model, msg messages.CategoryCreatedMsg) (Model, tea.Cmd) {
	if msg.Err != nil {
		return setError(m, msg.Err)
	}

	// Reload the category list to show the new category
	if m.currentTypeID == nil {
		// At root level
		m = transitionTo(m, StateBrowsingTypes)
		return m, loadRootTypes(m.handler)
	}

	// At child level
	m = transitionTo(m, StateBrowsingTypes)
	return m, loadChildTypes(m.handler, *m.currentTypeID)
}

// handleItemCreated processes item creation result and refreshes the list.
func handleItemCreated(m Model, msg messages.ItemCreatedMsg) (Model, tea.Cmd) {
	if msg.Err != nil {
		return setError(m, msg.Err)
	}

	// Reload the item list to show the new item
	if m.currentTypeID == nil {
		return setError(m, fmt.Errorf("unexpected state: no current type for items"))
	}

	m = transitionTo(m, StateViewingItems)
	return m, loadItemsForType(m.handler, *m.currentTypeID)
}
