-- +goose Up
CREATE TABLE locations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_id INTEGER,
    name TEXT NOT NULL,
    description TEXT,
    depth INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (parent_id) REFERENCES locations(id) ON DELETE RESTRICT,
    UNIQUE(parent_id, name)
);

CREATE INDEX idx_locations_parent ON locations(parent_id);
CREATE INDEX idx_locations_depth ON locations(depth);
CREATE UNIQUE INDEX idx_locations_root_name ON locations(name) WHERE parent_id IS NULL;

ALTER TABLE items ADD COLUMN location_id INTEGER REFERENCES locations(id) ON DELETE RESTRICT;

CREATE INDEX idx_items_location ON items(location_id);

-- +goose StatementBegin
CREATE TRIGGER enforce_leaf_location_on_insert
BEFORE INSERT ON items
WHEN NEW.location_id IS NOT NULL
BEGIN
    SELECT RAISE(ABORT, 'Items can only be assigned to leaf locations')
    WHERE EXISTS (SELECT 1 FROM locations WHERE parent_id = NEW.location_id);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER enforce_leaf_location_on_update
BEFORE UPDATE OF location_id ON items
WHEN NEW.location_id IS NOT NULL
BEGIN
    SELECT RAISE(ABORT, 'Items can only be assigned to leaf locations')
    WHERE EXISTS (SELECT 1 FROM locations WHERE parent_id = NEW.location_id);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER prevent_parent_location_with_items
BEFORE INSERT ON locations
WHEN NEW.parent_id IS NOT NULL
BEGIN
    SELECT RAISE(ABORT, 'Cannot create sublocation under a location that has items')
    WHERE EXISTS (SELECT 1 FROM items WHERE location_id = NEW.parent_id);
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS prevent_parent_location_with_items;
DROP TRIGGER IF EXISTS enforce_leaf_location_on_update;
DROP TRIGGER IF EXISTS enforce_leaf_location_on_insert;
DROP INDEX IF EXISTS idx_items_location;
ALTER TABLE items DROP COLUMN location_id;
DROP TABLE IF EXISTS locations;
