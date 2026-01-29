package categoryform

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents a category creation form
type Model struct {
	nameInput  textinput.Model
	descInput  textinput.Model
	focusIndex int
	error      string
	submitted  bool
	cancelled  bool
	width      int
	height     int
}

// New creates a new category form
func New(width, height int) Model {
	nameInput := textinput.New()
	nameInput.Placeholder = "e.g., Kitchen Appliances"
	nameInput.Focus()
	nameInput.CharLimit = 100
	nameInput.Width = 50

	descInput := textinput.New()
	descInput.Placeholder = "Optional description"
	descInput.CharLimit = 255
	descInput.Width = 50

	return Model{
		nameInput:  nameInput,
		descInput:  descInput,
		focusIndex: 0,
		width:      width,
		height:     height,
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
			if m.focusIndex == 1 {
				// On last field, submit
				if m.validate() {
					m.submitted = true
					return m, nil
				}
			} else {
				// Move to next field
				m.focusIndex++
				m.updateFocus()
			}

		case "tab":
			m.focusIndex++
			if m.focusIndex > 1 {
				m.focusIndex = 0
			}
			m.updateFocus()

		case "shift+tab":
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = 1
			}
			m.updateFocus()

		case "down":
			m.focusIndex++
			if m.focusIndex > 1 {
				m.focusIndex = 0
			}
			m.updateFocus()

		case "up":
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = 1
			}
			m.updateFocus()
		}
	}

	// Update the focused input
	var cmd tea.Cmd
	if m.focusIndex == 0 {
		m.nameInput, cmd = m.nameInput.Update(msg)
	} else {
		m.descInput, cmd = m.descInput.Update(msg)
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// updateFocus updates which input field is focused
func (m *Model) updateFocus() {
	if m.focusIndex == 0 {
		m.nameInput.Focus()
		m.descInput.Blur()
	} else {
		m.nameInput.Blur()
		m.descInput.Focus()
	}
}

// validate checks if the form data is valid
func (m *Model) validate() bool {
	m.error = ""
	if m.nameInput.Value() == "" {
		m.error = "Name is required"
		return false
	}
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

// Data returns the form data
type Data struct {
	Name        string
	Description string
}

// GetData returns the form data
func (m Model) GetData() Data {
	return Data{
		Name:        m.nameInput.Value(),
		Description: m.descInput.Value(),
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

// DescInput returns the description input for rendering
func (m Model) DescInput() textinput.Model {
	return m.descInput
}

// FocusIndex returns the current focus index
func (m Model) FocusIndex() int {
	return m.focusIndex
}
