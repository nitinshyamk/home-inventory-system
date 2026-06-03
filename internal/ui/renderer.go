package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"home-inventory-system/internal/domain"
	"home-inventory-system/internal/ui/components/modal"
	"home-inventory-system/internal/ui/styles"
)

// splitPaneMinWidth is the minimum terminal width required to show the split-pane layout.
const splitPaneMinWidth = 70

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
	case StateCreatingCategory, StateCreatingItem:
		return renderFormOverlay(m)
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

// renderMainView selects between split-pane and narrow layouts based on terminal width.
func renderMainView(m Model) string {
	if m.width >= splitPaneMinWidth {
		return renderSplitView(m)
	}
	return renderNarrowView(m)
}

// renderNarrowView renders the single-pane layout for narrow terminals.
func renderNarrowView(m Model) string {
	var b strings.Builder
	b.WriteString(renderBreadcrumb(m.breadcrumb))
	b.WriteString("\n\n")
	b.WriteString(m.itemList.View())
	b.WriteString("\n")
	b.WriteString(renderHelp(m.state))
	return b.String()
}

// renderSplitView renders the split-pane layout: left navigation + right detail.
func renderSplitView(m Model) string {
	leftWidth, rightWidth, paneHeight := splitPaneDimensions(m.width, m.height)

	leftPane := lipgloss.NewStyle().Width(leftWidth).Height(paneHeight).Render(m.itemList.View())

	sepLines := make([]string, paneHeight)
	for i := range sepLines {
		sepLines[i] = styles.DimStyle.Render("│")
	}
	sep := strings.Join(sepLines, "\n")

	rightPane := lipgloss.NewStyle().Width(rightWidth).Height(paneHeight).Render(renderRightPane(m))

	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, sep, rightPane)

	var b strings.Builder
	b.WriteString(renderBreadcrumb(m.breadcrumb))
	b.WriteString("\n")
	b.WriteString(body)
	b.WriteString(renderHelp(m.state))
	return b.String()
}

// renderRightPane renders the right pane content based on current model state.
// Task 1: placeholder only. Subsequent tasks populate this with live detail.
func renderRightPane(m Model) string {
	return styles.DimStyle.Render("Select an item to see details")
}

// splitPaneDimensions computes left/right pane widths and pane height.
func splitPaneDimensions(width, height int) (leftWidth, rightWidth, paneHeight int) {
	leftWidth = width * 2 / 5
	rightWidth = width - leftWidth - 1 // 1 char for separator
	paneHeight = height - 4            // breadcrumb(1) + blank(1) + help padding(1) + help text(1)
	if paneHeight < 1 {
		paneHeight = 1
	}
	return
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

// renderFormOverlay renders the form as a centered modal overlay.
func renderFormOverlay(m Model) string {
	// Render the appropriate form based on state
	var formContent string
	switch m.state {
	case StateCreatingCategory:
		formContent = m.categoryForm.View()
	case StateCreatingItem:
		formContent = m.itemForm.View()
	}

	// Use modal component for rendering
	opts := modal.DefaultOptions()
	return modal.Render(m.width, m.height, formContent, opts)
}
