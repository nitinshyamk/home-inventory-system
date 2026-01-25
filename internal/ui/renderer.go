package ui

import (
	"fmt"
	"strings"

	"home-inventory-system/internal/domain"
	"home-inventory-system/internal/ui/styles"
)

// View rendering logic for all application states.

// renderView dispatches to appropriate renderer based on model state.
func renderView(m Model) string {
	switch m.state {
	case StateLoading:
		return renderLoading()
	case StateBrowsingTypes, StateViewingItems:
		return renderMainView(m)
	case StateItemSelected:
		return renderItemDetails(m)
	case StateError:
		return renderError(m)
	default:
		return ""
	}
}

// renderLoading shows loading indicator.
func renderLoading() string {
	return styles.TitleStyle.Render("Loading inventory...")
}

// renderMainView renders the common layout with breadcrumb, content, and help.
func renderMainView(m Model) string {
	var b strings.Builder
	b.WriteString(renderBreadcrumb(m.breadcrumb))
	b.WriteString("\n\n")
	b.WriteString(m.itemList.View())
	b.WriteString("\n")
	b.WriteString(renderHelp(m.state))
	return b.String()
}

// renderItemDetails shows detailed view of selected item.
func renderItemDetails(m Model) string {
	if m.selectedItem == nil {
		return renderMainView(m)
	}

	var b strings.Builder

	// Breadcrumb
	b.WriteString(renderBreadcrumb(m.breadcrumb))
	b.WriteString("\n\n")

	// Item details
	b.WriteString(styles.TitleStyle.Render("Item Details"))
	b.WriteString("\n\n")
	b.WriteString(styles.NormalStyle.Render(fmt.Sprintf("Name: %s", m.selectedItem.Name)))
	b.WriteString("\n")
	b.WriteString(styles.NormalStyle.Render(fmt.Sprintf("Quantity: %.1f %s", m.selectedItem.Quantity, m.selectedItem.UnitType)))
	b.WriteString("\n")
	b.WriteString(styles.NormalStyle.Render(fmt.Sprintf("Category: %s", getCategoryPath(m.breadcrumb))))
	b.WriteString("\n")
	b.WriteString(styles.DimStyle.Render(fmt.Sprintf("Created: %s", m.selectedItem.CreatedAt)))
	b.WriteString("\n\n")
	b.WriteString(styles.HelpStyle.Render("Press ESC/C-b/← to go back, q to quit"))

	return b.String()
}

// getCategoryPath returns the full category path as a plain string.
// Example: "Home > Kitchen > Pantry Items > Spices"
func getCategoryPath(breadcrumb []domain.ItemType) string {
	if len(breadcrumb) == 0 {
		return "Home"
	}

	parts := make([]string, len(breadcrumb)+1)
	parts[0] = "Home"
	for i, t := range breadcrumb {
		parts[i+1] = t.Name
	}

	return strings.Join(parts, " > ")
}

// renderError shows error message.
func renderError(m Model) string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		styles.ErrorStyle.Render("An error occurred"),
		styles.NormalStyle.Render(m.err.Error()),
		styles.HelpStyle.Render("Press q to quit"),
	)
}

// renderHelp shows context-appropriate help text.
func renderHelp(state AppState) string {
	if state == StateViewingItems {
		return styles.HelpStyle.Render("C-p/C-n/↑/↓: navigate | Enter/C-f/→: select | ESC/C-b/←: back | q: quit")
	}
	return styles.HelpStyle.Render("C-p/C-n/↑/↓: navigate | Enter/C-f/→: drill down | ESC/C-b/←: back | q: quit")
}
