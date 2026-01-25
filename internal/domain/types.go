package domain

// UnitType represents the unit of measurement for an item's quantity
type UnitType string

const (
	UnitTypeCount  UnitType = "Count"
	UnitTypeGrams  UnitType = "Grams"
	UnitTypeLiters UnitType = "Liters"
)

// String returns the string representation of the UnitType
func (u UnitType) String() string {
	return string(u)
}

// IsValid checks if the UnitType is one of the allowed values
func (u UnitType) IsValid() bool {
	switch u {
	case UnitTypeCount, UnitTypeGrams, UnitTypeLiters:
		return true
	default:
		return false
	}
}

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
	Quantity   float64  // Numeric quantity (supports decimals)
	UnitType   UnitType // Unit of measurement
	CreatedAt  string
}
