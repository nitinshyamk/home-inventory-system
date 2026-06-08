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

// CreateItem creates a new item with an optional location
func (r *Repository) CreateItem(ctx context.Context, name string, typeID int64, locationID *int64, quantity float64, unitType domain.UnitType) (*domain.Item, error) {
	row, err := r.queries.CreateItem(ctx, sqlc.CreateItemParams{
		Name:       name,
		ItemTypeID: typeID,
		LocationID: toNullInt64Ptr(locationID),
		Quantity:   quantity,
		UnitType:   unitType.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	item := convertCreateItemRow(row)
	return &item, nil
}

// LiftAndDeleteItemType deletes a category and lifts its children to the grandparent:
//   - Subcategories are re-parented to the grandparent (their depths are decremented recursively).
//   - Items are moved to the grandparent (allowed because the trigger was dropped in migration 00003).
//   - If the category is root-level (parentID IS NULL) and has items, returns an error.
//   - If the category's parent has other children (after re-parenting A's children), moving
//     items to the parent would violate the structural leaf constraint — the DB INSERT trigger
//     (enforce_leaf_node_on_insert) would catch that; this is surfaced as an error.
func (r *Repository) LiftAndDeleteItemType(ctx context.Context, id int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	q := sqlc.New(tx)

	// Load the category to find its parent
	cat, err := q.GetItemType(ctx, id)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	// Check for items in this category
	itemCount, err := q.CountItemsByType(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to count items: %w", err)
	}

	// Root + has items → reject
	if !cat.ParentID.Valid && itemCount > 0 {
		return fmt.Errorf("cannot delete root category that contains items: move or delete the items first")
	}

	// Re-parent child categories to grandparent
	children, err := q.ListChildItemTypes(ctx, sql.NullInt64{Int64: id, Valid: true})
	if err != nil {
		return fmt.Errorf("failed to list children: %w", err)
	}

	if len(children) > 0 {
		newParent := cat.ParentID
		newDepth := cat.Depth // children's new depth = category's current depth (one level up)
		for _, child := range children {
			if _, err := tx.ExecContext(ctx,
				`UPDATE item_types SET parent_id = ?, depth = ? WHERE id = ?`,
				toNullInt64(newParent), newDepth, child.ID,
			); err != nil {
				return fmt.Errorf("failed to re-parent child %d: %w", child.ID, err)
			}
			// Recursively decrement depths of all descendants
			descendants, err := r.getDescendantsInTx(ctx, tx, child.ID)
			if err != nil {
				return err
			}
			for _, desc := range descendants {
				if _, err := tx.ExecContext(ctx,
					`UPDATE item_types SET depth = depth - 1 WHERE id = ?`, desc,
				); err != nil {
					return fmt.Errorf("failed to update descendant depth %d: %w", desc, err)
				}
			}
		}
	}

	// Move items to grandparent (only valid when parentID is not null)
	if itemCount > 0 && cat.ParentID.Valid {
		if _, err := tx.ExecContext(ctx,
			`UPDATE items SET item_type_id = ? WHERE item_type_id = ?`,
			cat.ParentID.Int64, id,
		); err != nil {
			return fmt.Errorf("failed to lift items: %w", err)
		}
	}

	// Delete the category (now has no children and no items)
	if _, err := tx.ExecContext(ctx, `DELETE FROM item_types WHERE id = ?`, id); err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return tx.Commit()
}

// getDescendantsInTx returns IDs of all descendants of typeID within an existing transaction.
func (r *Repository) getDescendantsInTx(ctx context.Context, tx interface {
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
}, typeID int64) ([]int64, error) {
	const q = `
WITH RECURSIVE descendants AS (
    SELECT id FROM item_types WHERE id = ?
    UNION ALL
    SELECT it.id FROM item_types it JOIN descendants d ON it.parent_id = d.id
)
SELECT id FROM descendants WHERE id != ?`

	rows, err := tx.QueryContext(ctx, q, typeID, typeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query descendants: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func toNullInt64(n sql.NullInt64) interface{} {
	if n.Valid {
		return n.Int64
	}
	return nil
}

// UpdateItem updates an existing item's name, type, location, quantity, and unit type.
// Validates that the target type has no children (structural leaf check) since
// the enforce_leaf_node_on_update trigger was dropped in migration 00003.
func (r *Repository) UpdateItem(ctx context.Context, id int64, name string, typeID int64, locationID *int64, quantity float64, unitType domain.UnitType) error {
	childCount, err := r.queries.CountChildItemTypes(ctx, sql.NullInt64{Int64: typeID, Valid: true})
	if err != nil {
		return fmt.Errorf("failed to check leaf status: %w", err)
	}
	if childCount > 0 {
		return fmt.Errorf("items can only be assigned to leaf nodes")
	}
	err = r.queries.UpdateItem(ctx, sqlc.UpdateItemParams{
		ID:         id,
		Name:       name,
		ItemTypeID: typeID,
		LocationID: toNullInt64Ptr(locationID),
		Quantity:   quantity,
		UnitType:   unitType.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}
	return nil
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
		LocationID: toOptionalInt64(row.LocationID),
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
		LocationID: toOptionalInt64(row.LocationID),
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
		LocationID: toOptionalInt64(row.LocationID),
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

// toNullInt64Ptr converts a *int64 to sql.NullInt64
func toNullInt64Ptr(p *int64) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *p, Valid: true}
}

// toOptionalInt64 converts a sql.NullInt64 to *int64
func toOptionalInt64(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	v := n.Int64
	return &v
}

// convertLocation converts a sqlc.Location to a domain.Location
func convertLocation(row sqlc.Location) domain.Location {
	var parentID *int64
	if row.ParentID.Valid {
		parentID = &row.ParentID.Int64
	}
	var desc string
	if row.Description.Valid {
		desc = row.Description.String
	}
	return domain.Location{
		ID:          row.ID,
		ParentID:    parentID,
		Name:        row.Name,
		Description: desc,
		Depth:       row.Depth,
		CreatedAt:   row.CreatedAt,
	}
}

// convertLocations converts a slice of sqlc.Location to []domain.Location
func convertLocations(rows []sqlc.Location) []domain.Location {
	locs := make([]domain.Location, len(rows))
	for i, row := range rows {
		locs[i] = convertLocation(row)
	}
	return locs
}

// GetLocation returns a single location by ID
func (r *Repository) GetLocation(ctx context.Context, id int64) (*domain.Location, error) {
	row, err := r.queries.GetLocation(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("location not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get location: %w", err)
	}
	loc := convertLocation(row)
	return &loc, nil
}

// GetRootLocations returns all root-level locations (no parent)
func (r *Repository) GetRootLocations(ctx context.Context) ([]domain.Location, error) {
	rows, err := r.queries.ListRootLocations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list root locations: %w", err)
	}
	return convertLocations(rows), nil
}

// GetChildLocations returns all direct children of a parent location
func (r *Repository) GetChildLocations(ctx context.Context, parentID int64) ([]domain.Location, error) {
	rows, err := r.queries.ListChildLocations(ctx, sql.NullInt64{Int64: parentID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list child locations: %w", err)
	}
	return convertLocations(rows), nil
}

// GetLocationPath returns the breadcrumb path from root to the given location
func (r *Repository) GetLocationPath(ctx context.Context, locationID int64) ([]domain.Location, error) {
	rows, err := r.queries.GetLocationPath(ctx, locationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get location path: %w", err)
	}
	locs := make([]domain.Location, len(rows))
	for i, row := range rows {
		var parentID *int64
		if row.ParentID.Valid {
			parentID = &row.ParentID.Int64
		}
		var desc string
		if row.Description.Valid {
			desc = row.Description.String
		}
		locs[i] = domain.Location{
			ID:          row.ID,
			ParentID:    parentID,
			Name:        row.Name,
			Description: desc,
			Depth:       row.Depth,
			CreatedAt:   row.CreatedAt,
		}
	}
	return locs, nil
}

// IsLeafLocation checks if a location contains items (making it a leaf node)
func (r *Repository) IsLeafLocation(ctx context.Context, locationID int64) (bool, error) {
	result, err := r.queries.IsLeafLocation(ctx, sql.NullInt64{Int64: locationID, Valid: true})
	if err != nil {
		return false, fmt.Errorf("failed to check location leaf status: %w", err)
	}
	return result == 1, nil
}

// ListLeafLocations returns all locations that have no children
func (r *Repository) ListLeafLocations(ctx context.Context) ([]domain.Location, error) {
	rows, err := r.queries.ListLeafLocations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list leaf locations: %w", err)
	}
	return convertLocations(rows), nil
}

// CountChildLocations returns the number of direct children of a location
func (r *Repository) CountChildLocations(ctx context.Context, parentID int64) (int64, error) {
	count, err := r.queries.CountChildLocations(ctx, sql.NullInt64{Int64: parentID, Valid: true})
	if err != nil {
		return 0, fmt.Errorf("failed to count child locations: %w", err)
	}
	return count, nil
}

// GetLocationDescendants returns all descendants of a location (including itself)
func (r *Repository) GetLocationDescendants(ctx context.Context, locationID int64) ([]domain.Location, error) {
	rows, err := r.queries.GetLocationDescendants(ctx, locationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get location descendants: %w", err)
	}
	locs := make([]domain.Location, len(rows))
	for i, row := range rows {
		var parentID *int64
		if row.ParentID.Valid {
			parentID = &row.ParentID.Int64
		}
		var desc string
		if row.Description.Valid {
			desc = row.Description.String
		}
		locs[i] = domain.Location{
			ID:          row.ID,
			ParentID:    parentID,
			Name:        row.Name,
			Description: desc,
			Depth:       row.Depth,
			CreatedAt:   row.CreatedAt,
		}
	}
	return locs, nil
}

// CreateRootLocation creates a new root-level location
func (r *Repository) CreateRootLocation(ctx context.Context, name, description string) (*domain.Location, error) {
	row, err := r.queries.CreateRootLocation(ctx, sqlc.CreateRootLocationParams{
		Name:        name,
		Description: toNullString(description),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create root location: %w", err)
	}
	loc := convertLocation(row)
	return &loc, nil
}

// CreateChildLocation creates a new child location under a parent
func (r *Repository) CreateChildLocation(ctx context.Context, parentID int64, name, description string) (*domain.Location, error) {
	parent, err := r.queries.GetLocation(ctx, parentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("parent location not found: %d", parentID)
		}
		return nil, fmt.Errorf("failed to get parent location: %w", err)
	}
	row, err := r.queries.CreateLocation(ctx, sqlc.CreateLocationParams{
		ParentID:    sql.NullInt64{Int64: parentID, Valid: true},
		Name:        name,
		Description: toNullString(description),
		Depth:       parent.Depth + 1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create child location: %w", err)
	}
	loc := convertLocation(row)
	return &loc, nil
}

// DeleteLocation deletes a location by ID
func (r *Repository) DeleteLocation(ctx context.Context, id int64) error {
	err := r.queries.DeleteLocation(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete location: %w", err)
	}
	return nil
}

// ListItemsByLocation returns all items assigned to a specific location
func (r *Repository) ListItemsByLocation(ctx context.Context, locationID int64) ([]domain.Item, error) {
	rows, err := r.queries.ListItemsByLocation(ctx, sql.NullInt64{Int64: locationID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list items by location: %w", err)
	}
	items := make([]domain.Item, len(rows))
	for i, row := range rows {
		items[i] = domain.Item{
			ID:         row.ID,
			Name:       row.Name,
			ItemTypeID: row.ItemTypeID,
			LocationID: toOptionalInt64(row.LocationID),
			Quantity:   row.Quantity,
			UnitType:   domain.UnitType(row.UnitType),
			CreatedAt:  row.CreatedAt,
		}
	}
	return items, nil
}

// CountItemsByLocation returns the number of items assigned to a specific location
func (r *Repository) CountItemsByLocation(ctx context.Context, locationID int64) (int64, error) {
	count, err := r.queries.CountItemsByLocation(ctx, sql.NullInt64{Int64: locationID, Valid: true})
	if err != nil {
		return 0, fmt.Errorf("failed to count items by location: %w", err)
	}
	return count, nil
}
