package form

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"home-inventory-system/internal/domain"
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

	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))
	b.WriteString(titleStyle.Render(m.Title))
	b.WriteString("\n\n")

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	// Name field
	if m.FocusIndex == 0 {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Render("• Name:"))
	} else {
		b.WriteString(labelStyle.Render("  Name:"))
	}
	b.WriteString("\n  ")
	b.WriteString(m.Inputs[0].View())
	b.WriteString("\n\n")

	// Quantity field
	if m.FocusIndex == 1 {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Render("• Quantity:"))
	} else {
		b.WriteString(labelStyle.Render("  Quantity:"))
	}
	b.WriteString("\n  ")
	b.WriteString(m.Inputs[1].View())
	b.WriteString("\n\n")

	// Unit type dropdown
	if m.FocusIndex == 2 {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Render("• Unit Type:"))
	} else {
		b.WriteString(labelStyle.Render("  Unit Type:"))
	}
	b.WriteString("\n  ")
	b.WriteString(m.renderUnitTypeSelector())
	b.WriteString("\n\n")

	// Error message
	if m.Error != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		b.WriteString(errorStyle.Render("✗ " + m.Error))
		b.WriteString("\n\n")
	}

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
	if m.FocusIndex == 2 {
		b.WriteString(helpStyle.Render("↑/↓: select unit • Enter: submit • Esc: cancel"))
	} else {
		b.WriteString(helpStyle.Render("Tab: next field • Enter: next/submit • Esc: cancel"))
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
