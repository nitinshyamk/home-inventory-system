package itemform

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"home-inventory-system/internal/domain"
)

// Model represents an item creation form
type Model struct {
	nameInput     textinput.Model
	quantityInput textinput.Model
	unitTypes     []domain.UnitType
	unitTypeIndex int
	focusIndex    int
	error         string
	submitted     bool
	cancelled     bool
	width         int
	height        int
}

// New creates a new item form
func New(width, height int) Model {
	nameInput := textinput.New()
	nameInput.Placeholder = "e.g., Rice Bag"
	nameInput.Focus()
	nameInput.CharLimit = 100
	nameInput.Width = 50

	quantityInput := textinput.New()
	quantityInput.Placeholder = "1.0"
	quantityInput.CharLimit = 10
	quantityInput.Width = 20
	quantityInput.SetValue("1.0")

	return Model{
		nameInput:     nameInput,
		quantityInput: quantityInput,
		unitTypes:     []domain.UnitType{domain.UnitTypeCount, domain.UnitTypeGrams, domain.UnitTypeLiters},
		unitTypeIndex: 0,
		focusIndex:    0,
		width:         width,
		height:        height,
	}
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
			m.cancelled = true
			return m, nil

		case "enter":
			if m.focusIndex == 2 {
				// On unit selector, submit
				if m.validate() {
					m.submitted = true
					return m, nil
				}
			} else if m.focusIndex == 1 {
				// On quantity, move to unit selector
				m.focusIndex = 2
				m.updateFocus()
			} else {
				// Move to next field
				m.focusIndex++
				m.updateFocus()
			}

		case "tab":
			m.focusIndex++
			if m.focusIndex > 2 {
				m.focusIndex = 0
			}
			m.updateFocus()

		case "shift+tab":
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = 2
			}
			m.updateFocus()

		case "down":
			if m.focusIndex == 2 {
				// On unit selector, change selection
				m.unitTypeIndex++
				if m.unitTypeIndex >= len(m.unitTypes) {
					m.unitTypeIndex = 0
				}
			} else {
				// Move to next field
				m.focusIndex++
				if m.focusIndex > 2 {
					m.focusIndex = 0
				}
				m.updateFocus()
			}

		case "up":
			if m.focusIndex == 2 {
				// On unit selector, change selection
				m.unitTypeIndex--
				if m.unitTypeIndex < 0 {
					m.unitTypeIndex = len(m.unitTypes) - 1
				}
			} else {
				// Move to previous field
				m.focusIndex--
				if m.focusIndex < 0 {
					m.focusIndex = 2
				}
				m.updateFocus()
			}
		}
	}

	// Update the focused text input (not unit selector)
	if m.focusIndex < 2 {
		var cmd tea.Cmd
		if m.focusIndex == 0 {
			m.nameInput, cmd = m.nameInput.Update(msg)
		} else {
			m.quantityInput, cmd = m.quantityInput.Update(msg)
		}
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// updateFocus updates which input field is focused
func (m *Model) updateFocus() {
	if m.focusIndex == 0 {
		m.nameInput.Focus()
		m.quantityInput.Blur()
	} else if m.focusIndex == 1 {
		m.nameInput.Blur()
		m.quantityInput.Focus()
	} else {
		m.nameInput.Blur()
		m.quantityInput.Blur()
	}
}

// validate checks if the form data is valid
func (m *Model) validate() bool {
	m.error = ""

	// Name is required
	if m.nameInput.Value() == "" {
		m.error = "Name is required"
		return false
	}

	// Quantity is required and must be valid number
	quantityStr := m.quantityInput.Value()
	if quantityStr == "" {
		m.error = "Quantity is required"
		return false
	}

	quantity, err := strconv.ParseFloat(quantityStr, 64)
	if err != nil {
		m.error = "Quantity must be a valid number"
		return false
	}

	if quantity <= 0 {
		m.error = "Quantity must be greater than 0"
		return false
	}

	// Round to 2 decimal places
	rounded := fmt.Sprintf("%.2f", quantity)
	m.quantityInput.SetValue(rounded)

	return true
}

// Submitted returns true if the form was submitted
func (m Model) Submitted() bool {
	return m.submitted
}

// Cancelled returns true if the form was cancelled
func (m Model) Cancelled() bool {
	return m.cancelled
}

// Data holds typed data for item forms
type Data struct {
	Name     string
	Quantity float64
	UnitType domain.UnitType
}

// GetData returns the form data
func (m Model) GetData() Data {
	quantity, _ := strconv.ParseFloat(m.quantityInput.Value(), 64)
	return Data{
		Name:     m.nameInput.Value(),
		Quantity: quantity,
		UnitType: m.unitTypes[m.unitTypeIndex],
	}
}

// Error returns any validation error
func (m Model) Error() string {
	return m.error
}

// NameInput returns the name input for rendering
func (m Model) NameInput() textinput.Model {
	return m.nameInput
}

// QuantityInput returns the quantity input for rendering
func (m Model) QuantityInput() textinput.Model {
	return m.quantityInput
}

// FocusIndex returns the current focus index
func (m Model) FocusIndex() int {
	return m.focusIndex
}

// UnitTypes returns the available unit types
func (m Model) UnitTypes() []domain.UnitType {
	return m.unitTypes
}

// UnitTypeIndex returns the selected unit type index
func (m Model) UnitTypeIndex() int {
	return m.unitTypeIndex
}
