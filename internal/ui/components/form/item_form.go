package form

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"home-inventory-system/internal/domain"
	"home-inventory-system/internal/ui/components/shared"
)

// NewItemForm creates a new form for creating items
func NewItemForm(width, height int) Model {
	// Create text inputs
	nameInput := textinput.New()
	nameInput.Placeholder = "e.g., Rice Bag"
	nameInput.Focus()
	nameInput.CharLimit = 100
	nameInput.Width = 50

	quantityInput := textinput.New()
	quantityInput.Placeholder = "1.0"
	quantityInput.CharLimit = 10
	quantityInput.Width = 20

	// Initialize with smart defaults
	quantityInput.SetValue("1.0")

	return Model{
		FormType:      FormTypeItem,
		Title:         "Add New Item",
		Inputs:        []textinput.Model{nameInput, quantityInput},
		FocusIndex:    0,
		UnitTypes:     []domain.UnitType{domain.UnitTypeCount, domain.UnitTypeGrams, domain.UnitTypeLiters},
		UnitTypeIndex: 0, // Default to Count
		Width:         width,
		Height:        height,
	}
}

// renderItemForm renders the item creation form
func (m Model) renderItemForm() string {
	if m.FormType != FormTypeItem {
		return ""
	}

	theme := shared.DefaultFormTheme()
	var b strings.Builder

	// Title
	b.WriteString(theme.TitleStyle.Render(m.Title))
	b.WriteString("\n\n")

	// Name field
	if m.FocusIndex == 0 {
		b.WriteString(theme.FocusedLabelStyle.Render(shared.FocusIndicator + "Name:"))
	} else {
		b.WriteString(theme.LabelStyle.Render(shared.UnfocusedPrefix + "Name:"))
	}
	b.WriteString("\n  ")
	b.WriteString(m.Inputs[0].View())
	b.WriteString("\n\n")

	// Quantity field
	if m.FocusIndex == 1 {
		b.WriteString(theme.FocusedLabelStyle.Render(shared.FocusIndicator + "Quantity:"))
	} else {
		b.WriteString(theme.LabelStyle.Render(shared.UnfocusedPrefix + "Quantity:"))
	}
	b.WriteString("\n  ")
	b.WriteString(m.Inputs[1].View())
	b.WriteString("\n\n")

	// Unit type dropdown
	if m.FocusIndex == 2 {
		b.WriteString(theme.FocusedLabelStyle.Render(shared.FocusIndicator + "Unit Type:"))
	} else {
		b.WriteString(theme.LabelStyle.Render(shared.UnfocusedPrefix + "Unit Type:"))
	}
	b.WriteString("\n  ")
	b.WriteString(m.renderUnitTypeSelector())
	b.WriteString("\n\n")

	// Error message
	if m.Error != "" {
		b.WriteString(theme.ErrorStyle.Render("✗ " + m.Error))
		b.WriteString("\n\n")
	}

	// Help text
	if m.FocusIndex == 2 {
		b.WriteString(theme.HelpStyle.Render("↑/↓: select unit • Enter: submit • Esc: cancel"))
	} else {
		b.WriteString(theme.HelpStyle.Render("Tab: next field • Enter: next/submit • Esc: cancel"))
	}

	return b.String()
}

// renderUnitTypeSelector renders the unit type dropdown
func (m Model) renderUnitTypeSelector() string {
	var b strings.Builder

	// Use same color scheme as theme for consistency
	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)
	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	b.WriteString("[")
	for i, unit := range m.UnitTypes {
		if i > 0 {
			b.WriteString(" | ")
		}
		if i == m.UnitTypeIndex {
			b.WriteString(selectedStyle.Render(unit.String()))
		} else {
			b.WriteString(normalStyle.Render(unit.String()))
		}
	}
	b.WriteString("]")

	return b.String()
}

// validateItemForm validates item-specific fields
func (m *Model) validateItemForm() bool {
	if m.FormType != FormTypeItem {
		return false
	}

	m.Error = ""

	// Name is required
	if len(m.Inputs) > 0 && m.Inputs[0].Value() == "" {
		m.Error = "Name is required"
		return false
	}

	// Quantity is required and must be valid number
	if len(m.Inputs) > 1 {
		quantityStr := m.Inputs[1].Value()
		if quantityStr == "" {
			m.Error = "Quantity is required"
			return false
		}

		quantity, err := strconv.ParseFloat(quantityStr, 64)
		if err != nil {
			m.Error = "Quantity must be a valid number"
			return false
		}

		if quantity <= 0 {
			m.Error = "Quantity must be greater than 0"
			return false
		}

		// Round to 2 decimal places and validate
		rounded := fmt.Sprintf("%.2f", quantity)
		m.Inputs[1].SetValue(rounded)
	}

	return true
}
