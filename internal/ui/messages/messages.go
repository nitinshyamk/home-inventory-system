package messages

import "home-inventory-system/internal/domain"

// ErrorMsg is sent when an error occurs
type ErrorMsg struct {
	Err error
}

// TypesLoadedMsg is sent when item types have been loaded
type TypesLoadedMsg struct {
	Types    []domain.ItemType
	ParentID *int64 // nil for root types
	Err      error
}

// TypeSelectedMsg is sent when a user selects a type (to drill down)
type TypeSelectedMsg struct {
	Type domain.ItemType
}

// BreadcrumbLoadedMsg is sent when breadcrumb path is loaded
type BreadcrumbLoadedMsg struct {
	Path []domain.ItemType
	Err  error
}

// LeafItemsLoadedMsg is sent when items at a leaf node are loaded
type LeafItemsLoadedMsg struct {
	TypeID int64
	Items  []domain.Item
	Err    error
}

// LeafCheckCompleteMsg is sent when leaf type check completes
type LeafCheckCompleteMsg struct {
	TypeID int64
	IsLeaf bool
	Err    error
}

// FormCancelledMsg is sent when a form is cancelled
type FormCancelledMsg struct{}

// CategoryCreatedMsg is sent when a category is created
type CategoryCreatedMsg struct {
	Category *domain.ItemType
	Err      error
}

// ItemCreatedMsg is sent when an item is created
type ItemCreatedMsg struct {
	Item *domain.Item
	Err  error
}

// CategoryDeletedMsg is sent when a category is successfully deleted
type CategoryDeletedMsg struct {
	DeletedID int64
	ParentID  *int64
	Err       error
}

// CategoryChildSummaryLoadedMsg carries the child summary for the delete confirmation
type CategoryChildSummaryLoadedMsg struct {
	ChildTypeCount int
	ItemCount      int
	ParentID       *int64
	Err            error
}

// ItemDeletedMsg is sent when an item is successfully deleted
type ItemDeletedMsg struct {
	DeletedID int64
	TypeID    int64
	Err       error
}

// ItemUpdatedMsg is sent when an item is successfully updated
type ItemUpdatedMsg struct {
	Item *domain.Item
	Err  error
}

// CategoryUpdatedMsg is sent when a category is successfully updated
type CategoryUpdatedMsg struct {
	Category *domain.ItemType
	Err      error
}

// LeafTypesWithPathsLoadedMsg carries leaf types and their ancestor paths for the category picker.
type LeafTypesWithPathsLoadedMsg struct {
	LeafTypes []domain.ItemType
	Paths     map[int64][]domain.ItemType // typeID → path from root to that type
	Err       error
}

// RightPaneDetailMsg carries loaded count data for the right pane category detail.
type RightPaneDetailMsg struct {
	CategoryID int64
	IsLeaf     bool
	ChildCount int64
	ItemCount  int64
	Err        error
}
