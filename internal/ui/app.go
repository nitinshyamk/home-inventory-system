package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"home-inventory-system/internal/domain"
	"home-inventory-system/internal/service"
	"home-inventory-system/internal/ui/components/categoryedit"
	"home-inventory-system/internal/ui/components/categoryform"
	"home-inventory-system/internal/ui/components/categorypicker"
	"home-inventory-system/internal/ui/components/itemedit"
	"home-inventory-system/internal/ui/components/itemform"
	"home-inventory-system/internal/ui/components/itemlist"
	"home-inventory-system/internal/ui/formcontroller"
	"home-inventory-system/internal/ui/messages"
)

// rightPaneContent holds cached data for the right pane detail view.
type rightPaneContent struct {
	// Category detail (category != nil)
	category   *domain.ItemType
	isLeaf     bool
	childCount int64
	itemCount  int64

	// Item detail (item != nil)
	item         *domain.Item
	itemTypePath []domain.ItemType // breadcrumb at the time the item was highlighted
}

// Model is the main application model
type Model struct {
	state              AppState
	itemList           itemlist.Model
	categoryForm       categoryform.Model
	categoryEditForm   categoryedit.Model
	itemEditForm       itemedit.Model
	categoryPicker     categorypicker.Model
	itemForm           itemform.Model
	deleteConfirmFocus int // 0=Confirm, 1=Cancel (for delete confirmation dialogs)
	formController   *formcontroller.Controller
	handler          *service.Handler
	err              error
	width            int
	height           int
	breadcrumb       []domain.ItemType // Current path in hierarchy
	currentTypeID    *int64            // nil = root level
	items            []domain.Item     // Items at current leaf node
	rightPane        rightPaneContent  // Cached detail for the right pane
}

// NewModel creates a new application model
func NewModel(handler *service.Handler) Model {
	return Model{
		state:          StateLoading,
		itemList:       itemlist.New(),
		formController: formcontroller.New(handler),
		handler:        handler,
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

	case messages.FormCancelledMsg:
		return handleFormCancelled(m)

	case messages.CategoryCreatedMsg:
		return handleCategoryCreated(m, msg)

	case messages.ItemCreatedMsg:
		return handleItemCreated(m, msg)

	case messages.RightPaneDetailMsg:
		return handleRightPaneDetail(m, msg)

	case messages.CategoryUpdatedMsg:
		return handleCategoryUpdated(m, msg)

	case messages.ItemUpdatedMsg:
		return handleItemUpdated(m, msg)

	case messages.LeafTypesWithPathsLoadedMsg:
		return handleLeafTypesWithPathsLoaded(m, msg)

	case messages.ItemDeletedMsg:
		return handleItemDeleted(m, msg)
	}

	// Pass messages to item list when in browsing state
	if m.state.AllowsItemListDelegation() {
		return delegateToItemList(m, msg)
	}

	return m, nil
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// When in right-pane edit states, delegate all keys to the relevant form (except ctrl+c)
	if m.state == StateEditingCategory {
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return delegateToCategoryEditForm(m, msg)
	}
	if m.state == StateEditingItem {
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return delegateToItemEditForm(m, msg)
	}
	if m.state == StatePickingCategory {
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return delegateToCategoryPicker(m, msg)
	}
	if m.state == StateDeletingItem {
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return handleDeleteItemConfirmKey(m, msg)
	}

	// When in modal form state, delegate all keys to the form (except ctrl+c)
	if m.state == StateCreatingCategory {
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return delegateToCategoryForm(m, msg)
	}
	if m.state == StateCreatingItem {
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return delegateToItemForm(m, msg)
	}

	// Global key handling for other states
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "esc", "ctrl+b", "left":
		return navigateUp(m)

	case "enter", "ctrl+f", "right":
		return selectCurrent(m)

	case "e":
		return openEditForm(m)

	case "d":
		return openDeleteConfirm(m)
	}

	// Pass to itemlist for navigation; also refreshes right pane detail
	if m.state.AllowsItemListDelegation() {
		return delegateToItemList(m, msg)
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
