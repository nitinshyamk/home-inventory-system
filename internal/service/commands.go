package service

import "home-inventory-system/internal/domain"

// ValidationError represents a validation failure
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

// Query represents a read-only operation
type Query interface {
	isQuery()
}

// Command represents a state-modifying operation
type Command interface {
	isCommand()
}

// --- Queries ---

// GetItemQuery requests a single item by ID
type GetItemQuery struct {
	ID int64
}

func (GetItemQuery) isQuery() {}

// GetItemTypeQuery requests a single item type by ID
type GetItemTypeQuery struct {
	ID int64
}

func (GetItemTypeQuery) isQuery() {}

// ListRootTypesQuery requests all root-level item types
type ListRootTypesQuery struct{}

func (ListRootTypesQuery) isQuery() {}

// ListChildTypesQuery requests all direct children of a parent type
type ListChildTypesQuery struct {
	ParentID int64
}

func (ListChildTypesQuery) isQuery() {}

// GetTypePathQuery requests the breadcrumb path from root to a type
type GetTypePathQuery struct {
	TypeID int64
}

func (GetTypePathQuery) isQuery() {}

// IsLeafTypeQuery checks if a type has no children
type IsLeafTypeQuery struct {
	TypeID int64
}

func (IsLeafTypeQuery) isQuery() {}

// ListItemsByTypeQuery requests all items of a specific type
type ListItemsByTypeQuery struct {
	TypeID int64
}

func (ListItemsByTypeQuery) isQuery() {}

// CountItemsByTypeQuery requests the count of items for a type
type CountItemsByTypeQuery struct {
	TypeID int64
}

func (CountItemsByTypeQuery) isQuery() {}

// ListLeafTypesQuery requests all types that have no children
type ListLeafTypesQuery struct{}

func (ListLeafTypesQuery) isQuery() {}

// GetCategoryChildSummaryQuery requests child type count, item count, and parent ID for a category
type GetCategoryChildSummaryQuery struct {
	ID int64
}

func (GetCategoryChildSummaryQuery) isQuery() {}

// GetCategoryChildSummaryResult contains the category child summary
type GetCategoryChildSummaryResult struct {
	ChildTypeCount int
	ItemCount      int
	ParentID       *int64
	Err            error
}

func (GetCategoryChildSummaryResult) isQueryResult() {}

// --- Query Results ---

// QueryResult represents the result of a query operation
type QueryResult interface {
	isQueryResult()
}

// GetItemResult contains the result of GetItemQuery
type GetItemResult struct {
	Item *domain.Item
	Err  error
}

func (GetItemResult) isQueryResult() {}

// GetItemTypeResult contains the result of GetItemTypeQuery
type GetItemTypeResult struct {
	Type *domain.ItemType
	Err  error
}

func (GetItemTypeResult) isQueryResult() {}

// ListRootTypesResult contains the result of ListRootTypesQuery
type ListRootTypesResult struct {
	Types []domain.ItemType
	Err   error
}

func (ListRootTypesResult) isQueryResult() {}

// ListChildTypesResult contains the result of ListChildTypesQuery
type ListChildTypesResult struct {
	Types []domain.ItemType
	Err   error
}

func (ListChildTypesResult) isQueryResult() {}

// GetTypePathResult contains the breadcrumb path result
type GetTypePathResult struct {
	Path []domain.ItemType
	Err  error
}

func (GetTypePathResult) isQueryResult() {}

// IsLeafTypeResult contains the leaf check result
type IsLeafTypeResult struct {
	IsLeaf bool
	Err    error
}

func (IsLeafTypeResult) isQueryResult() {}

// ListItemsByTypeResult contains items for a specific type
type ListItemsByTypeResult struct {
	Items []domain.Item
	Err   error
}

func (ListItemsByTypeResult) isQueryResult() {}

// CountItemsByTypeResult contains the item count for a type
type CountItemsByTypeResult struct {
	Count int64
	Err   error
}

func (CountItemsByTypeResult) isQueryResult() {}

// ListLeafTypesResult contains all leaf types
type ListLeafTypesResult struct {
	Types []domain.ItemType
	Err   error
}

func (ListLeafTypesResult) isQueryResult() {}

// --- Commands ---

// CreateRootTypeCommand creates a new root-level item type
type CreateRootTypeCommand struct {
	Name        string
	Description string
}

func (CreateRootTypeCommand) isCommand() {}

// CreateChildTypeCommand creates a new child item type
type CreateChildTypeCommand struct {
	ParentID    int64
	Name        string
	Description string
}

func (CreateChildTypeCommand) isCommand() {}

// CreateItemCommand creates a new item
type CreateItemCommand struct {
	Name     string
	TypeID   int64
	Quantity float64         // Numeric quantity (supports decimals)
	UnitType domain.UnitType // Unit of measurement
}

func (CreateItemCommand) isCommand() {}

// UpdateItemCommand updates an existing item's fields
type UpdateItemCommand struct {
	ID       int64
	Name     string
	TypeID   int64
	Quantity float64
	UnitType domain.UnitType
}

func (UpdateItemCommand) isCommand() {}

// UpdateItemResult contains the updated item
type UpdateItemResult struct {
	Item            *domain.Item
	ValidationError *ValidationError
	Err             error
}

func (UpdateItemResult) isCommandResult() {}

// UpdateCategoryCommand updates an existing item type's name and description
type UpdateCategoryCommand struct {
	ID          int64
	Name        string
	Description string
}

func (UpdateCategoryCommand) isCommand() {}

// DeleteItemCommand deletes an item
type DeleteItemCommand struct {
	ID int64
}

func (DeleteItemCommand) isCommand() {}

// DeleteItemTypeCommand deletes an item type
type DeleteItemTypeCommand struct {
	ID int64
}

func (DeleteItemTypeCommand) isCommand() {}

// --- Command Results ---

// CommandResult represents the result of a command operation
type CommandResult interface {
	isCommandResult()
}

// CreateRootTypeResult contains the created root type
type CreateRootTypeResult struct {
	Type            *domain.ItemType
	ValidationError *ValidationError
	Err             error
}

func (CreateRootTypeResult) isCommandResult() {}

// CreateChildTypeResult contains the created child type
type CreateChildTypeResult struct {
	Type            *domain.ItemType
	ValidationError *ValidationError
	Err             error
}

func (CreateChildTypeResult) isCommandResult() {}

// CreateItemResult contains the created item
type CreateItemResult struct {
	Item            *domain.Item
	ValidationError *ValidationError
	Err             error
}

func (CreateItemResult) isCommandResult() {}

// UpdateCategoryResult contains the updated item type
type UpdateCategoryResult struct {
	Type            *domain.ItemType
	ValidationError *ValidationError
	Err             error
}

func (UpdateCategoryResult) isCommandResult() {}

// DeleteItemResult contains the delete operation result
type DeleteItemResult struct {
	Err error
}

func (DeleteItemResult) isCommandResult() {}

// DeleteItemTypeResult contains the delete operation result
type DeleteItemTypeResult struct {
	Err error
}

func (DeleteItemTypeResult) isCommandResult() {}
