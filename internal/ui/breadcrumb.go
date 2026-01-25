package ui

import (
	"strings"

	"home-inventory-system/internal/domain"
	"home-inventory-system/internal/ui/styles"
)

// Breadcrumb operations and rendering for hierarchical navigation.

// getParent returns the parent type from breadcrumb, or nil if at root.
// Parent is the second-to-last element in the breadcrumb trail.
func getParent(breadcrumb []domain.ItemType) *domain.ItemType {
	if len(breadcrumb) < 2 {
		return nil
	}
	parent := breadcrumb[len(breadcrumb)-2]
	return &parent
}

// getGrandparent returns the grandparent type from breadcrumb, or nil if depth < 3.
func getGrandparent(breadcrumb []domain.ItemType) *domain.ItemType {
	if len(breadcrumb) < 3 {
		return nil
	}
	grandparent := breadcrumb[len(breadcrumb)-3]
	return &grandparent
}

// getCurrentType returns the current (last) type in breadcrumb, or nil if empty.
func getCurrentType(breadcrumb []domain.ItemType) *domain.ItemType {
	if len(breadcrumb) == 0 {
		return nil
	}
	current := breadcrumb[len(breadcrumb)-1]
	return &current
}

// isAtRoot returns true if breadcrumb represents root level (empty or single item).
func isAtRoot(breadcrumb []domain.ItemType) bool {
	return len(breadcrumb) <= 1
}

// getDepth returns the current depth in the hierarchy (0 = root).
func getDepth(breadcrumb []domain.ItemType) int {
	return len(breadcrumb)
}

// clearBreadcrumb clears the breadcrumb trail (used when navigating to root).
func clearBreadcrumb(m Model) Model {
	m.breadcrumb = nil
	return m
}

// setBreadcrumb sets the breadcrumb to the given path.
func setBreadcrumb(m Model, path []domain.ItemType) Model {
	m.breadcrumb = path
	return m
}

// renderBreadcrumb formats the breadcrumb trail for display.
// Renders as "Home > Parent > Current" with the last item highlighted.
func renderBreadcrumb(breadcrumb []domain.ItemType) string {
	if len(breadcrumb) == 0 {
		return styles.TitleStyle.Render("Home")
	}

	parts := make([]string, len(breadcrumb))
	for i, t := range breadcrumb {
		if i == len(breadcrumb)-1 {
			// Highlight current location
			parts[i] = styles.SelectedStyle.Render(t.Name)
		} else {
			// Dim parent paths
			parts[i] = styles.DimStyle.Render(t.Name)
		}
	}

	separator := styles.DimStyle.Render(" > ")
	return styles.TitleStyle.Render("Home") + separator + strings.Join(parts, separator)
}
