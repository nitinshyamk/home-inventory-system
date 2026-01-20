package itemlist

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"home-inventory-system/internal/repository"
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
	data repository.ItemType
}

func (i typeItem) Title() string {
	return i.data.Name
}

func (i typeItem) Description() string {
	if i.data.Description.Valid {
		return i.data.Description.String
	}
	return ""
}

func (i typeItem) FilterValue() string { return i.data.Name }

// leafItem implements list.Item for items at leaf nodes
type leafItem struct {
	data repository.Item
}

func (i leafItem) Title() string       { return i.data.Name }
func (i leafItem) Description() string { return fmt.Sprintf("ID: %d", i.data.ID) }
func (i leafItem) FilterValue() string { return i.data.Name }

// Model represents the item list component state
type Model struct {
	list          list.Model
	mode          ListMode
	types         []repository.ItemType
	leafItems     []repository.Item
	selectedType  *repository.ItemType
	selectedItem  *repository.Item
	err           error
}

// New creates a new item list model
func New() Model {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = styles.SelectedStyle
	delegate.Styles.SelectedDesc = styles.DimStyle

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Categories"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
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
func (m *Model) SetTypes(types []repository.ItemType) {
	m.types = types
	m.mode = ModeTypes
	m.list.Title = "Categories"

	listItems := make([]list.Item, len(types))
	for i, t := range types {
		listItems[i] = typeItem{data: t}
	}
	m.list.SetItems(listItems)
	m.list.ResetSelected()
}

// SetLeafItems updates the list with items at a leaf node
func (m *Model) SetLeafItems(items []repository.Item) {
	m.leafItems = items
	m.mode = ModeItems
	m.list.Title = "Items"

	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = leafItem{data: item}
	}
	m.list.SetItems(listItems)
	m.list.ResetSelected()
}

// SetError sets an error state
func (m *Model) SetError(err error) {
	m.err = err
}

// SelectedType returns the currently selected type, if any
func (m *Model) SelectedType() *repository.ItemType {
	if m.mode != ModeTypes {
		return nil
	}
	if selectedItem, ok := m.list.SelectedItem().(typeItem); ok {
		return &selectedItem.data
	}
	return nil
}

// SelectedLeafItem returns the currently selected leaf item, if any
func (m *Model) SelectedLeafItem() *repository.Item {
	if m.mode != ModeItems {
		return nil
	}
	if selectedItem, ok := m.list.SelectedItem().(leafItem); ok {
		return &selectedItem.data
	}
	return nil
}

// ClearSelection clears the current selection
func (m *Model) ClearSelection() {
	m.selectedType = nil
	m.selectedItem = nil
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
