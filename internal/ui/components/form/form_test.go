package form

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestItemFormSubmission(t *testing.T) {
	// Create item form
	m := NewItemForm(80, 24)

	// Type name
	for _, ch := range "Test Item" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}

	// Tab to quantity
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})

	// Clear default quantity and type new value
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	for _, ch := range "5.0" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}

	// Tab to unit selector
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})

	// Verify focus is at unit selector
	if m.FocusIndex != 2 {
		t.Fatalf("expected FocusIndex to be 2, got %d", m.FocusIndex)
	}

	// Press Enter to submit
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Check if form was submitted
	if !m.Submitted {
		t.Errorf("expected form to be submitted, but Submitted = %v", m.Submitted)
		t.Logf("Error: %s", m.Error)
		t.Logf("Name: %s", m.Inputs[0].Value())
		t.Logf("Quantity: %s", m.Inputs[1].Value())
	}
}
