// Package seed provides functions to populate the database with sample data.
// It uses a declarative hierarchy structure for easy maintenance and extension.
package seed

import (
	"context"
	"fmt"

	"home-inventory-system/internal/domain"
	"home-inventory-system/internal/repository"
)

// Node represents a category or item in the hierarchy
type Node struct {
	Name        string   // Name of the category
	Description string   // Optional description (for categories only)
	Children    []Node   // Subcategories
	Items       []string // Items at this leaf node
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
					{Name: "Spices", Description: "Cooking spices and seasonings", Items: []string{
						"Black Cardamom", "Cinnamon Sticks", "Cumin Seeds", "Turmeric Powder", "Red Chili Flakes",
					}},
					{
						Name: "Grains", Description: "Rice and grains",
						Children: []Node{
							{Name: "Rice", Description: "Various types of rice", Items: []string{
								"Jasmine Rice", "Basmati Rice", "Brown Rice", "Arborio Rice",
							}},
						},
					},
				},
			},
			{
				Name: "Disposable Containers", Description: "Single-use storage containers",
				Children: []Node{
					{Name: "Ziploc Bags", Description: "Resealable plastic bags", Items: []string{
						"1/2 Gallon Ziploc Bags", "Quart Ziploc Bags", "Snack Size Ziploc Bags",
					}},
				},
			},
			{Name: "Appliances", Description: "Kitchen appliances", Items: []string{
				"Blender", "Coffee Maker", "Toaster", "Food Processor", "Stand Mixer",
			}},
		},
	},
	{
		Name: "Electronics", Description: "Electronic devices and gadgets",
		Children: []Node{
			{
				Name: "Computers", Description: "Computing devices",
				Children: []Node{
					{Name: "Laptops", Description: "Portable computers", Items: []string{
						"MacBook Pro 14\"", "Dell XPS 15", "ThinkPad X1 Carbon",
					}},
					{Name: "Accessories", Description: "Computer accessories", Items: []string{
						"Wireless Mouse", "Mechanical Keyboard", "USB-C Hub", "External SSD",
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
					{Name: "Hand Tools", Description: "Manual tools", Items: []string{
						"Hammer", "Screwdriver Set", "Wrench Set", "Pliers", "Tape Measure",
					}},
					{Name: "Power Tools", Description: "Electric and battery-powered tools", Items: []string{
						"Drill", "Circular Saw", "Jigsaw", "Sander", "Impact Driver",
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
					{Name: "Spices", Items: []string{"Cumin", "Turmeric"}},
				},
			},
		},
	},
	{
		Name: "Garage", Description: "Garage items",
		Children: []Node{
			{Name: "Tools", Items: []string{"Hammer"}},
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
	for _, itemName := range node.Items {
		if _, err := repo.CreateItem(ctx, itemName, itemType.ID); err != nil {
			return fmt.Errorf("failed to create item %s: %w", itemName, err)
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
