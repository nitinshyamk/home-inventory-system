package categoryedit

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"home-inventory-system/internal/ui/styles"
)

var (
	focusedField = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	normalField  = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dimButton    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	activeButton = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
)

// View renders the category edit form for the right pane.
func (m Model) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Edit Category"))
	b.WriteString("\n\n")

	// Name field
	nameLabel := renderLabel("Name", m.focusIndex == FocusName)
	b.WriteString(nameLabel + "  " + m.nameInput.View())
	b.WriteString("\n\n")

	// Description field
	descLabel := renderLabel("Description", m.focusIndex == FocusDesc)
	b.WriteString(descLabel + "  " + m.descInput.View())
	b.WriteString("\n\n")

	// Inline error
	if m.errorMsg != "" {
		b.WriteString(styles.ErrorStyle.Render(m.errorMsg))
		b.WriteString("\n\n")
	}

	// Save button
	saveStyle := dimButton
	saveLabel := "[ Save ]"
	if m.focusIndex == FocusSave {
		saveStyle = activeButton
	}
	if !m.HasChanges() {
		saveStyle = dimButton
		saveLabel = "[ Save ]"
	}
	b.WriteString(saveStyle.Render(saveLabel))
	b.WriteString("   ")

	// Cancel button
	cancelStyle := normalField
	if m.focusIndex == FocusCancel {
		cancelStyle = activeButton
	}
	b.WriteString(cancelStyle.Render("[ Cancel ]"))

	return b.String()
}

func renderLabel(text string, focused bool) string {
	label := text + ":"
	padded := label + strings.Repeat(" ", 12-len(label))
	if focused {
		return focusedField.Render(padded)
	}
	return dimButton.Render(padded)
}
