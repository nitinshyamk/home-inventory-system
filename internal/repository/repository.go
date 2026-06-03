package repository

import (
	"context"
	"database/sql"
	"fmt"

	"home-inventory-system/internal/db/sqlc"
	"home-inventory-system/internal/domain"
)

// Repository provides data access operations
type Repository struct {
	queries *sqlc.Queries
	db      *sql.DB
}

// New creates a new Repository
func New(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		db:      db,
	}
}

// DB returns the underlying database connection
func (r *Repository) DB() *sql.DB {
	return r.db
}

// ListItemTypes returns all item types
func (r *Repository) ListItemTypes(ctx context.Context) ([]domain.ItemType, error) {
	rows, err := r.queries.ListItemTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list item types: %w", err)
	}

	return convertItemTypes(rows), nil
}

// GetItem returns a single item by ID
func (r *Repository) GetItem(ctx context.Context, id int64) (*domain.Item, error) {
	row, err := r.queries.GetItem(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("item not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	item := convertGetItemRow(row)
	return &item, nil
}

// GetItemType returns a single item type by ID
func (r *Repository) GetItemType(ctx context.Context, id int64) (*domain.ItemType, error) {
	row, err := r.queries.GetItemType(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("item type not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get item type: %w", err)
	}

	itemType := convertItemType(row)
	return &itemType, nil
}

// GetRootTypes returns all root-level item types (no parent)
func (r *Repository) GetRootTypes(ctx context.Context) ([]domain.ItemType, error) {
	rows, err := r.queries.ListRootItemTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list root types: %w", err)
	}

	return convertItemTypes(rows), nil
}

// GetChildTypes returns all direct children of a parent type
func (r *Repository) GetChildTypes(ctx context.Context, parentID int64) ([]domain.ItemType, error) {
	rows, err := r.queries.ListChildItemTypes(ctx, sql.NullInt64{Int64: parentID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list child types: %w", err)
	}

	return convertItemTypes(rows), nil
}

// GetTypePath returns the breadcrumb path from root to the given type
func (r *Repository) GetTypePath(ctx context.Context, typeID int64) ([]domain.ItemType, error) {
	rows, err := r.queries.GetItemTypePath(ctx, typeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get type path: %w", err)
	}

	types := make([]domain.ItemType, len(rows))
	for i, row := range rows {
		var parentID *int64
		if row.ParentID.Valid {
			parentID = &row.ParentID.Int64
		}
		var desc string
		if row.Description.Valid {
			desc = row.Description.String
		}
		types[i] = domain.ItemType{
			ID:          row.ID,
			ParentID:    parentID,
			Name:        row.Name,
			Description: desc,
			Depth:       row.Depth,
			CreatedAt:   row.CreatedAt,
		}
	}

	return types, nil
}

// IsLeafType checks if a type contains items (making it a leaf node)
func (r *Repository) IsLeafType(ctx context.Context, typeID int64) (bool, error) {
	result, err := r.queries.IsLeafItemType(ctx, typeID)
	if err != nil {
		return false, fmt.Errorf("failed to check leaf status: %w", err)
	}

	// The result is an int64 (1 for true, 0 for false)
	return result == 1, nil
}

// CountChildTypes returns the number of direct children of a type
func (r *Repository) CountChildTypes(ctx context.Context, parentID int64) (int64, error) {
	count, err := r.queries.CountChildItemTypes(ctx, sql.NullInt64{Int64: parentID, Valid: true})
	if err != nil {
		return 0, fmt.Errorf("failed to count child types: %w", err)
	}

	return count, nil
}

// CreateRootType creates a new root-level item type
func (r *Repository) CreateRootType(ctx context.Context, name, description string) (*domain.ItemType, error) {
	row, err := r.queries.CreateRootItemType(ctx, sqlc.CreateRootItemTypeParams{
		Name:        name,
		Description: toNullString(description),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create root type: %w", err)
	}

	itemType := convertItemType(row)
	return &itemType, nil
}

// CreateChildType creates a new child item type under a parent
func (r *Repository) CreateChildType(ctx context.Context, parentID int64, name, description string) (*domain.ItemType, error) {
	// Get parent to determine depth
	parent, err := r.queries.GetItemType(ctx, parentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("parent type not found: %d", parentID)
		}
		return nil, fmt.Errorf("failed to get parent type: %w", err)
	}

	row, err := r.queries.CreateItemType(ctx, sqlc.CreateItemTypeParams{
		ParentID:    sql.NullInt64{Int64: parentID, Valid: true},
		Name:        name,
		Description: toNullString(description),
		Depth:       parent.Depth + 1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create child type: %w", err)
	}

	itemType := convertItemType(row)
	return &itemType, nil
}

// ListLeafTypes returns all types that have no children
func (r *Repository) ListLeafTypes(ctx context.Context) ([]domain.ItemType, error) {
	rows, err := r.queries.ListLeafItemTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list leaf types: %w", err)
	}

	return convertItemTypes(rows), nil
}

// ListItemsByType returns all items of a specific type
func (r *Repository) ListItemsByType(ctx context.Context, typeID int64) ([]domain.Item, error) {
	rows, err := r.queries.ListItemsByType(ctx, typeID)
	if err != nil {
		return nil, fmt.Errorf("failed to list items by type: %w", err)
	}

	items := make([]domain.Item, len(rows))
	for i, row := range rows {
		items[i] = convertListItemsByTypeRow(row)
	}

	return items, nil
}

// CountItemsByType returns the number of items of a specific type
func (r *Repository) CountItemsByType(ctx context.Context, typeID int64) (int64, error) {
	count, err := r.queries.CountItemsByType(ctx, typeID)
	if err != nil {
		return 0, fmt.Errorf("failed to count items by type: %w", err)
	}

	return count, nil
}

// CreateItem creates a new item
func (r *Repository) CreateItem(ctx context.Context, name string, typeID int64, quantity float64, unitType domain.UnitType) (*domain.Item, error) {
	row, err := r.queries.CreateItem(ctx, sqlc.CreateItemParams{
		Name:       name,
		ItemTypeID: typeID,
		Quantity:   quantity,
		UnitType:   unitType.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	item := convertCreateItemRow(row)
	return &item, nil
}

// UpdateItemType updates the name and description of an item type
func (r *Repository) UpdateItemType(ctx context.Context, id int64, name, description string) error {
	err := r.queries.UpdateItemType(ctx, sqlc.UpdateItemTypeParams{
		ID:          id,
		Name:        name,
		Description: toNullString(description),
	})
	if err != nil {
		return fmt.Errorf("failed to update item type: %w", err)
	}
	return nil
}

// DeleteItem deletes an item by ID
func (r *Repository) DeleteItem(ctx context.Context, id int64) error {
	err := r.queries.DeleteItem(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}
	return nil
}

// DeleteItemType deletes an item type by ID
func (r *Repository) DeleteItemType(ctx context.Context, id int64) error {
	err := r.queries.DeleteItemType(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete item type: %w", err)
	}
	return nil
}

// GetTypeDescendants returns all descendants of a type (including the type itself)
func (r *Repository) GetTypeDescendants(ctx context.Context, typeID int64) ([]domain.ItemType, error) {
	rows, err := r.queries.GetItemTypeDescendants(ctx, typeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get type descendants: %w", err)
	}

	types := make([]domain.ItemType, len(rows))
	for i, row := range rows {
		var parentID *int64
		if row.ParentID.Valid {
			parentID = &row.ParentID.Int64
		}
		var desc string
		if row.Description.Valid {
			desc = row.Description.String
		}
		types[i] = domain.ItemType{
			ID:          row.ID,
			ParentID:    parentID,
			Name:        row.Name,
			Description: desc,
			Depth:       row.Depth,
			CreatedAt:   row.CreatedAt,
		}
	}

	return types, nil
}

// convertItemType converts a sqlc.ItemType to a domain.ItemType
func convertItemType(row sqlc.ItemType) domain.ItemType {
	var parentID *int64
	if row.ParentID.Valid {
		parentID = &row.ParentID.Int64
	}
	var desc string
	if row.Description.Valid {
		desc = row.Description.String
	}
	return domain.ItemType{
		ID:          row.ID,
		ParentID:    parentID,
		Name:        row.Name,
		Description: desc,
		Depth:       row.Depth,
		CreatedAt:   row.CreatedAt,
	}
}

// convertItemTypes converts a slice of sqlc.ItemType to domain.ItemType
func convertItemTypes(rows []sqlc.ItemType) []domain.ItemType {
	types := make([]domain.ItemType, len(rows))
	for i, row := range rows {
		types[i] = convertItemType(row)
	}
	return types
}

// convertGetItemRow converts a GetItemRow to a domain.Item
func convertGetItemRow(row sqlc.GetItemRow) domain.Item {
	return domain.Item{
		ID:         row.ID,
		Name:       row.Name,
		ItemTypeID: row.ItemTypeID,
		Quantity:   row.Quantity,
		UnitType:   domain.UnitType(row.UnitType),
		CreatedAt:  row.CreatedAt,
	}
}

// convertCreateItemRow converts a CreateItemRow to a domain.Item
func convertCreateItemRow(row sqlc.CreateItemRow) domain.Item {
	return domain.Item{
		ID:         row.ID,
		Name:       row.Name,
		ItemTypeID: row.ItemTypeID,
		Quantity:   row.Quantity,
		UnitType:   domain.UnitType(row.UnitType),
		CreatedAt:  row.CreatedAt,
	}
}

// convertListItemsByTypeRow converts a ListItemsByTypeRow to a domain.Item
func convertListItemsByTypeRow(row sqlc.ListItemsByTypeRow) domain.Item {
	return domain.Item{
		ID:         row.ID,
		Name:       row.Name,
		ItemTypeID: row.ItemTypeID,
		Quantity:   row.Quantity,
		UnitType:   domain.UnitType(row.UnitType),
		CreatedAt:  row.CreatedAt,
	}
}

// toNullString converts a string to sql.NullString
func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
