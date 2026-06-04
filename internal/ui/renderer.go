package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"home-inventory-system/internal/domain"
	"home-inventory-system/internal/ui/components/itemlist"
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
	case StateBrowsingTypes, StateViewingItems, StateEditingCategory, StateEditingItem,
		StatePickingCategory, StateDeletingItem, StateDeletingCategory:
		return renderMainView(m)
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

	rightPane := lipgloss.NewStyle().Width(rightWidth).Height(paneHeight).Render(renderRightPaneForModel(m))

	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, sep, rightPane)

	var b strings.Builder
	b.WriteString(renderBreadcrumb(m.breadcrumb))
	b.WriteString("\n")
	b.WriteString(body)
	b.WriteString(renderHelp(m.state))
	return b.String()
}

// renderRightPaneForModel renders the right pane accounting for edit states.
func renderRightPaneForModel(m Model) string {
	switch m.state {
	case StateEditingCategory:
		return m.categoryEditForm.View()
	case StateEditingItem:
		return m.itemEditForm.View()
	case StatePickingCategory:
		return m.categoryPicker.View()
	case StateDeletingItem:
		return renderItemDeleteConfirm(m)
	case StateDeletingCategory:
		return renderCategoryDeleteConfirm(m)
	}
	return renderRightPane(m.rightPane)
}

// renderRightPane renders the right pane content from cached detail data.
func renderRightPane(p rightPaneContent) string {
	if p.item != nil {
		return renderItemDetailPane(p)
	}
	if p.category != nil {
		return renderCategoryDetailPane(p)
	}
	return styles.DimStyle.Render("Select an item to see details")
}

// renderItemDeleteConfirm renders the item delete confirmation in the right pane.
func renderItemDeleteConfirm(m Model) string {
	if m.rightPane.item == nil {
		return renderRightPane(m.rightPane)
	}
	item := m.rightPane.item
	var b strings.Builder
	b.WriteString(styles.TitleStyle.Render("Delete Item"))
	b.WriteString("\n\n")
	b.WriteString(styles.NormalStyle.Render(fmt.Sprintf("Delete %q?", item.Name)))
	b.WriteString("\n\n")
	b.WriteString(styles.DimStyle.Render("This action cannot be undone."))
	b.WriteString("\n\n")
	b.WriteString(renderConfirmButtons(m.deleteConfirmFocus))
	return b.String()
}

// renderConfirmButtons renders [ Confirm Delete ] / [ Cancel ] with focus indicator.
func renderConfirmButtons(focusIndex int) string {
	confirmStyle := styles.NormalStyle
	cancelStyle := styles.NormalStyle
	focused := styles.SelectedStyle
	if focusIndex == 0 {
		confirmStyle = focused
	} else {
		cancelStyle = focused
	}
	return confirmStyle.Render("[ Confirm Delete ]") + "   " + cancelStyle.Render("[ Cancel ]")
}

// renderCategoryDeleteConfirm renders the category delete confirmation in the right pane.
func renderCategoryDeleteConfirm(m Model) string {
	if m.rightPane.category == nil {
		return renderRightPane(m.rightPane)
	}
	cat := m.rightPane.category
	sum := m.deleteSummary

	var b strings.Builder
	b.WriteString(styles.TitleStyle.Render("Delete Category"))
	b.WriteString("\n\n")

	if sum == nil {
		// Still loading the summary
		b.WriteString(styles.DimStyle.Render("Loading..."))
		return b.String()
	}

	// Root-level with items → rejection
	if sum.parentID == nil && sum.itemCount > 0 {
		b.WriteString(styles.ErrorStyle.Render(fmt.Sprintf("Cannot delete %q", cat.Name)))
		b.WriteString("\n\n")
		b.WriteString(styles.NormalStyle.Render("This category contains items that"))
		b.WriteString("\n")
		b.WriteString(styles.NormalStyle.Render("cannot be lifted to the root level."))
		b.WriteString("\n\n")
		b.WriteString(styles.DimStyle.Render("Move or delete the items first."))
		b.WriteString("\n\n")
		b.WriteString(styles.NormalStyle.Render("[ Cancel ]"))
		return b.String()
	}

	b.WriteString(styles.NormalStyle.Render(fmt.Sprintf("Delete %q?", cat.Name)))
	b.WriteString("\n")

	if sum.childTypeCount > 0 {
		parentName := "Home"
		if len(m.breadcrumb) >= 2 {
			parentName = m.breadcrumb[len(m.breadcrumb)-2].Name
		}
		b.WriteString(styles.DimStyle.Render(fmt.Sprintf(
			"%d subcategor%s will be moved up to %q.",
			sum.childTypeCount, pluralIes(sum.childTypeCount), parentName,
		)))
	} else if sum.itemCount > 0 {
		parentName := "Home"
		if len(m.breadcrumb) >= 2 {
			parentName = m.breadcrumb[len(m.breadcrumb)-2].Name
		}
		b.WriteString(styles.DimStyle.Render(fmt.Sprintf(
			"%d item%s will be moved up to %q.",
			sum.itemCount, pluralS(sum.itemCount), parentName,
		)))
	} else {
		b.WriteString(styles.DimStyle.Render("This action cannot be undone."))
	}

	b.WriteString("\n\n")
	b.WriteString(renderConfirmButtons(m.deleteConfirmFocus))
	return b.String()
}

func pluralIes(n int) string {
	if n == 1 {
		return "y"
	}
	return "ies"
}

func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// renderCategoryDetailPane renders a category's name, description, and child/item count.
func renderCategoryDetailPane(p rightPaneContent) string {
	var b strings.Builder
	b.WriteString(styles.TitleStyle.Render(p.category.Name))
	b.WriteString("\n")
	if p.category.Description != "" {
		b.WriteString(styles.DimStyle.Render("Description: " + p.category.Description))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	if p.isLeaf {
		b.WriteString(styles.DimStyle.Render(fmt.Sprintf("Items:         %d", p.itemCount)))
	} else {
		b.WriteString(styles.DimStyle.Render(fmt.Sprintf("Subcategories: %d", p.childCount)))
	}
	return b.String()
}

// renderItemDetailPane renders an item's name, quantity, category path, and created date.
func renderItemDetailPane(p rightPaneContent) string {
	item := p.item
	var b strings.Builder
	b.WriteString(styles.TitleStyle.Render(item.Name))
	b.WriteString("\n")
	b.WriteString(styles.DimStyle.Render("Category:  " + getCategoryPath(p.itemTypePath)))
	b.WriteString("\n")
	b.WriteString(styles.NormalStyle.Render(
		fmt.Sprintf("Quantity:  %s", itemlist.FormatQuantity(item.Quantity, item.UnitType)),
	))
	b.WriteString("\n")
	b.WriteString(styles.DimStyle.Render("Created:   " + item.CreatedAt))
	return b.String()
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
