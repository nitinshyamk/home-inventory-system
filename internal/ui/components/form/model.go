package form

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"home-inventory-system/internal/domain"
)

// FormType distinguishes between different form types
type FormType int

const (
	FormTypeCategory FormType = iota
	FormTypeItem
)

// Model represents a form component for creating categories or items
type Model struct {
	FormType FormType
	Title    string

	// Text input fields
	Inputs     []textinput.Model
	FocusIndex int

	// For item forms - unit type selection
	UnitTypeIndex int
	UnitTypes     []domain.UnitType

	// Terminal dimensions
	Width  int
	Height int

	// Form state
	Submitted bool
	Cancelled bool

	// Validation error
	Error string
}

// Init initializes the form
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles form updates
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.Cancelled = true
			return m, nil

		case "enter":
			// If we're on the last field, submit
			if m.FocusIndex == len(m.Inputs)-1 {
				if m.validate() {
					m.Submitted = true
					return m, nil
				}
			} else {
				// Move to next field
				m.FocusIndex++
				m.updateFocus()
			}

		case "tab", "down":
			m.FocusIndex++
			if m.FocusIndex >= len(m.Inputs) {
				m.FocusIndex = 0
			}
			m.updateFocus()

		case "shift+tab", "up":
			m.FocusIndex--
			if m.FocusIndex < 0 {
				m.FocusIndex = len(m.Inputs) - 1
			}
			m.updateFocus()
		}
	}

	// Update the focused input
	if m.FocusIndex >= 0 && m.FocusIndex < len(m.Inputs) {
		var cmd tea.Cmd
		m.Inputs[m.FocusIndex], cmd = m.Inputs[m.FocusIndex].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the form
func (m Model) View() string {
	// Basic view - will be enhanced in later stages
	return "Form view placeholder"
}

// updateFocus updates which input field is focused
func (m *Model) updateFocus() {
	for i := range m.Inputs {
		if i == m.FocusIndex {
			m.Inputs[i].Focus()
		} else {
			m.Inputs[i].Blur()
		}
	}
}

// validate checks if the form data is valid
func (m *Model) validate() bool {
	m.Error = ""

	// Name (first field) is always required
	if len(m.Inputs) > 0 && m.Inputs[0].Value() == "" {
		m.Error = "Name is required"
		return false
	}

	return true
}

// GetData returns the form data as a map
func (m Model) GetData() map[string]interface{} {
	data := make(map[string]interface{})

	switch m.FormType {
	case FormTypeCategory:
		if len(m.Inputs) >= 1 {
			data["name"] = m.Inputs[0].Value()
		}
		if len(m.Inputs) >= 2 {
			data["description"] = m.Inputs[1].Value()
		}

	case FormTypeItem:
		if len(m.Inputs) >= 1 {
			data["name"] = m.Inputs[0].Value()
		}
		// Quantity and unit type will be added in later stages
	}

	return data
}
