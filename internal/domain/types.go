package domain

// ItemType represents a category of items in a hierarchy
type ItemType struct {
	ID          int64
	ParentID    *int64 // nil for root types
	Name        string
	Description string // empty string if not set
	Depth       int64
	CreatedAt   string
}

// Item represents an inventory item
type Item struct {
	ID         int64
	Name       string
	ItemTypeID int64
	Quantity   float64 // Numeric quantity (supports decimals)
	UnitType   string  // One of: "Count", "Grams", "Liters"
	CreatedAt  string
}
