// Package seed provides functions to populate the database with sample data.
// It uses the repository layer to respect business logic and constraints.
package seed

import (
	"context"
	"database/sql"
	"fmt"

	"home-inventory-system/internal/repository"
)

// Seed populates the database with comprehensive sample hierarchical data.
// This creates a realistic home inventory structure for testing and development.
func Seed(ctx context.Context, repo *repository.Repository) error {
	// Kitchen hierarchy
	kitchen, err := repo.CreateRootType(ctx, "Kitchen", sql.NullString{String: "Kitchen and cooking related items", Valid: true})
	if err != nil {
		return fmt.Errorf("failed to create Kitchen: %w", err)
	}

	pantryItems, err := repo.CreateChildType(ctx, kitchen.ID, "Pantry Items", sql.NullString{String: "Food storage items", Valid: true})
	if err != nil {
		return fmt.Errorf("failed to create Pantry Items: %w", err)
	}

	spices, err := repo.CreateChildType(ctx, pantryItems.ID, "Spices", sql.NullString{String: "Cooking spices and seasonings", Valid: true})
	if err != nil {
		return fmt.Errorf("failed to create Spices: %w", err)
	}

	// Add spice items
	spiceItems := []string{"Black Cardamom", "Cinnamon Sticks", "Cumin Seeds", "Turmeric Powder", "Red Chili Flakes"}
	for _, name := range spiceItems {
		if _, err := repo.CreateItem(ctx, name, spices.ID); err != nil {
			return fmt.Errorf("failed to create spice %s: %w", name, err)
		}
	}

	grains, err := repo.CreateChildType(ctx, pantryItems.ID, "Grains", sql.NullString{String: "Rice and grains", Valid: true})
	if err != nil {
		return fmt.Errorf("failed to create Grains: %w", err)
	}

	rice, err := repo.CreateChildType(ctx, grains.ID, "Rice", sql.NullString{String: "Various types of rice", Valid: true})
	if err != nil {
		return fmt.Errorf("failed to create Rice: %w", err)
	}

	// Add rice items
	riceItems := []string{"Jasmine Rice", "Basmati Rice", "Brown Rice", "Arborio Rice"}
	for _, name := range riceItems {
		if _, err := repo.CreateItem(ctx, name, rice.ID); err != nil {
			return fmt.Errorf("failed to create rice %s: %w", name, err)
		}
	}

	// Disposable containers
	disposable, err := repo.CreateChildType(ctx, kitchen.ID, "Disposable Containers", sql.NullString{String: "Single-use storage containers", Valid: true})
	if err != nil {
		return fmt.Errorf("failed to create Disposable Containers: %w", err)
	}

	ziploc, err := repo.CreateChildType(ctx, disposable.ID, "Ziploc Bags", sql.NullString{String: "Resealable plastic bags", Valid: true})
	if err != nil {
		return fmt.Errorf("failed to create Ziploc Bags: %w", err)
	}

	// Add ziploc items
	ziplocItems := []string{"1/2 Gallon Ziploc Bags", "Quart Ziploc Bags", "Snack Size Ziploc Bags"}
	for _, name := range ziplocItems {
		if _, err := repo.CreateItem(ctx, name, ziploc.ID); err != nil {
			return fmt.Errorf("failed to create ziploc %s: %w", name, err)
		}
	}

	// Kitchen appliances
	appliances, err := repo.CreateChildType(ctx, kitchen.ID, "Appliances", sql.NullString{String: "Kitchen appliances", Valid: true})
	if err != nil {
		return fmt.Errorf("failed to create Appliances: %w", err)
	}

	// Add appliance items
	applianceItems := []string{"Blender", "Coffee Maker", "Toaster", "Food Processor", "Stand Mixer"}
	for _, name := range applianceItems {
		if _, err := repo.CreateItem(ctx, name, appliances.ID); err != nil {
			return fmt.Errorf("failed to create appliance %s: %w", name, err)
		}
	}

	// Electronics hierarchy
	electronics, err := repo.CreateRootType(ctx, "Electronics", sql.NullString{String: "Electronic devices and gadgets", Valid: true})
	if err != nil {
		return fmt.Errorf("failed to create Electronics: %w", err)
	}

	computers, err := repo.CreateChildType(ctx, electronics.ID, "Computers", sql.NullString{String: "Computing devices", Valid: true})
	if err != nil {
		return fmt.Errorf("failed to create Computers: %w", err)
	}

	laptops, err := repo.CreateChildType(ctx, computers.ID, "Laptops", sql.NullString{String: "Portable computers", Valid: true})
	if err != nil {
		return fmt.Errorf("failed to create Laptops: %w", err)
	}

	// Add laptop items
	laptopItems := []string{"MacBook Pro 14\"", "Dell XPS 15", "ThinkPad X1 Carbon"}
	for _, name := range laptopItems {
		if _, err := repo.CreateItem(ctx, name, laptops.ID); err != nil {
			return fmt.Errorf("failed to create laptop %s: %w", name, err)
		}
	}

	accessories, err := repo.CreateChildType(ctx, computers.ID, "Accessories", sql.NullString{String: "Computer accessories", Valid: true})
	if err != nil {
		return fmt.Errorf("failed to create Accessories: %w", err)
	}

	// Add accessory items
	accessoryItems := []string{"Wireless Mouse", "Mechanical Keyboard", "USB-C Hub", "External SSD"}
	for _, name := range accessoryItems {
		if _, err := repo.CreateItem(ctx, name, accessories.ID); err != nil {
			return fmt.Errorf("failed to create accessory %s: %w", name, err)
		}
	}

	// Audio/Video
	audioVideo, err := repo.CreateChildType(ctx, electronics.ID, "Audio/Video", sql.NullString{String: "Audio and video equipment", Valid: true})
	if err != nil {
		return fmt.Errorf("failed to create Audio/Video: %w", err)
	}

	// Add AV items
	avItems := []string{"Sony TV 55\"", "Sonos Soundbar", "Apple TV 4K", "AirPods Pro"}
	for _, name := range avItems {
		if _, err := repo.CreateItem(ctx, name, audioVideo.ID); err != nil {
			return fmt.Errorf("failed to create AV item %s: %w", name, err)
		}
	}

	// Garage hierarchy
	garage, err := repo.CreateRootType(ctx, "Garage", sql.NullString{String: "Garage storage and tools", Valid: true})
	if err != nil {
		return fmt.Errorf("failed to create Garage: %w", err)
	}

	tools, err := repo.CreateChildType(ctx, garage.ID, "Tools", sql.NullString{String: "Hand and power tools", Valid: true})
	if err != nil {
		return fmt.Errorf("failed to create Tools: %w", err)
	}

	handTools, err := repo.CreateChildType(ctx, tools.ID, "Hand Tools", sql.NullString{String: "Manual tools", Valid: true})
	if err != nil {
		return fmt.Errorf("failed to create Hand Tools: %w", err)
	}

	// Add hand tool items
	handToolItems := []string{"Hammer", "Screwdriver Set", "Wrench Set", "Pliers", "Tape Measure"}
	for _, name := range handToolItems {
		if _, err := repo.CreateItem(ctx, name, handTools.ID); err != nil {
			return fmt.Errorf("failed to create hand tool %s: %w", name, err)
		}
	}

	powerTools, err := repo.CreateChildType(ctx, tools.ID, "Power Tools", sql.NullString{String: "Electric and battery-powered tools", Valid: true})
	if err != nil {
		return fmt.Errorf("failed to create Power Tools: %w", err)
	}

	// Add power tool items
	powerToolItems := []string{"Drill", "Circular Saw", "Jigsaw", "Sander", "Impact Driver"}
	for _, name := range powerToolItems {
		if _, err := repo.CreateItem(ctx, name, powerTools.ID); err != nil {
			return fmt.Errorf("failed to create power tool %s: %w", name, err)
		}
	}

	return nil
}

// SeedMinimal creates minimal test data (fewer records) for quick testing.
func SeedMinimal(ctx context.Context, repo *repository.Repository) error {
	// Simple hierarchy: Kitchen > Pantry > Spices (with 2 items)
	kitchen, err := repo.CreateRootType(ctx, "Kitchen", sql.NullString{String: "Kitchen items", Valid: true})
	if err != nil {
		return fmt.Errorf("failed to create Kitchen: %w", err)
	}

	pantry, err := repo.CreateChildType(ctx, kitchen.ID, "Pantry", sql.NullString{})
	if err != nil {
		return fmt.Errorf("failed to create Pantry: %w", err)
	}

	spices, err := repo.CreateChildType(ctx, pantry.ID, "Spices", sql.NullString{})
	if err != nil {
		return fmt.Errorf("failed to create Spices: %w", err)
	}

	if _, err := repo.CreateItem(ctx, "Cumin", spices.ID); err != nil {
		return fmt.Errorf("failed to create Cumin: %w", err)
	}
	if _, err := repo.CreateItem(ctx, "Turmeric", spices.ID); err != nil {
		return fmt.Errorf("failed to create Turmeric: %w", err)
	}

	// Add another root: Garage > Tools (leaf with 1 item)
	garage, err := repo.CreateRootType(ctx, "Garage", sql.NullString{String: "Garage items", Valid: true})
	if err != nil {
		return fmt.Errorf("failed to create Garage: %w", err)
	}

	tools, err := repo.CreateChildType(ctx, garage.ID, "Tools", sql.NullString{})
	if err != nil {
		return fmt.Errorf("failed to create Tools: %w", err)
	}

	if _, err := repo.CreateItem(ctx, "Hammer", tools.ID); err != nil {
		return fmt.Errorf("failed to create Hammer: %w", err)
	}

	return nil
}
