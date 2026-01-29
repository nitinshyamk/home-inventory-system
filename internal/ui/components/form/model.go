package form

import (
	"strconv"

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
			// For item forms with unit selector
			if m.FormType == FormTypeItem && m.FocusIndex == 2 {
				// On unit selector, Enter submits
				if m.validate() {
					m.Submitted = true
					return m, nil
				}
			} else if m.FocusIndex == len(m.Inputs)-1 {
				// On last text input
				if m.FormType == FormTypeItem {
					// Move to unit selector
					m.FocusIndex = 2
				} else {
					// Submit form
					if m.validate() {
						m.Submitted = true
						return m, nil
					}
				}
			} else {
				// Move to next field
				m.FocusIndex++
				m.updateFocus()
			}

		case "tab":
			m.FocusIndex++
			// For item forms, cycle through 0, 1, 2 (name, quantity, unit selector)
			if m.FormType == FormTypeItem {
				if m.FocusIndex > 2 {
					m.FocusIndex = 0
				}
			} else {
				// For other forms, cycle through text inputs only
				if m.FocusIndex >= len(m.Inputs) {
					m.FocusIndex = 0
				}
			}
			m.updateFocus()

		case "shift+tab":
			m.FocusIndex--
			if m.FocusIndex < 0 {
				if m.FormType == FormTypeItem {
					m.FocusIndex = 2
				} else {
					m.FocusIndex = len(m.Inputs) - 1
				}
			}
			m.updateFocus()

		case "down":
			// On unit selector, change selection
			if m.FormType == FormTypeItem && m.FocusIndex == 2 {
				m.UnitTypeIndex++
				if m.UnitTypeIndex >= len(m.UnitTypes) {
					m.UnitTypeIndex = 0
				}
			} else {
				// Otherwise move to next field
				m.FocusIndex++
				if m.FormType == FormTypeItem && m.FocusIndex > 2 {
					m.FocusIndex = 0
				} else if m.FocusIndex >= len(m.Inputs) {
					m.FocusIndex = 0
				}
				m.updateFocus()
			}

		case "up":
			// On unit selector, change selection
			if m.FormType == FormTypeItem && m.FocusIndex == 2 {
				m.UnitTypeIndex--
				if m.UnitTypeIndex < 0 {
					m.UnitTypeIndex = len(m.UnitTypes) - 1
				}
			} else {
				// Otherwise move to previous field
				m.FocusIndex--
				if m.FocusIndex < 0 {
					if m.FormType == FormTypeItem {
						m.FocusIndex = 2
					} else {
						m.FocusIndex = len(m.Inputs) - 1
					}
				}
				m.updateFocus()
			}
		}
	}

	// Update the focused input (only for text inputs, not unit selector)
	if m.FocusIndex >= 0 && m.FocusIndex < len(m.Inputs) {
		var cmd tea.Cmd
		m.Inputs[m.FocusIndex], cmd = m.Inputs[m.FocusIndex].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the form
func (m Model) View() string {
	switch m.FormType {
	case FormTypeCategory:
		return m.renderCategoryForm()
	case FormTypeItem:
		return m.renderItemForm()
	default:
		return "Unknown form type"
	}
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

	switch m.FormType {
	case FormTypeCategory:
		// Name (first field) is always required
		if len(m.Inputs) > 0 && m.Inputs[0].Value() == "" {
			m.Error = "Name is required"
			return false
		}
		return true

	case FormTypeItem:
		return m.validateItemForm()

	default:
		return false
	}
}

// CategoryData holds typed data for category forms
type CategoryData struct {
	Name        string
	Description string
}

// GetCategoryData returns typed category form data
func (m Model) GetCategoryData() CategoryData {
	var data CategoryData
	if len(m.Inputs) >= 1 {
		data.Name = m.Inputs[0].Value()
	}
	if len(m.Inputs) >= 2 {
		data.Description = m.Inputs[1].Value()
	}
	return data
}

// ItemData holds typed data for item forms
type ItemData struct {
	Name     string
	Quantity float64
	UnitType domain.UnitType
}

// GetItemData returns typed item form data
func (m Model) GetItemData() ItemData {
	var data ItemData
	if len(m.Inputs) >= 1 {
		data.Name = m.Inputs[0].Value()
	}
	if len(m.Inputs) >= 2 {
		data.Quantity, _ = strconv.ParseFloat(m.Inputs[1].Value(), 64)
	}
	if m.UnitTypeIndex >= 0 && m.UnitTypeIndex < len(m.UnitTypes) {
		data.UnitType = m.UnitTypes[m.UnitTypeIndex]
	}
	return data
}
