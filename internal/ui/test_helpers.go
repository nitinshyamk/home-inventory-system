package ui

// Test helper methods for accessing internal state
// These are only used by the testing package

// State returns the current application state
func (m Model) State() AppState {
	return m.state
}

// Error returns the current error if in error state
func (m Model) Error() error {
	return m.err
}
