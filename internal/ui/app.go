package ui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"home-inventory-system/internal/repository"
	"home-inventory-system/internal/service"
	"home-inventory-system/internal/ui/components/itemlist"
	"home-inventory-system/internal/ui/messages"
	"home-inventory-system/internal/ui/styles"
)

// AppState represents the current state of the application
type AppState int

const (
	StateLoading AppState = iota
	StateBrowsingTypes  // Navigating the type hierarchy
	StateViewingItems   // Viewing items at a leaf node
	StateItemSelected   // Viewing item details
	StateError
)

// Model is the main application model
type Model struct {
	state         AppState
	itemList      itemlist.Model
	handler       *service.Handler
	err           error
	width         int
	height        int
	breadcrumb    []repository.ItemType // Current path in hierarchy
	currentTypeID *int64                // nil = root level
	items         []repository.Item     // Items at current leaf node
	selectedItem  *repository.Item
}

// NewModel creates a new application model
func NewModel(handler *service.Handler) Model {
	return Model{
		state:    StateLoading,
		itemList: itemlist.New(),
		handler:  handler,
	}
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return m.loadRootTypes
}

// loadRootTypes fetches root-level types
func (m Model) loadRootTypes() tea.Msg {
	result := m.handler.HandleQuery(context.Background(), service.ListRootTypesQuery{})
	if typesResult, ok := result.(service.ListRootTypesResult); ok {
		return messages.TypesLoadedMsg{Types: typesResult.Types, ParentID: nil, Err: typesResult.Err}
	}
	return messages.ErrorMsg{Err: fmt.Errorf("unexpected query result type")}
}

// loadChildTypes fetches child types for a parent
func (m Model) loadChildTypes(parentID int64) tea.Cmd {
	return func() tea.Msg {
		result := m.handler.HandleQuery(context.Background(), service.ListChildTypesQuery{ParentID: parentID})
		if typesResult, ok := result.(service.ListChildTypesResult); ok {
			return messages.TypesLoadedMsg{Types: typesResult.Types, ParentID: &parentID, Err: typesResult.Err}
		}
		return messages.ErrorMsg{Err: fmt.Errorf("unexpected query result type")}
	}
}

// loadItemsForType fetches items at a leaf type
func (m Model) loadItemsForType(typeID int64) tea.Cmd {
	return func() tea.Msg {
		result := m.handler.HandleQuery(context.Background(), service.ListItemsByTypeQuery{TypeID: typeID})
		if itemsResult, ok := result.(service.ListItemsByTypeResult); ok {
			return messages.LeafItemsLoadedMsg{TypeID: typeID, Items: itemsResult.Items, Err: itemsResult.Err}
		}
		return messages.ErrorMsg{Err: fmt.Errorf("unexpected query result type")}
	}
}

// loadBreadcrumb fetches the path from root to current type
func (m Model) loadBreadcrumb(typeID int64) tea.Cmd {
	return func() tea.Msg {
		result := m.handler.HandleQuery(context.Background(), service.GetTypePathQuery{TypeID: typeID})
		if pathResult, ok := result.(service.GetTypePathResult); ok {
			return messages.BreadcrumbLoadedMsg{Path: pathResult.Path, Err: pathResult.Err}
		}
		return messages.ErrorMsg{Err: fmt.Errorf("unexpected query result type")}
	}
}

// Update handles all messages for the application
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.itemList.SetSize(msg.Width, msg.Height-6) // Leave room for breadcrumb and help
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case messages.TypesLoadedMsg:
		return m.handleTypesLoaded(msg)

	case messages.LeafItemsLoadedMsg:
		return m.handleLeafItemsLoaded(msg)

	case messages.BreadcrumbLoadedMsg:
		return m.handleBreadcrumbLoaded(msg)

	case messages.TypeSelectedMsg:
		return m.handleTypeSelected(msg)

	case messages.ErrorMsg:
		return m.setError(msg.Err)
	}

	// Pass messages to item list when in browsing state
	if m.state == StateBrowsingTypes || m.state == StateViewingItems {
		var cmd tea.Cmd
		m.itemList, cmd = m.itemList.Update(msg)
		return m, cmd
	}

	return m, nil
}

// setError transitions the model to error state and returns it
func (m Model) setError(err error) (tea.Model, tea.Cmd) {
	m.state = StateError
	m.err = err
	return m, nil
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "esc", "ctrl+b", "left":
		return m.navigateUp()

	case "enter", "ctrl+f", "right":
		return m.selectCurrent()
	}

	// Pass to itemlist for navigation
	if m.state == StateBrowsingTypes || m.state == StateViewingItems {
		var cmd tea.Cmd
		m.itemList, cmd = m.itemList.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleTypesLoaded(msg messages.TypesLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		return m.setError(msg.Err)
	}

	m.itemList.SetTypes(msg.Types)
	m.state = StateBrowsingTypes
	m.currentTypeID = msg.ParentID

	// Load breadcrumb if not at root
	if msg.ParentID != nil {
		return m, m.loadBreadcrumb(*msg.ParentID)
	}

	m.breadcrumb = nil
	return m, nil
}

func (m Model) handleLeafItemsLoaded(msg messages.LeafItemsLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		return m.setError(msg.Err)
	}

	m.items = msg.Items
	m.itemList.SetLeafItems(msg.Items)
	m.state = StateViewingItems
	return m, nil
}

func (m Model) handleBreadcrumbLoaded(msg messages.BreadcrumbLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		// Non-fatal, just don't show breadcrumb
		return m, nil
	}
	m.breadcrumb = msg.Path
	return m, nil
}

func (m Model) handleTypeSelected(msg messages.TypeSelectedMsg) (tea.Model, tea.Cmd) {
	// Check if this is a leaf type (no children)
	result := m.handler.HandleQuery(context.Background(), service.IsLeafTypeQuery{TypeID: msg.Type.ID})
	leafResult, ok := result.(service.IsLeafTypeResult)
	if !ok {
		return m.setError(fmt.Errorf("failed to check leaf status"))
	}
	if leafResult.Err != nil {
		return m.setError(leafResult.Err)
	}

	if leafResult.IsLeaf {
		// Load items at this leaf
		m.currentTypeID = &msg.Type.ID
		return m, tea.Batch(
			m.loadItemsForType(msg.Type.ID),
			m.loadBreadcrumb(msg.Type.ID),
		)
	}

	// Not a leaf, load child types
	return m, m.loadChildTypes(msg.Type.ID)
}

// goToRoot navigates back to the root level
func (m Model) goToRoot() (tea.Model, tea.Cmd) {
	m.currentTypeID = nil
	m.breadcrumb = nil
	return m, m.loadRootTypes
}

// goToParent navigates to a specific parent type
func (m Model) goToParent(parentID int64) (tea.Model, tea.Cmd) {
	m.currentTypeID = &parentID
	return m, m.loadChildTypes(parentID)
}

func (m Model) navigateUp() (tea.Model, tea.Cmd) {
	if m.state == StateItemSelected {
		m.state = StateViewingItems
		m.selectedItem = nil
		return m, nil
	}

	if m.state == StateViewingItems {
		// Go back to parent type or types view
		if len(m.breadcrumb) > 1 {
			parentType := m.breadcrumb[len(m.breadcrumb)-2]
			return m.goToParent(parentType.ID)
		}
		return m.goToRoot()
	}

	if m.state == StateBrowsingTypes && m.currentTypeID != nil {
		// Navigate up one level in the type hierarchy
		if len(m.breadcrumb) > 1 {
			// Go to grandparent
			grandparent := m.breadcrumb[len(m.breadcrumb)-2]
			if grandparent.ParentID.Valid {
				return m.goToParent(grandparent.ParentID.Int64)
			}
		}
		return m.goToRoot()
	}

	return m, nil
}

func (m Model) selectCurrent() (tea.Model, tea.Cmd) {
	if m.state == StateBrowsingTypes {
		selected := m.itemList.SelectedType()
		if selected != nil {
			return m.handleTypeSelected(messages.TypeSelectedMsg{Type: *selected})
		}
	}

	if m.state == StateViewingItems {
		selected := m.itemList.SelectedLeafItem()
		if selected != nil {
			m.selectedItem = selected
			m.state = StateItemSelected
		}
	}

	return m, nil
}

// View renders the application
func (m Model) View() string {
	switch m.state {
	case StateLoading:
		return m.viewLoading()
	case StateBrowsingTypes:
		return m.viewBrowsingTypes()
	case StateViewingItems:
		return m.viewItems()
	case StateItemSelected:
		return m.viewItemSelected()
	case StateError:
		return m.viewError()
	default:
		return ""
	}
}

func (m Model) viewLoading() string {
	return styles.TitleStyle.Render("Loading inventory...")
}

func (m Model) viewBrowsingTypes() string {
	var b strings.Builder

	// Breadcrumb
	b.WriteString(m.renderBreadcrumb())
	b.WriteString("\n\n")

	// List
	b.WriteString(m.itemList.View())
	b.WriteString("\n")

	// Help
	b.WriteString(m.viewHelp())

	return b.String()
}

func (m Model) viewItems() string {
	var b strings.Builder

	// Breadcrumb
	b.WriteString(m.renderBreadcrumb())
	b.WriteString("\n\n")

	// Items list
	b.WriteString(m.itemList.View())
	b.WriteString("\n")

	// Help
	b.WriteString(m.viewHelp())

	return b.String()
}

func (m Model) viewItemSelected() string {
	if m.selectedItem == nil {
		return m.viewItems()
	}

	var b strings.Builder

	// Breadcrumb
	b.WriteString(m.renderBreadcrumb())
	b.WriteString("\n\n")

	b.WriteString(styles.TitleStyle.Render("Item Details"))
	b.WriteString("\n\n")
	b.WriteString(styles.NormalStyle.Render(fmt.Sprintf("Name: %s", m.selectedItem.Name)))
	b.WriteString("\n")
	b.WriteString(styles.DimStyle.Render(fmt.Sprintf("ID: %d", m.selectedItem.ID)))
	b.WriteString("\n")
	b.WriteString(styles.DimStyle.Render(fmt.Sprintf("Created: %s", m.selectedItem.CreatedAt)))
	b.WriteString("\n\n")
	b.WriteString(styles.HelpStyle.Render("Press ESC/C-b/← to go back, q to quit"))

	return b.String()
}

func (m Model) viewError() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		styles.ErrorStyle.Render("An error occurred"),
		styles.NormalStyle.Render(m.err.Error()),
		styles.HelpStyle.Render("Press q to quit"),
	)
}

func (m Model) renderBreadcrumb() string {
	if len(m.breadcrumb) == 0 {
		return styles.TitleStyle.Render("Home")
	}

	parts := make([]string, len(m.breadcrumb))
	for i, t := range m.breadcrumb {
		if i == len(m.breadcrumb)-1 {
			parts[i] = styles.SelectedStyle.Render(t.Name)
		} else {
			parts[i] = styles.DimStyle.Render(t.Name)
		}
	}

	return styles.TitleStyle.Render("Home") + styles.DimStyle.Render(" > ") + strings.Join(parts, styles.DimStyle.Render(" > "))
}

func (m Model) viewHelp() string {
	if m.state == StateViewingItems {
		return styles.HelpStyle.Render("C-p/C-n/↑/↓: navigate | Enter/C-f/→: select | ESC/C-b/←: back | q: quit")
	}
	return styles.HelpStyle.Render("C-p/C-n/↑/↓: navigate | Enter/C-f/→: drill down | ESC/C-b/←: back | q: quit")
}

// Run starts the application
func Run(handler *service.Handler) error {
	model := NewModel(handler)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
