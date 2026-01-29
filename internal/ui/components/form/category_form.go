package form

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"

	"home-inventory-system/internal/ui/components/shared"
)

// NewCategoryForm creates a new form for creating categories
func NewCategoryForm(width, height int) Model {
	// Create text inputs
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
		FormType:   FormTypeCategory,
		Title:      "Create New Category",
		Inputs:     []textinput.Model{nameInput, descInput},
		FocusIndex: 0,
		Width:      width,
		Height:     height,
	}
}

// renderCategoryForm renders the category creation form
func (m Model) renderCategoryForm() string {
	if m.FormType != FormTypeCategory {
		return ""
	}

	theme := shared.DefaultFormTheme()
	var b strings.Builder

	// Title
	b.WriteString(theme.TitleStyle.Render(m.Title))
	b.WriteString("\n\n")

	// Name field
	if m.FocusIndex == 0 {
		b.WriteString(theme.FocusedLabelStyle.Render(shared.FocusIndicator + "Name:"))
	} else {
		b.WriteString(theme.LabelStyle.Render(shared.UnfocusedPrefix + "Name:"))
	}
	b.WriteString("\n  ")
	b.WriteString(m.Inputs[0].View())
	b.WriteString("\n\n")

	// Description field
	if m.FocusIndex == 1 {
		b.WriteString(theme.FocusedLabelStyle.Render(shared.FocusIndicator + "Description:"))
	} else {
		b.WriteString(theme.LabelStyle.Render(shared.UnfocusedPrefix + "Description:"))
	}
	b.WriteString("\n  ")
	b.WriteString(m.Inputs[1].View())
	b.WriteString("\n\n")

	// Error message
	if m.Error != "" {
		b.WriteString(theme.ErrorStyle.Render("✗ " + m.Error))
		b.WriteString("\n\n")
	}

	// Help text
	b.WriteString(theme.HelpStyle.Render("Tab: next field • Enter: submit • Esc: cancel"))

	return b.String()
}
