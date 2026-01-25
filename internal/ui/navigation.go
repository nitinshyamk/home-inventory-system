package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"home-inventory-system/internal/ui/messages"
)

// Navigation logic for hierarchical type and item traversal.

// navigateUp moves up one level in the hierarchy based on current state.
func navigateUp(m Model) (Model, tea.Cmd) {
	switch m.state {
	case StateItemSelected:
		return navigateUpFromItemSelected(m)
	case StateViewingItems:
		return navigateUpFromItems(m)
	case StateBrowsingTypes:
		return navigateUpFromTypes(m)
	default:
		return m, nil
	}
}

// navigateUpFromItemSelected returns to item list view.
func navigateUpFromItemSelected(m Model) (Model, tea.Cmd) {
	m = transitionTo(m, StateViewingItems)
	m.selectedItem = nil
	return m, nil
}

// navigateUpFromItems navigates from item view back to parent type or root.
func navigateUpFromItems(m Model) (Model, tea.Cmd) {
	if parent := getParent(m.breadcrumb); parent != nil {
		return goToParent(m, parent.ID)
	}
	return goToRoot(m)
}

// navigateUpFromTypes navigates up one level in the type hierarchy.
func navigateUpFromTypes(m Model) (Model, tea.Cmd) {
	if m.currentTypeID == nil {
		// Already at root
		return m, nil
	}

	if isAtRoot(m.breadcrumb) {
		// One level down, go to root
		return goToRoot(m)
	}

	// Multiple levels deep - navigate to parent's parent
	// Note: breadcrumb[len-2] is the parent of current location
	if parent := getParent(m.breadcrumb); parent != nil {
		if parent.ParentID != nil {
			// Parent has a parent, go to it
			return goToParent(m, *parent.ParentID)
		}
		// Parent is root-level, go to root
		return goToRoot(m)
	}

	return goToRoot(m)
}

// selectCurrent handles selection of the current item/type.
func selectCurrent(m Model) (Model, tea.Cmd) {
	switch m.state {
	case StateBrowsingTypes:
		return selectType(m)
	case StateViewingItems:
		return selectItem(m)
	default:
		return m, nil
	}
}

// selectType handles selection of a type (triggers leaf check).
func selectType(m Model) (Model, tea.Cmd) {
	selected := m.itemList.SelectedType()
	if selected == nil {
		return m, nil
	}

	return handleTypeSelected(m, messages.TypeSelectedMsg{Type: *selected})
}

// selectItem handles selection of an item (shows details).
func selectItem(m Model) (Model, tea.Cmd) {
	selected := m.itemList.SelectedLeafItem()
	if selected == nil {
		return m, nil
	}

	m.selectedItem = selected
	m = transitionTo(m, StateItemSelected)
	return m, nil
}

// goToRoot navigates back to the root level.
func goToRoot(m Model) (Model, tea.Cmd) {
	m.currentTypeID = nil
	m = clearBreadcrumb(m)
	return m, loadRootTypes(m.handler)
}

// goToParent navigates to a specific parent type.
func goToParent(m Model, parentID int64) (Model, tea.Cmd) {
	m.currentTypeID = &parentID
	return m, loadChildTypes(m.handler, parentID)
}
