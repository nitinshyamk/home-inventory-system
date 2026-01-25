package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"home-inventory-system/internal/domain"
	"home-inventory-system/internal/service"
	"home-inventory-system/internal/ui/components/form"
	"home-inventory-system/internal/ui/components/itemlist"
	"home-inventory-system/internal/ui/messages"
)

// Model is the main application model
type Model struct {
	state         AppState
	itemList      itemlist.Model
	form          form.Model
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

	case messages.FormSubmittedMsg:
		return handleFormSubmitted(m, msg)

	case messages.FormCancelledMsg:
		return handleFormCancelled(m)

	case messages.CategoryCreatedMsg:
		return handleCategoryCreated(m, msg)

	case messages.ItemCreatedMsg:
		return handleItemCreated(m, msg)
	}

	// Pass messages to item list when in browsing state
	if m.state.AllowsItemListDelegation() {
		return delegateToItemList(m, msg)
	}

	// Pass messages to form when in creation state
	if m.state == StateCreatingCategory || m.state == StateCreatingItem {
		return delegateToForm(m, msg)
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
	return renderView(m)
}

// Run starts the application
func Run(handler *service.Handler) error {
	model := NewModel(handler)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
