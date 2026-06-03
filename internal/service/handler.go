package service

import (
	"context"

	"home-inventory-system/internal/repository"
)

// Handler processes queries and commands by transforming them to repository operations
type Handler struct {
	repo *repository.Repository
}

// NewHandler creates a new service handler
func NewHandler(repo *repository.Repository) *Handler {
	return &Handler{repo: repo}
}

// Repository returns the underlying repository (for seed operations)
func (h *Handler) Repository() *repository.Repository {
	return h.repo
}

// HandleQuery processes a query and returns the appropriate result
func (h *Handler) HandleQuery(ctx context.Context, q Query) QueryResult {
	switch query := q.(type) {
	case GetItemQuery:
		return h.handleGetItem(ctx, query)
	case GetItemTypeQuery:
		return h.handleGetItemType(ctx, query)
	case ListRootTypesQuery:
		return h.handleListRootTypes(ctx)
	case ListChildTypesQuery:
		return h.handleListChildTypes(ctx, query)
	case GetTypePathQuery:
		return h.handleGetTypePath(ctx, query)
	case IsLeafTypeQuery:
		return h.handleIsLeafType(ctx, query)
	case ListItemsByTypeQuery:
		return h.handleListItemsByType(ctx, query)
	case CountItemsByTypeQuery:
		return h.handleCountItemsByType(ctx, query)
	case ListLeafTypesQuery:
		return h.handleListLeafTypes(ctx)
	default:
		return nil
	}
}

// HandleCommand processes a command and returns the appropriate result
func (h *Handler) HandleCommand(ctx context.Context, c Command) CommandResult {
	switch cmd := c.(type) {
	case CreateRootTypeCommand:
		return h.handleCreateRootType(ctx, cmd)
	case CreateChildTypeCommand:
		return h.handleCreateChildType(ctx, cmd)
	case CreateItemCommand:
		return h.handleCreateItem(ctx, cmd)
	case UpdateCategoryCommand:
		return h.handleUpdateCategory(ctx, cmd)
	case DeleteItemCommand:
		return h.handleDeleteItem(ctx, cmd)
	case DeleteItemTypeCommand:
		return h.handleDeleteItemType(ctx, cmd)
	default:
		return nil
	}
}

// --- Query Handlers ---

func (h *Handler) handleGetItem(ctx context.Context, q GetItemQuery) GetItemResult {
	item, err := h.repo.GetItem(ctx, q.ID)
	return GetItemResult{Item: item, Err: err}
}

func (h *Handler) handleGetItemType(ctx context.Context, q GetItemTypeQuery) GetItemTypeResult {
	itemType, err := h.repo.GetItemType(ctx, q.ID)
	return GetItemTypeResult{Type: itemType, Err: err}
}

func (h *Handler) handleListRootTypes(ctx context.Context) ListRootTypesResult {
	types, err := h.repo.GetRootTypes(ctx)
	return ListRootTypesResult{Types: types, Err: err}
}

func (h *Handler) handleListChildTypes(ctx context.Context, q ListChildTypesQuery) ListChildTypesResult {
	types, err := h.repo.GetChildTypes(ctx, q.ParentID)
	return ListChildTypesResult{Types: types, Err: err}
}

func (h *Handler) handleGetTypePath(ctx context.Context, q GetTypePathQuery) GetTypePathResult {
	path, err := h.repo.GetTypePath(ctx, q.TypeID)
	return GetTypePathResult{Path: path, Err: err}
}

func (h *Handler) handleIsLeafType(ctx context.Context, q IsLeafTypeQuery) IsLeafTypeResult {
	isLeaf, err := h.repo.IsLeafType(ctx, q.TypeID)
	return IsLeafTypeResult{IsLeaf: isLeaf, Err: err}
}

func (h *Handler) handleListItemsByType(ctx context.Context, q ListItemsByTypeQuery) ListItemsByTypeResult {
	items, err := h.repo.ListItemsByType(ctx, q.TypeID)
	return ListItemsByTypeResult{Items: items, Err: err}
}

func (h *Handler) handleCountItemsByType(ctx context.Context, q CountItemsByTypeQuery) CountItemsByTypeResult {
	count, err := h.repo.CountItemsByType(ctx, q.TypeID)
	return CountItemsByTypeResult{Count: count, Err: err}
}

func (h *Handler) handleListLeafTypes(ctx context.Context) ListLeafTypesResult {
	types, err := h.repo.ListLeafTypes(ctx)
	return ListLeafTypesResult{Types: types, Err: err}
}

// --- Command Handlers ---

func (h *Handler) handleCreateRootType(ctx context.Context, cmd CreateRootTypeCommand) CreateRootTypeResult {
	// Validate input
	if cmd.Name == "" {
		return CreateRootTypeResult{
			ValidationError: &ValidationError{Field: "name", Message: "Name is required"},
		}
	}

	itemType, err := h.repo.CreateRootType(ctx, cmd.Name, cmd.Description)
	return CreateRootTypeResult{Type: itemType, Err: err}
}

func (h *Handler) handleCreateChildType(ctx context.Context, cmd CreateChildTypeCommand) CreateChildTypeResult {
	// Validate input
	if cmd.Name == "" {
		return CreateChildTypeResult{
			ValidationError: &ValidationError{Field: "name", Message: "Name is required"},
		}
	}

	itemType, err := h.repo.CreateChildType(ctx, cmd.ParentID, cmd.Name, cmd.Description)
	return CreateChildTypeResult{Type: itemType, Err: err}
}

func (h *Handler) handleCreateItem(ctx context.Context, cmd CreateItemCommand) CreateItemResult {
	// Validate input
	if cmd.Name == "" {
		return CreateItemResult{
			ValidationError: &ValidationError{Field: "name", Message: "Name is required"},
		}
	}
	if cmd.Quantity <= 0 {
		return CreateItemResult{
			ValidationError: &ValidationError{Field: "quantity", Message: "Quantity must be greater than 0"},
		}
	}

	item, err := h.repo.CreateItem(ctx, cmd.Name, cmd.TypeID, cmd.Quantity, cmd.UnitType)
	return CreateItemResult{Item: item, Err: err}
}

func (h *Handler) handleUpdateCategory(ctx context.Context, cmd UpdateCategoryCommand) UpdateCategoryResult {
	if cmd.Name == "" {
		return UpdateCategoryResult{
			ValidationError: &ValidationError{Field: "name", Message: "Name is required"},
		}
	}
	if err := h.repo.UpdateItemType(ctx, cmd.ID, cmd.Name, cmd.Description); err != nil {
		return UpdateCategoryResult{Err: err}
	}
	updated, err := h.repo.GetItemType(ctx, cmd.ID)
	return UpdateCategoryResult{Type: updated, Err: err}
}

func (h *Handler) handleDeleteItem(ctx context.Context, cmd DeleteItemCommand) DeleteItemResult {
	err := h.repo.DeleteItem(ctx, cmd.ID)
	return DeleteItemResult{Err: err}
}

func (h *Handler) handleDeleteItemType(ctx context.Context, cmd DeleteItemTypeCommand) DeleteItemTypeResult {
	err := h.repo.DeleteItemType(ctx, cmd.ID)
	return DeleteItemTypeResult{Err: err}
}
