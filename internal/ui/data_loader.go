package ui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"home-inventory-system/internal/service"
	"home-inventory-system/internal/ui/messages"
)

// Data loading operations that execute service queries and emit messages.

// loadRootTypes fetches root-level item types.
func loadRootTypes(handler *service.Handler) tea.Cmd {
	return func() tea.Msg {
		result := handler.HandleQuery(context.Background(), service.ListRootTypesQuery{})
		if typesResult, ok := result.(service.ListRootTypesResult); ok {
			return messages.TypesLoadedMsg{
				Types:    typesResult.Types,
				ParentID: nil,
				Err:      typesResult.Err,
			}
		}
		return messages.ErrorMsg{Err: fmt.Errorf("unexpected query result type")}
	}
}

// loadChildTypes fetches child types for a parent.
func loadChildTypes(handler *service.Handler, parentID int64) tea.Cmd {
	return func() tea.Msg {
		result := handler.HandleQuery(context.Background(), service.ListChildTypesQuery{ParentID: parentID})
		if typesResult, ok := result.(service.ListChildTypesResult); ok {
			return messages.TypesLoadedMsg{
				Types:    typesResult.Types,
				ParentID: &parentID,
				Err:      typesResult.Err,
			}
		}
		return messages.ErrorMsg{Err: fmt.Errorf("unexpected query result type")}
	}
}

// loadItemsForType fetches items at a leaf type.
func loadItemsForType(handler *service.Handler, typeID int64) tea.Cmd {
	return func() tea.Msg {
		result := handler.HandleQuery(context.Background(), service.ListItemsByTypeQuery{TypeID: typeID})
		if itemsResult, ok := result.(service.ListItemsByTypeResult); ok {
			return messages.LeafItemsLoadedMsg{
				TypeID: typeID,
				Items:  itemsResult.Items,
				Err:    itemsResult.Err,
			}
		}
		return messages.ErrorMsg{Err: fmt.Errorf("unexpected query result type")}
	}
}

// loadBreadcrumb fetches the path from root to current type.
func loadBreadcrumb(handler *service.Handler, typeID int64) tea.Cmd {
	return func() tea.Msg {
		result := handler.HandleQuery(context.Background(), service.GetTypePathQuery{TypeID: typeID})
		if pathResult, ok := result.(service.GetTypePathResult); ok {
			return messages.BreadcrumbLoadedMsg{
				Path: pathResult.Path,
				Err:  pathResult.Err,
			}
		}
		return messages.ErrorMsg{Err: fmt.Errorf("unexpected query result type")}
	}
}

// checkLeafType asynchronously checks if a type is a leaf (has no children).
// Emits LeafCheckCompleteMsg when done.
func checkLeafType(handler *service.Handler, typeID int64) tea.Cmd {
	return func() tea.Msg {
		result := handler.HandleQuery(context.Background(), service.IsLeafTypeQuery{TypeID: typeID})
		leafResult, ok := result.(service.IsLeafTypeResult)
		if !ok {
			return messages.ErrorMsg{Err: fmt.Errorf("failed to check leaf status: unexpected result type")}
		}

		return messages.LeafCheckCompleteMsg{
			TypeID: typeID,
			IsLeaf: leafResult.IsLeaf,
			Err:    leafResult.Err,
		}
	}
}
