-- +goose Up
-- The enforce_leaf_node_on_update trigger prevents moving items to a category
-- that still has sub-categories. The lift-and-delete operation needs to move
-- items from a category to its parent atomically. Because the ordering
-- constraint (delete child before moving items) conflicts with FK RESTRICT,
-- we enforce the leaf-node rule at the application layer instead.
-- The INSERT trigger (enforce_leaf_node_on_insert) is preserved.
DROP TRIGGER IF EXISTS enforce_leaf_node_on_update;

-- +goose Down
CREATE TRIGGER enforce_leaf_node_on_update
BEFORE UPDATE OF item_type_id ON items
BEGIN
    SELECT RAISE(ABORT, 'Items can only be assigned to leaf nodes')
    WHERE EXISTS (SELECT 1 FROM item_types WHERE parent_id = NEW.item_type_id);
END;
