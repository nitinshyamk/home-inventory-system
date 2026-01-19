-- +goose Up
CREATE TABLE item_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_id INTEGER,
    name TEXT NOT NULL,
    description TEXT,
    depth INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (parent_id) REFERENCES item_types(id) ON DELETE RESTRICT,
    UNIQUE(parent_id, name)
);

CREATE INDEX idx_item_types_parent ON item_types(parent_id);
CREATE INDEX idx_item_types_depth ON item_types(depth);
CREATE UNIQUE INDEX idx_item_types_root_name ON item_types(name) WHERE parent_id IS NULL;

CREATE TABLE items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    item_type_id INTEGER NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (item_type_id) REFERENCES item_types(id) ON DELETE RESTRICT
);

CREATE INDEX idx_items_type ON items(item_type_id);

-- +goose StatementBegin
CREATE TRIGGER enforce_leaf_node_on_insert
BEFORE INSERT ON items
BEGIN
    SELECT RAISE(ABORT, 'Items can only be assigned to leaf nodes')
    WHERE EXISTS (SELECT 1 FROM item_types WHERE parent_id = NEW.item_type_id);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER enforce_leaf_node_on_update
BEFORE UPDATE OF item_type_id ON items
BEGIN
    SELECT RAISE(ABORT, 'Items can only be assigned to leaf nodes')
    WHERE EXISTS (SELECT 1 FROM item_types WHERE parent_id = NEW.item_type_id);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER prevent_parent_with_items
BEFORE INSERT ON item_types
WHEN NEW.parent_id IS NOT NULL
BEGIN
    SELECT RAISE(ABORT, 'Cannot create subcategory under a category that has items')
    WHERE EXISTS (SELECT 1 FROM items WHERE item_type_id = NEW.parent_id);
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS prevent_parent_with_items;
DROP TRIGGER IF EXISTS enforce_leaf_node_on_update;
DROP TRIGGER IF EXISTS enforce_leaf_node_on_insert;
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS item_types;
