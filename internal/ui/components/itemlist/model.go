package itemlist

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"home-inventory-system/internal/domain"
	"home-inventory-system/internal/ui/styles"
)

// ListMode determines what the list is displaying
type ListMode int

const (
	ModeTypes ListMode = iota
	ModeItems
)

// typeItem implements list.Item for item types
type typeItem struct {
	data domain.ItemType
}

func (i typeItem) Title() string {
	return i.data.Name
}

func (i typeItem) Description() string {
	return i.data.Description
}

func (i typeItem) FilterValue() string { return i.data.Name }

// leafItem implements list.Item for items at leaf nodes
type leafItem struct {
	data domain.Item
}

func (i leafItem) Title() string       { return i.data.Name }
func (i leafItem) Description() string { return formatQuantity(i.data.Quantity, i.data.UnitType) }
func (i leafItem) FilterValue() string { return i.data.Name }

// addNewCategoryItem implements list.Item for the "+ Add New Category" entry
type addNewCategoryItem struct{}

func (i addNewCategoryItem) Title() string       { return "+ Add New Category" }
func (i addNewCategoryItem) Description() string { return "Create a new category" }
func (i addNewCategoryItem) FilterValue() string { return "+ Add New Category" }

// addNewItemItem implements list.Item for the "+ Add New Item" entry
type addNewItemItem struct{}

func (i addNewItemItem) Title() string       { return "+ Add New Item" }
func (i addNewItemItem) Description() string { return "Create a new item" }
func (i addNewItemItem) FilterValue() string { return "+ Add New Item" }

// formatQuantity formats a quantity and unit type for display with automatic unit conversions.
//
// Conversion rules:
// - Count: Rounded to nearest integer (e.g., "5" not "5.0")
// - Grams: Displays as oz; if >= 16 oz, displays as lbs with decimal (e.g., "1.5 lbs")
// - Liters: Displays as fl oz; if >= 128 fl oz (1 gallon), displays as gal with decimal (e.g., "2.3 gal")
func formatQuantity(quantity float64, unitType domain.UnitType) string {
	switch unitType {
	case domain.UnitTypeCount:
		// Round to nearest integer for count
		return fmt.Sprintf("%d", int(quantity+0.5))

	case domain.UnitTypeGrams:
		// Convert grams to ounces (1 oz = 28.3495 grams)
		ounces := quantity / 28.3495
		if ounces >= 16.0 {
			// Convert to pounds (16 oz = 1 lb)
			pounds := ounces / 16.0
			return fmt.Sprintf("%.1f lbs", pounds)
		}
		return fmt.Sprintf("%.1f oz", ounces)

	case domain.UnitTypeLiters:
		// Convert liters to fluid ounces (1 liter = 33.814 fl oz)
		fluidOunces := quantity * 33.814
		if fluidOunces >= 128.0 {
			// Convert to gallons (128 fl oz = 1 gallon)
			gallons := fluidOunces / 128.0
			return fmt.Sprintf("%.1f gal", gallons)
		}
		return fmt.Sprintf("%.1f fl oz", fluidOunces)

	default:
		// Fallback for unknown unit types
		return fmt.Sprintf("%.1f %s", quantity, unitType.String())
	}
}

// compactDelegate is a custom delegate for compact tabular list rendering
type compactDelegate struct {
	nameWidth int
	descWidth int
}

func newCompactDelegate() compactDelegate {
	return compactDelegate{
		nameWidth: 25,
		descWidth: 35,
	}
}

func (d compactDelegate) Height() int                               { return 1 }
func (d compactDelegate) Spacing() int                              { return 0 }
func (d compactDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d compactDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	if item == nil {
		return
	}

	title := item.FilterValue()
	var desc string
	isAddNew := false

	// Get description based on item type
	switch i := item.(type) {
	case typeItem:
		desc = i.data.Description
	case leafItem:
		desc = i.Description()
	case addNewCategoryItem:
		desc = i.Description()
		isAddNew = true
	case addNewItemItem:
		desc = i.Description()
		isAddNew = true
	}

	// Truncate/pad title to fixed width
	title = d.truncateOrPad(title, d.nameWidth)

	// Truncate/pad description to fixed width
	desc = d.truncateOrPad(desc, d.descWidth)

	// Style based on selection
	isSelected := index == m.Index()

	var indicator string
	var titleStyle, descStyle lipgloss.Style

	if isSelected {
		indicator = styles.SelectedStyle.Render("│ ")
		if isAddNew {
			// Use cyan for "+ Add" items when selected
			titleStyle = styles.SelectedStyle.Foreground(lipgloss.Color("86"))
			descStyle = styles.DimStyle.Foreground(lipgloss.Color("86"))
		} else {
			titleStyle = styles.SelectedStyle
			descStyle = styles.DimStyle.Foreground(lipgloss.Color("205"))
		}
	} else {
		indicator = styles.DimStyle.Render("  ")
		if isAddNew {
			// Use dim cyan for "+ Add" items when not selected
			titleStyle = styles.NormalStyle.Foreground(lipgloss.Color("75"))
			descStyle = styles.DimStyle.Foreground(lipgloss.Color("75"))
		} else {
			titleStyle = styles.NormalStyle
			descStyle = styles.DimStyle
		}
	}

	line := indicator + titleStyle.Render(title) + "  " + descStyle.Render(desc)
	fmt.Fprint(w, line)
}

func (d compactDelegate) truncateOrPad(s string, width int) string {
	if len(s) > width {
		if width > 3 {
			return s[:width-3] + "..."
		}
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

// Model represents the item list component state
type Model struct {
	list      list.Model
	mode      ListMode
	types     []domain.ItemType
	leafItems []domain.Item
	err       error
}

// New creates a new item list model
func New() Model {
	delegate := newCompactDelegate()

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Categories"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.Styles.Title = styles.TitleStyle

	// Use emacs-style keybindings
	l.KeyMap.CursorUp.SetKeys("up", "ctrl+p")
	l.KeyMap.CursorDown.SetKeys("down", "ctrl+n")
	l.KeyMap.GoToStart.SetKeys("home", "ctrl+a")
	l.KeyMap.GoToEnd.SetKeys("end", "ctrl+e")

	return Model{
		list: l,
		mode: ModeTypes,
	}
}

// SetSize updates the list dimensions
func (m *Model) SetSize(width, height int) {
	m.list.SetSize(width, height)
}

// SetTypes updates the list with item types for hierarchy browsing
func (m *Model) SetTypes(types []domain.ItemType) {
	m.types = types
	m.mode = ModeTypes
	m.list.Title = "Categories"

	// Add regular type items
	var listItems []list.Item
	for _, t := range types {
		listItems = append(listItems, typeItem{data: t})
	}

	// Always append "+ Add New Category"
	listItems = append(listItems, addNewCategoryItem{})

	// If the category list is empty, also add "+ Add New Item" option
	// This allows users to choose whether to make this an empty category a branch or leaf
	if len(types) == 0 {
		listItems = append(listItems, addNewItemItem{})
	}

	m.list.SetItems(listItems)
	m.list.ResetSelected()
}

// SetLeafItems updates the list with items at a leaf node
func (m *Model) SetLeafItems(items []domain.Item) {
	m.leafItems = items
	m.mode = ModeItems
	m.list.Title = "Items"

	// Add regular item entries
	listItems := make([]list.Item, len(items)+1)
	for i, item := range items {
		listItems[i] = leafItem{data: item}
	}
	// Append "+ Add New Item" at the end
	listItems[len(items)] = addNewItemItem{}

	m.list.SetItems(listItems)
	m.list.ResetSelected()
}

// SetError sets an error state
func (m *Model) SetError(err error) {
	m.err = err
}

// SelectedType returns the currently selected type, if any
func (m *Model) SelectedType() *domain.ItemType {
	if m.mode != ModeTypes {
		return nil
	}
	if selectedItem, ok := m.list.SelectedItem().(typeItem); ok {
		return &selectedItem.data
	}
	return nil
}

// SelectedLeafItem returns the currently selected leaf item, if any
func (m *Model) SelectedLeafItem() *domain.Item {
	if m.mode != ModeItems {
		return nil
	}
	if selectedItem, ok := m.list.SelectedItem().(leafItem); ok {
		return &selectedItem.data
	}
	return nil
}

// Mode returns the current display mode
func (m *Model) Mode() ListMode {
	return m.mode
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the item list
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the item list
func (m Model) View() string {
	return m.list.View()
}

// SelectedAddNewCategory returns true if the "+ Add New Category" item is selected
func (m *Model) SelectedAddNewCategory() bool {
	if m.mode != ModeTypes {
		return false
	}
	_, ok := m.list.SelectedItem().(addNewCategoryItem)
	return ok
}

// SelectedAddNewItem returns true if the "+ Add New Item" item is selected
func (m *Model) SelectedAddNewItem() bool {
	// Can be selected in ModeItems, or in ModeTypes when the category list is empty
	_, ok := m.list.SelectedItem().(addNewItemItem)
	return ok
}
