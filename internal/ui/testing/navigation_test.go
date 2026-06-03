package testing

import (
	"context"
	"strings"
	"testing"

	"home-inventory-system/internal/service"
	"home-inventory-system/internal/ui"
)

func TestBasicNavigation(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	// Seed some test data
	ctx := context.Background()
	sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{
		Name:        "Kitchen",
		Description: "Kitchen items",
	})
	sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{
		Name:        "Garage",
		Description: "Tools and equipment",
	})

	// Reload the UI to show seeded data
	sim.Reload()

	// Verify initial state
	sim.AssertState(t, ui.StateBrowsingTypes)
	sim.AssertListContains(t, "Kitchen")
	sim.AssertListContains(t, "Garage")
	sim.AssertNoError(t)
}

func TestNavigateDown(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	// Seed test data
	ctx := context.Background()
	kitchenResult := sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{
		Name: "Kitchen",
	}).(service.CreateRootTypeResult)

	sim.handler.HandleCommand(ctx, service.CreateChildTypeCommand{
		ParentID: kitchenResult.Type.ID,
		Name:     "Pantry",
	})

	// Reload
	sim.Reload()

	// Navigate to Kitchen
	sim.SendKeys(KeyEnter)

	// Should now be viewing Kitchen's children
	sim.AssertState(t, ui.StateBrowsingTypes)
	sim.AssertListContains(t, "Pantry")
	sim.AssertBreadcrumbContains(t, "Kitchen")
}

func TestRightPaneUpdatesOnNavigation(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	ctx := context.Background()
	sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{Name: "Appliances"})
	sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{Name: "Kitchen"})

	sim.Reload()

	// The right pane should show whichever category is highlighted.
	// After reload the list is at the first item (alphabetical order: Appliances).
	sim.AssertListContains(t, "Appliances")

	// Navigate down to highlight "Kitchen"
	sim.SendKeys(KeyDown)

	// Right pane should now show "Kitchen"
	view := sim.View()
	if !strings.Contains(view, "Kitchen") {
		t.Errorf("Expected right pane to contain 'Kitchen' after navigating to it, got:\n%s", view)
	}
}

func TestNavigateBack(t *testing.T) {
	sim := NewSimulator(t)
	defer sim.Cleanup()

	// Seed test data
	ctx := context.Background()
	kitchenResult := sim.handler.HandleCommand(ctx, service.CreateRootTypeCommand{
		Name: "Kitchen",
	}).(service.CreateRootTypeResult)

	sim.handler.HandleCommand(ctx, service.CreateChildTypeCommand{
		ParentID: kitchenResult.Type.ID,
		Name:     "Pantry",
	})

	// Reload and navigate to Kitchen
	sim.Reload()
	sim.SendKeys(KeyEnter)

	// Verify we're in Kitchen
	sim.AssertListContains(t, "Pantry")

	// Navigate back
	sim.SendKeys(KeyEsc)

	// Should be back at root
	sim.AssertState(t, ui.StateBrowsingTypes)
	sim.AssertListContains(t, "Kitchen")
	sim.AssertListNotContains(t, "Pantry")
}
