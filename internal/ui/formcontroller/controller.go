package formcontroller

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"home-inventory-system/internal/service"
	"home-inventory-system/internal/ui/components/categoryform"
	"home-inventory-system/internal/ui/components/itemform"
	"home-inventory-system/internal/ui/messages"
)

// Controller centralizes form lifecycle management
type Controller struct {
	handler *service.Handler
}

// New creates a new form controller
func New(handler *service.Handler) *Controller {
	return &Controller{
		handler: handler,
	}
}

// FormContext provides context for form operations
type FormContext struct {
	CurrentTypeID *int64
	Width         int
	Height        int
}

// OpenCategoryForm creates a new category form model
func (c *Controller) OpenCategoryForm(ctx FormContext) categoryform.Model {
	return categoryform.New(ctx.Width, ctx.Height)
}

// OpenItemForm creates a new item form model
func (c *Controller) OpenItemForm(ctx FormContext) itemform.Model {
	return itemform.New(ctx.Width, ctx.Height)
}

// SubmitCategoryForm handles category form submission and returns a command
func (c *Controller) SubmitCategoryForm(data categoryform.Data, formCtx FormContext) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		if formCtx.CurrentTypeID == nil {
			// Creating at root level
			result := c.handler.HandleCommand(ctx, service.CreateRootTypeCommand{
				Name:        data.Name,
				Description: data.Description,
			})
			if createResult, ok := result.(service.CreateRootTypeResult); ok {
				return messages.CategoryCreatedMsg{Category: createResult.Type, Err: createResult.Err}
			}
			return messages.CategoryCreatedMsg{Err: fmt.Errorf("unexpected result type")}
		}

		// Creating as child of current type
		result := c.handler.HandleCommand(ctx, service.CreateChildTypeCommand{
			ParentID:    *formCtx.CurrentTypeID,
			Name:        data.Name,
			Description: data.Description,
		})
		if createResult, ok := result.(service.CreateChildTypeResult); ok {
			return messages.CategoryCreatedMsg{Category: createResult.Type, Err: createResult.Err}
		}
		return messages.CategoryCreatedMsg{Err: fmt.Errorf("unexpected result type")}
	}
}

// SubmitItemForm handles item form submission and returns a command
func (c *Controller) SubmitItemForm(data itemform.Data, formCtx FormContext) tea.Cmd {
	return func() tea.Msg {
		// Items can only be created at leaf nodes (when CurrentTypeID is set)
		if formCtx.CurrentTypeID == nil {
			return messages.ItemCreatedMsg{Err: fmt.Errorf("cannot create items at root level")}
		}

		ctx := context.Background()
		result := c.handler.HandleCommand(ctx, service.CreateItemCommand{
			Name:     data.Name,
			TypeID:   *formCtx.CurrentTypeID,
			Quantity: data.Quantity,
			UnitType: data.UnitType,
		})

		if createResult, ok := result.(service.CreateItemResult); ok {
			return messages.ItemCreatedMsg{Item: createResult.Item, Err: createResult.Err}
		}
		return messages.ItemCreatedMsg{Err: fmt.Errorf("unexpected result type")}
	}
}
