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

// loadCategoryDetail fetches category count data for the right pane detail view.
// Returns a RightPaneDetailMsg with IsLeaf, ChildCount, or ItemCount.
func loadCategoryDetail(handler *service.Handler, typeID int64) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		leafResult, ok := handler.HandleQuery(ctx, service.IsLeafTypeQuery{TypeID: typeID}).(service.IsLeafTypeResult)
		if !ok || leafResult.Err != nil {
			return messages.RightPaneDetailMsg{CategoryID: typeID, Err: leafResult.Err}
		}

		if leafResult.IsLeaf {
			countResult, ok := handler.HandleQuery(ctx, service.CountItemsByTypeQuery{TypeID: typeID}).(service.CountItemsByTypeResult)
			if !ok || countResult.Err != nil {
				return messages.RightPaneDetailMsg{CategoryID: typeID, Err: countResult.Err}
			}
			return messages.RightPaneDetailMsg{
				CategoryID: typeID,
				IsLeaf:     true,
				ItemCount:  countResult.Count,
			}
		}

		childResult, ok := handler.HandleQuery(ctx, service.ListChildTypesQuery{ParentID: typeID}).(service.ListChildTypesResult)
		if !ok || childResult.Err != nil {
			return messages.RightPaneDetailMsg{CategoryID: typeID, Err: childResult.Err}
		}
		return messages.RightPaneDetailMsg{
			CategoryID: typeID,
			IsLeaf:     false,
			ChildCount: int64(len(childResult.Types)),
		}
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
