package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"home-inventory-system/internal/domain"
	"home-inventory-system/internal/service"
	"home-inventory-system/internal/ui/components/itemlist"
	"home-inventory-system/internal/ui/messages"
	"home-inventory-system/internal/ui/styles"
)

// Model is the main application model
type Model struct {
	state         AppState
	itemList      itemlist.Model
	handler       *service.Handler
	err           error
	width         int
	height        int
	breadcrumb    []domain.ItemType // Current path in hierarchy
	currentTypeID *int64            // nil = root level
	items         []domain.Item     // Items at current leaf node
	selectedItem  *domain.Item
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
	return loadRootTypes(m.handler)
}

// Update handles all messages for the application
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return handleWindowResize(m, msg)

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case messages.TypesLoadedMsg:
		return handleTypesLoaded(m, msg)

	case messages.LeafItemsLoadedMsg:
		return handleLeafItemsLoaded(m, msg)

	case messages.BreadcrumbLoadedMsg:
		return handleBreadcrumbLoaded(m, msg)

	case messages.TypeSelectedMsg:
		return handleTypeSelected(m, msg)

	case messages.LeafCheckCompleteMsg:
		return handleLeafCheckComplete(m, msg)

	case messages.ErrorMsg:
		return setError(m, msg.Err)
	}

	// Pass messages to item list when in browsing state
	if m.state.AllowsItemListDelegation() {
		return delegateToItemList(m, msg)
	}

	return m, nil
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "esc", "ctrl+b", "left":
		return navigateUp(m)

	case "enter", "ctrl+f", "right":
		return selectCurrent(m)
	}

	// Pass to itemlist for navigation
	if m.state.AllowsItemListDelegation() {
		var cmd tea.Cmd
		m.itemList, cmd = m.itemList.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the application
func (m Model) View() string {
	switch m.state {
	case StateLoading:
		return m.viewLoading()
	case StateBrowsingTypes:
		return m.renderMainView(m.itemList.View())
	case StateViewingItems:
		return m.renderMainView(m.itemList.View())
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

// renderMainView renders the common layout with breadcrumb, content, and help
func (m Model) renderMainView(content string) string {
	var b strings.Builder
	b.WriteString(renderBreadcrumb(m.breadcrumb))
	b.WriteString("\n\n")
	b.WriteString(content)
	b.WriteString("\n")
	b.WriteString(m.viewHelp())
	return b.String()
}

func (m Model) viewItemSelected() string {
	if m.selectedItem == nil {
		return m.renderMainView(m.itemList.View())
	}

	var b strings.Builder

	// Breadcrumb
	b.WriteString(renderBreadcrumb(m.breadcrumb))
	b.WriteString("\n\n")

	b.WriteString(styles.TitleStyle.Render("Item Details"))
	b.WriteString("\n\n")
	b.WriteString(styles.NormalStyle.Render(fmt.Sprintf("Name: %s", m.selectedItem.Name)))
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
