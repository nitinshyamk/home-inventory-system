package categoryform

import (
	"strings"

	"home-inventory-system/internal/ui/components/shared"
)

// View renders the category form
func (m Model) View() string {
	theme := shared.DefaultFormTheme()
	var b strings.Builder

	// Title
	b.WriteString(theme.TitleStyle.Render("Create New Category"))
	b.WriteString("\n\n")

	// Name field
	nameField := shared.Field{
		Label:     "Name",
		Input:     m.nameInput,
		IsFocused: m.focusIndex == 0,
	}
	b.WriteString(nameField.Render(theme))
	b.WriteString("\n\n")

	// Description field
	descField := shared.Field{
		Label:     "Description",
		Input:     m.descInput,
		IsFocused: m.focusIndex == 1,
	}
	b.WriteString(descField.Render(theme))
	b.WriteString("\n\n")

	// Error message
	if m.error != "" {
		b.WriteString(theme.ErrorStyle.Render("✗ " + m.error))
		b.WriteString("\n\n")
	}

	// Help text
	b.WriteString(theme.HelpStyle.Render("Tab: next field • Enter: submit • Esc: cancel"))

	return b.String()
}
