package form

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
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

	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))
	b.WriteString(titleStyle.Render(m.Title))
	b.WriteString("\n\n")

	// Name field
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	if m.FocusIndex == 0 {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Render("• Name:"))
	} else {
		b.WriteString(labelStyle.Render("  Name:"))
	}
	b.WriteString("\n  ")
	b.WriteString(m.Inputs[0].View())
	b.WriteString("\n\n")

	// Description field
	if m.FocusIndex == 1 {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Render("• Description:"))
	} else {
		b.WriteString(labelStyle.Render("  Description:"))
	}
	b.WriteString("\n  ")
	b.WriteString(m.Inputs[1].View())
	b.WriteString("\n\n")

	// Error message
	if m.Error != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		b.WriteString(errorStyle.Render("✗ " + m.Error))
		b.WriteString("\n\n")
	}

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
	b.WriteString(helpStyle.Render("Tab: next field • Enter: submit • Esc: cancel"))

	return b.String()
}
