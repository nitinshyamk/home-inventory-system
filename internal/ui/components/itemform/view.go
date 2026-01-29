package itemform

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"home-inventory-system/internal/ui/components/shared"
)

// View renders the item form
func (m Model) View() string {
	theme := shared.DefaultFormTheme()
	var b strings.Builder

	// Title
	b.WriteString(theme.TitleStyle.Render("Add New Item"))
	b.WriteString("\n\n")

	// Name field
	nameField := shared.Field{
		Label:     "Name",
		Input:     m.nameInput,
		IsFocused: m.focusIndex == 0,
	}
	b.WriteString(nameField.Render(theme))
	b.WriteString("\n\n")

	// Quantity field
	quantityField := shared.Field{
		Label:     "Quantity",
		Input:     m.quantityInput,
		IsFocused: m.focusIndex == 1,
	}
	b.WriteString(quantityField.Render(theme))
	b.WriteString("\n\n")

	// Unit type dropdown (custom rendering)
	if m.focusIndex == 2 {
		b.WriteString(theme.FocusedLabelStyle.Render(shared.FocusIndicator + "Unit Type:"))
	} else {
		b.WriteString(theme.LabelStyle.Render(shared.UnfocusedPrefix + "Unit Type:"))
	}
	b.WriteString("\n  ")
	b.WriteString(m.renderUnitTypeSelector())
	b.WriteString("\n\n")

	// Error message
	if m.error != "" {
		b.WriteString(theme.ErrorStyle.Render("✗ " + m.error))
		b.WriteString("\n\n")
	}

	// Help text
	if m.focusIndex == 2 {
		b.WriteString(theme.HelpStyle.Render("↑/↓: select unit • Enter: submit • Esc: cancel"))
	} else {
		b.WriteString(theme.HelpStyle.Render("Tab: next field • Enter: next/submit • Esc: cancel"))
	}

	return b.String()
}

// renderUnitTypeSelector renders the unit type dropdown
func (m Model) renderUnitTypeSelector() string {
	var b strings.Builder

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)
	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	b.WriteString("[")
	for i, unit := range m.unitTypes {
		if i > 0 {
			b.WriteString(" | ")
		}
		if i == m.unitTypeIndex {
			b.WriteString(selectedStyle.Render(unit.String()))
		} else {
			b.WriteString(normalStyle.Render(unit.String()))
		}
	}
	b.WriteString("]")

	return b.String()
}
