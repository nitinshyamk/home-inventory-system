package categoryedit

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// focusIndex positions: 0=Name, 1=Description, 2=Save, 3=Cancel
const (
	focusName = iota
	focusDesc
	focusSave
	focusCancel
	focusCount
)

// Model represents a category edit form rendered in the right pane.
type Model struct {
	nameInput   textinput.Model
	descInput   textinput.Model
	focusIndex  int
	originalName string
	originalDesc string
	errorMsg    string
	submitted   bool
	cancelled   bool
}

// New creates an edit form pre-filled with current category data.
func New(name, description string) Model {
	nameInput := textinput.New()
	nameInput.SetValue(name)
	nameInput.CharLimit = 100
	nameInput.Width = 40
	nameInput.Focus()

	descInput := textinput.New()
	descInput.SetValue(description)
	descInput.CharLimit = 255
	descInput.Width = 40

	return Model{
		nameInput:    nameInput,
		descInput:    descInput,
		focusIndex:   focusName,
		originalName: name,
		originalDesc: description,
	}
}

// Update handles key events for the edit form.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "esc":
		m.cancelled = true
		return m, nil

	case "tab", "shift+tab":
		if keyMsg.String() == "shift+tab" {
			m.focusIndex = (m.focusIndex - 1 + focusCount) % focusCount
		} else {
			m.focusIndex = (m.focusIndex + 1) % focusCount
		}
		return m.applyFocus(), nil

	case "enter":
		switch m.focusIndex {
		case focusSave:
			if m.HasChanges() {
				return m.submit(), nil
			}
		case focusCancel:
			m.cancelled = true
		case focusName:
			m.focusIndex = focusDesc
			return m.applyFocus(), nil
		case focusDesc:
			m.focusIndex = focusSave
			return m.applyFocus(), nil
		}
		return m, nil
	}

	// Delegate to focused text input
	var cmd tea.Cmd
	switch m.focusIndex {
	case focusName:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case focusDesc:
		m.descInput, cmd = m.descInput.Update(msg)
	}
	return m, cmd
}

func (m Model) applyFocus() Model {
	if m.focusIndex == focusName {
		m.nameInput.Focus()
		m.descInput.Blur()
	} else {
		m.nameInput.Blur()
		m.descInput.Blur()
	}
	return m
}

func (m Model) submit() Model {
	if m.nameInput.Value() == "" {
		m.errorMsg = "Name is required"
		return m
	}
	m.submitted = true
	return m
}

// HasChanges reports whether the form values differ from the originals.
func (m Model) HasChanges() bool {
	return m.nameInput.Value() != m.originalName || m.descInput.Value() != m.originalDesc
}

// Submitted reports whether the form was submitted.
func (m Model) Submitted() bool { return m.submitted }

// Cancelled reports whether the form was cancelled.
func (m Model) Cancelled() bool { return m.cancelled }

// Name returns the current name field value.
func (m Model) Name() string { return m.nameInput.Value() }

// Description returns the current description field value.
func (m Model) Description() string { return m.descInput.Value() }

// Error returns any inline validation error message.
func (m Model) Error() string { return m.errorMsg }

// FocusIndex returns which element currently has focus (for rendering).
func (m Model) FocusIndex() int { return m.focusIndex }

// SetError sets an inline validation error message (e.g. from the server response).
func (m *Model) SetError(msg string) {
	m.errorMsg = msg
	m.submitted = false
}

// FocusName, FocusDesc, FocusSave, FocusCancel are exported focus constants.
const (
	FocusName   = focusName
	FocusDesc   = focusDesc
	FocusSave   = focusSave
	FocusCancel = focusCancel
)
