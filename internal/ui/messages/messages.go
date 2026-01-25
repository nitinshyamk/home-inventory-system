package messages

import (
	"home-inventory-system/internal/domain"
	"home-inventory-system/internal/ui/components/form"
)

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

// FormSubmittedMsg is sent when a form is submitted
type FormSubmittedMsg struct {
	FormType form.FormType
	Data     map[string]interface{}
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
