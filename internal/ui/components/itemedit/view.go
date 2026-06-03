package itemedit

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"home-inventory-system/internal/ui/styles"
)

var (
	focusedLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	normalLabel  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	activeButton = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	dimButton    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

// View renders the item edit form for the right pane.
func (m Model) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Edit Item"))
	b.WriteString("\n\n")

	// Name field
	b.WriteString(renderLabel("Name", m.focusIndex == FocusName))
	b.WriteString("  " + m.nameInput.View())
	b.WriteString("\n\n")

	// Quantity field
	b.WriteString(renderLabel("Quantity", m.focusIndex == FocusQty))
	b.WriteString("  " + m.qtyInput.View())
	b.WriteString("\n\n")

	// Unit selector
	unitLabel := renderLabel("Unit", m.focusIndex == FocusUnit)
	unit := unitOrder[m.unitIndex]
	var unitStr string
	if m.focusIndex == FocusUnit {
		unitStr = focusedLabel.Render("[ " + unit.String() + " ▲▼ ]")
	} else {
		unitStr = normalLabel.Render("  " + unit.String())
	}
	b.WriteString(unitLabel + "  " + unitStr)
	b.WriteString("\n\n")

	// Inline error
	if m.errorMsg != "" {
		b.WriteString(styles.ErrorStyle.Render(m.errorMsg))
		b.WriteString("\n\n")
	}

	// Save button
	saveStyle := dimButton
	if m.focusIndex == FocusSave && m.HasChanges() {
		saveStyle = activeButton
	}
	b.WriteString(saveStyle.Render("[ Save ]"))
	b.WriteString("   ")

	// Cancel button
	cancelStyle := styles.NormalStyle
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
		return focusedLabel.Render(padded)
	}
	return normalLabel.Render(padded)
}
