package messages

import "home-inventory-system/internal/repository"

// ItemsLoadedMsg is sent when items have been loaded from the database
type ItemsLoadedMsg struct {
	Items []repository.ItemWithType
	Err   error
}

// ItemSelectedMsg is sent when a user selects an item
type ItemSelectedMsg struct {
	Item repository.ItemWithType
}

// ErrorMsg is sent when an error occurs
type ErrorMsg struct {
	Err error
}

// TypesLoadedMsg is sent when item types have been loaded
type TypesLoadedMsg struct {
	Types    []repository.ItemType
	ParentID *int64 // nil for root types
	Err      error
}

// TypeSelectedMsg is sent when a user selects a type (to drill down)
type TypeSelectedMsg struct {
	Type repository.ItemType
}

// NavigateUpMsg is sent when user wants to go up in the hierarchy
type NavigateUpMsg struct{}

// BreadcrumbLoadedMsg is sent when breadcrumb path is loaded
type BreadcrumbLoadedMsg struct {
	Path []repository.ItemType
	Err  error
}

// LeafItemsLoadedMsg is sent when items at a leaf node are loaded
type LeafItemsLoadedMsg struct {
	TypeID int64
	Items  []repository.Item
	Err    error
}

// ItemCountLoadedMsg is sent when item count for a type is loaded
type ItemCountLoadedMsg struct {
	TypeID int64
	Count  int64
	Err    error
}

// ChildCountsLoadedMsg is sent when child counts are loaded for display
type ChildCountsLoadedMsg struct {
	Counts map[int64]int64 // TypeID -> item count
	Err    error
}
