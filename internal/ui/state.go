package ui

// State management for the application state machine.

// AppState represents the current state of the application
type AppState int

const (
	StateLoading       AppState = iota
	StateBrowsingTypes          // Navigating the type hierarchy
	StateViewingItems           // Viewing items at a leaf node
	StateItemSelected           // Viewing item details
	StateError
)

// String returns human-readable state name for debugging and logging
func (s AppState) String() string {
	switch s {
	case StateLoading:
		return "Loading"
	case StateBrowsingTypes:
		return "BrowsingTypes"
	case StateViewingItems:
		return "ViewingItems"
	case StateItemSelected:
		return "ItemSelected"
	case StateError:
		return "Error"
	default:
		return "Unknown"
	}
}

// CanNavigateUp returns true if the current state allows upward navigation
func (s AppState) CanNavigateUp() bool {
	return s == StateBrowsingTypes || s == StateViewingItems || s == StateItemSelected
}

// CanSelectItem returns true if the current state allows selecting items
func (s AppState) CanSelectItem() bool {
	return s == StateBrowsingTypes || s == StateViewingItems
}

// AllowsItemListDelegation returns true if keypresses should be delegated to itemlist
func (s AppState) AllowsItemListDelegation() bool {
	return s == StateBrowsingTypes || s == StateViewingItems
}

// transitionTo updates the model's state to the specified new state.
// Currently performs no validation, but provides a central location for
// future state transition rules and logging.
func transitionTo(m Model, newState AppState) Model {
	m.state = newState
	return m
}
