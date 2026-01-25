package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"home-inventory-system/internal/ui/messages"
)

// Message handlers transform incoming messages into model updates.
// All handlers follow the pattern: (Model, Msg) -> (Model, Cmd)

// handleWindowResize updates model dimensions when terminal is resized.
func handleWindowResize(m Model, msg tea.WindowSizeMsg) (Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.itemList.SetSize(msg.Width, msg.Height-6) // Leave room for breadcrumb and help
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

	// Load breadcrumb if not at root
	if msg.ParentID != nil {
		return m, loadBreadcrumb(m.handler, *msg.ParentID)
	}

	m = clearBreadcrumb(m)
	return m, nil
}

// handleLeafItemsLoaded processes loaded items and switches to viewing state.
func handleLeafItemsLoaded(m Model, msg messages.LeafItemsLoadedMsg) (Model, tea.Cmd) {
	if msg.Err != nil {
		return setError(m, msg.Err)
	}

	m.items = msg.Items
	m.itemList.SetLeafItems(msg.Items)
	m = transitionTo(m, StateViewingItems)
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

// delegateToItemList passes messages to the itemlist component for handling.
func delegateToItemList(m Model, msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.itemList, cmd = m.itemList.Update(msg)
	return m, cmd
}
