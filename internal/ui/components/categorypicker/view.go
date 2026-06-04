package categorypicker

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"home-inventory-system/internal/ui/styles"
)

var (
	selectedEntry = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	normalEntry   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
)

// View renders the inline category picker.
func (m Model) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Change Category"))
	b.WriteString("\n\n")
	b.WriteString(styles.DimStyle.Render("Filter: "))
	b.WriteString(m.filterInput.View())
	b.WriteString("\n\n")

	if len(m.filtered) == 0 {
		b.WriteString(styles.DimStyle.Render("No categories match"))
		return b.String()
	}

	// Show up to 10 entries
	limit := 10
	if len(m.filtered) < limit {
		limit = len(m.filtered)
	}

	for i := 0; i < limit; i++ {
		entry := m.filtered[i]
		var indicator string
		var style lipgloss.Style
		if i == m.cursor {
			indicator = "│ "
			style = selectedEntry
		} else {
			indicator = "  "
			style = normalEntry
		}
		b.WriteString(fmt.Sprintf("%s%s\n", indicator, style.Render(entry.FullPath)))
	}

	if len(m.filtered) > limit {
		b.WriteString(styles.DimStyle.Render(fmt.Sprintf("  ... %d more", len(m.filtered)-limit)))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("↑/↓: navigate  Enter: select  Esc: cancel"))

	return b.String()
}
