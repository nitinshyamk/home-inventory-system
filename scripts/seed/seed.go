// Package seed provides functions to populate the database with sample data.
// It uses a declarative hierarchy structure for easy maintenance and extension.
package seed

import (
	"context"
	"fmt"

	"home-inventory-system/internal/domain"
	"home-inventory-system/internal/repository"
)

// ItemData represents an item with quantity and unit
type ItemData struct {
	Name     string
	Quantity float64
	Unit     string // One of: "Count", "Grams", "Liters"
}

// Node represents a category or item in the hierarchy
type Node struct {
	Name        string     // Name of the category
	Description string     // Optional description (for categories only)
	Children    []Node     // Subcategories
	Items       []ItemData // Items at this leaf node
}

// Hierarchy represents the complete seed data
type Hierarchy []Node

// DefaultHierarchy is the comprehensive seed data for home inventory
var DefaultHierarchy = Hierarchy{
	{
		Name: "Kitchen", Description: "Kitchen and cooking related items",
		Children: []Node{
			{
				Name: "Pantry Items", Description: "Food storage items",
				Children: []Node{
					{Name: "Spices", Description: "Cooking spices and seasonings", Items: []ItemData{
						{Name: "Black Cardamom", Quantity: 50.0, Unit: "Grams"},
						{Name: "Cinnamon Sticks", Quantity: 30.0, Unit: "Grams"},
						{Name: "Cumin Seeds", Quantity: 100.0, Unit: "Grams"},
						{Name: "Turmeric Powder", Quantity: 75.0, Unit: "Grams"},
						{Name: "Red Chili Flakes", Quantity: 40.0, Unit: "Grams"},
					}},
					{
						Name: "Grains", Description: "Rice and grains",
						Children: []Node{
							{Name: "Rice", Description: "Various types of rice", Items: []ItemData{
								{Name: "Jasmine Rice", Quantity: 2000.0, Unit: "Grams"},
								{Name: "Basmati Rice", Quantity: 5000.0, Unit: "Grams"},
								{Name: "Brown Rice", Quantity: 1000.0, Unit: "Grams"},
								{Name: "Arborio Rice", Quantity: 500.0, Unit: "Grams"},
							}},
						},
					},
				},
			},
			{
				Name: "Disposable Containers", Description: "Single-use storage containers",
				Children: []Node{
					{Name: "Ziploc Bags", Description: "Resealable plastic bags", Items: []ItemData{
						{Name: "1/2 Gallon Ziploc Bags", Quantity: 20.0, Unit: "Count"},
						{Name: "Quart Ziploc Bags", Quantity: 50.0, Unit: "Count"},
						{Name: "Snack Size Ziploc Bags", Quantity: 100.0, Unit: "Count"},
					}},
				},
			},
			{Name: "Appliances", Description: "Kitchen appliances", Items: []ItemData{
				{Name: "Blender", Quantity: 1.0, Unit: "Count"},
				{Name: "Coffee Maker", Quantity: 1.0, Unit: "Count"},
				{Name: "Toaster", Quantity: 1.0, Unit: "Count"},
				{Name: "Food Processor", Quantity: 1.0, Unit: "Count"},
				{Name: "Stand Mixer", Quantity: 1.0, Unit: "Count"},
			}},
		},
	},
	{
		Name: "Electronics", Description: "Electronic devices and gadgets",
		Children: []Node{
			{
				Name: "Computers", Description: "Computing devices",
				Children: []Node{
					{Name: "Laptops", Description: "Portable computers", Items: []ItemData{
						{Name: "MacBook Pro 14\"", Quantity: 1.0, Unit: "Count"},
						{Name: "Dell XPS 15", Quantity: 1.0, Unit: "Count"},
						{Name: "ThinkPad X1 Carbon", Quantity: 1.0, Unit: "Count"},
					}},
					{Name: "Accessories", Description: "Computer accessories", Items: []ItemData{
						{Name: "Wireless Mouse", Quantity: 2.0, Unit: "Count"},
						{Name: "Mechanical Keyboard", Quantity: 1.0, Unit: "Count"},
						{Name: "USB-C Hub", Quantity: 3.0, Unit: "Count"},
						{Name: "External SSD", Quantity: 2.0, Unit: "Count"},
					}},
				},
			},
		},
	},
	{
		Name: "Home tools", Description: "Home electronics and tools",
		Children: []Node{
			{
				Name: "Tools", Description: "Hand and power tools",
				Children: []Node{
					{Name: "Hand Tools", Description: "Manual tools", Items: []ItemData{
						{Name: "Hammer", Quantity: 2.0, Unit: "Count"},
						{Name: "Screwdriver Set", Quantity: 1.0, Unit: "Count"},
						{Name: "Wrench Set", Quantity: 1.0, Unit: "Count"},
						{Name: "Pliers", Quantity: 3.0, Unit: "Count"},
						{Name: "Tape Measure", Quantity: 2.0, Unit: "Count"},
					}},
					{Name: "Power Tools", Description: "Electric and battery-powered tools", Items: []ItemData{
						{Name: "Drill", Quantity: 1.0, Unit: "Count"},
						{Name: "Circular Saw", Quantity: 1.0, Unit: "Count"},
						{Name: "Jigsaw", Quantity: 1.0, Unit: "Count"},
						{Name: "Sander", Quantity: 1.0, Unit: "Count"},
						{Name: "Impact Driver", Quantity: 1.0, Unit: "Count"},
					}},
				},
			},
		},
	},
}

// MinimalHierarchy is a small subset for quick testing
var MinimalHierarchy = Hierarchy{
	{
		Name: "Kitchen", Description: "Kitchen items",
		Children: []Node{
			{
				Name: "Pantry",
				Children: []Node{
					{Name: "Spices", Items: []ItemData{
						{Name: "Cumin", Quantity: 50.0, Unit: "Grams"},
						{Name: "Turmeric", Quantity: 75.0, Unit: "Grams"},
					}},
				},
			},
		},
	},
	{
		Name: "Garage", Description: "Garage items",
		Children: []Node{
			{Name: "Tools", Items: []ItemData{
				{Name: "Hammer", Quantity: 1.0, Unit: "Count"},
			}},
		},
	},
}

// Seed populates the database with comprehensive sample hierarchical data.
func Seed(ctx context.Context, repo *repository.Repository) error {
	return SeedHierarchy(ctx, repo, DefaultHierarchy)
}

// SeedMinimal creates minimal test data for quick testing.
func SeedMinimal(ctx context.Context, repo *repository.Repository) error {
	return SeedHierarchy(ctx, repo, MinimalHierarchy)
}

// SeedHierarchy inserts a complete hierarchy into the database
func SeedHierarchy(ctx context.Context, repo *repository.Repository, h Hierarchy) error {
	for _, node := range h {
		if err := insertNode(ctx, repo, node, nil); err != nil {
			return err
		}
	}
	return nil
}

// insertNode recursively inserts a node and its children/items
func insertNode(ctx context.Context, repo *repository.Repository, node Node, parentID *int64) error {
	var itemType *domain.ItemType
	var err error

	if parentID == nil {
		itemType, err = repo.CreateRootType(ctx, node.Name, node.Description)
	} else {
		itemType, err = repo.CreateChildType(ctx, *parentID, node.Name, node.Description)
	}
	if err != nil {
		return fmt.Errorf("failed to create type %s: %w", node.Name, err)
	}

	// Insert items at this node
	for _, item := range node.Items {
		if _, err := repo.CreateItem(ctx, item.Name, itemType.ID, item.Quantity, item.Unit); err != nil {
			return fmt.Errorf("failed to create item %s: %w", item.Name, err)
		}
	}

	// Recursively insert children
	for _, child := range node.Children {
		if err := insertNode(ctx, repo, child, &itemType.ID); err != nil {
			return err
		}
	}

	return nil
}
