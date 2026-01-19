package service

import "home-inventory-system/internal/repository"

// Query represents a read-only operation
type Query interface {
	isQuery()
}

// Command represents a state-modifying operation
type Command interface {
	isCommand()
}

// --- Queries ---

// ListItemsQuery requests all items with their type information
type ListItemsQuery struct{}

func (ListItemsQuery) isQuery() {}

// ListItemTypesQuery requests all item types
type ListItemTypesQuery struct{}

func (ListItemTypesQuery) isQuery() {}

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

// --- Query Results ---

// QueryResult represents the result of a query operation
type QueryResult interface {
	isQueryResult()
}

// ListItemsResult contains the result of ListItemsQuery
type ListItemsResult struct {
	Items []repository.ItemWithType
	Err   error
}

func (ListItemsResult) isQueryResult() {}

// ListItemTypesResult contains the result of ListItemTypesQuery
type ListItemTypesResult struct {
	Types []repository.ItemType
	Err   error
}

func (ListItemTypesResult) isQueryResult() {}

// GetItemResult contains the result of GetItemQuery
type GetItemResult struct {
	Item *repository.Item
	Err  error
}

func (GetItemResult) isQueryResult() {}

// GetItemTypeResult contains the result of GetItemTypeQuery
type GetItemTypeResult struct {
	Type *repository.ItemType
	Err  error
}

func (GetItemTypeResult) isQueryResult() {}

// ListRootTypesResult contains the result of ListRootTypesQuery
type ListRootTypesResult struct {
	Types []repository.ItemType
	Err   error
}

func (ListRootTypesResult) isQueryResult() {}

// ListChildTypesResult contains the result of ListChildTypesQuery
type ListChildTypesResult struct {
	Types []repository.ItemType
	Err   error
}

func (ListChildTypesResult) isQueryResult() {}

// GetTypePathResult contains the breadcrumb path result
type GetTypePathResult struct {
	Path []repository.ItemType
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
	Items []repository.Item
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
	Types []repository.ItemType
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
	Name   string
	TypeID int64
}

func (CreateItemCommand) isCommand() {}

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
	Type *repository.ItemType
	Err  error
}

func (CreateRootTypeResult) isCommandResult() {}

// CreateChildTypeResult contains the created child type
type CreateChildTypeResult struct {
	Type *repository.ItemType
	Err  error
}

func (CreateChildTypeResult) isCommandResult() {}

// CreateItemResult contains the created item
type CreateItemResult struct {
	Item *repository.Item
	Err  error
}

func (CreateItemResult) isCommandResult() {}

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
