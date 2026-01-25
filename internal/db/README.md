# Database Layer Documentation

## Overview

The database layer uses SQLite with automatic migrations (Goose) and type-safe query generation (SQLC). All database code is generated from SQL definitions in `queries/` and migrations run automatically on application startup.

## Schema

### item_types

Hierarchical category tree with self-referencing parent_id foreign key.

```sql
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
```

**Constraints:**
- Root nodes have `parent_id = NULL`
- Unique name constraint scoped per parent (allows same name at different hierarchy levels)
- Separate unique index for root-level names
- CASCADE delete prevented to avoid accidental data loss

**Indexes:**
- `idx_item_types_parent`: Fast lookup of children
- `idx_item_types_depth`: Efficient depth-based queries
- `idx_item_types_root_name`: Enforce unique root names

### items

Physical inventory objects, restricted to leaf nodes only.

```sql
CREATE TABLE items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    item_type_id INTEGER NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (item_type_id) REFERENCES item_types(id) ON DELETE RESTRICT
);
```

**Indexes:**
- `idx_items_type`: Fast retrieval of all items for a given type

## Leaf-Node Constraint

The core rule—**items can only exist at leaf nodes**—is enforced via SQLite triggers.

### Trigger: enforce_leaf_node_on_insert

Prevents inserting items into non-leaf types (types that have children).

```sql
CREATE TRIGGER enforce_leaf_node_on_insert
BEFORE INSERT ON items
BEGIN
    SELECT RAISE(ABORT, 'Items can only be assigned to leaf nodes')
    WHERE EXISTS (SELECT 1 FROM item_types WHERE parent_id = NEW.item_type_id);
END;
```

### Trigger: enforce_leaf_node_on_update

Prevents moving items to non-leaf types.

```sql
CREATE TRIGGER enforce_leaf_node_on_update
BEFORE UPDATE OF item_type_id ON items
BEGIN
    SELECT RAISE(ABORT, 'Items can only be assigned to leaf nodes')
    WHERE EXISTS (SELECT 1 FROM item_types WHERE parent_id = NEW.item_type_id);
END;
```

### Trigger: prevent_parent_with_items

Prevents creating subcategories under types that already contain items.

```sql
CREATE TRIGGER prevent_parent_with_items
BEFORE INSERT ON item_types
WHEN NEW.parent_id IS NOT NULL
BEGIN
    SELECT RAISE(ABORT, 'Cannot create subcategory under a category that has items')
    WHERE EXISTS (SELECT 1 FROM items WHERE item_type_id = NEW.parent_id);
END;
```

## Recursive Queries

Hierarchy traversal uses PostgreSQL-compatible recursive CTEs.

### Ancestor Path (Bottom-up)

Get all ancestors from a node to root, ordered by depth (root first):

```sql
WITH RECURSIVE ancestors AS (
    SELECT id, parent_id, name, description, depth, created_at
    FROM item_types
    WHERE id = ?
    UNION ALL
    SELECT it.id, it.parent_id, it.name, it.description, it.depth, it.created_at
    FROM item_types it, ancestors a
    WHERE it.id = a.parent_id
)
SELECT * FROM ancestors ORDER BY depth;
```

**Used for:** Breadcrumb navigation in UI (`GetItemTypePath` query).

### Descendant Tree (Top-down)

Get all descendants from a node downward:

```sql
WITH RECURSIVE descendants AS (
    SELECT id, parent_id, name, description, depth, created_at
    FROM item_types
    WHERE id = ?
    UNION ALL
    SELECT it.id, it.parent_id, it.name, it.description, it.depth, it.created_at
    FROM item_types it, descendants d
    WHERE it.parent_id = d.id
)
SELECT * FROM descendants ORDER BY depth, name;
```

**Used for:** Bulk operations on subtrees (`GetItemTypeDescendants` query).

### Leaf Detection

Check if a type is a leaf (has no children):

```sql
SELECT CASE
    WHEN COUNT(*) = 0 THEN 1
    ELSE 0
END AS is_leaf
FROM item_types
WHERE parent_id = ?;
```

**Used for:** Determining whether to show items or subcategories in UI.

## SQLC Integration

All queries are defined in `queries/*.sql` with SQLC annotations:

```sql
-- name: GetItemType :one
SELECT id, parent_id, name, description, depth, created_at
FROM item_types
WHERE id = ?;

-- name: CreateRootItemType :one
INSERT INTO item_types (parent_id, name, description, depth)
VALUES (NULL, ?, ?, 0)
RETURNING id, parent_id, name, description, depth, created_at;
```

**Generate Go code:**
```bash
make generate  # Runs sqlc generate
```

Generated code appears in `internal/db/sqlc/` with type-safe parameter binding and struct mapping to `domain.ItemType` and `domain.Item`.

## Migrations

Migrations use Goose with embedded files via `//go:embed`:

```go
//go:embed migrations/*.sql
var embedMigrations embed.FS
```

**Run automatically on startup:**
```go
db.RunMigrations(database)  // cmd/inventory/main.go:37
```

**Migration files:**
- `migrations/00001_init_schema.sql` - Initial schema with triggers
- Future migrations: `00002_*.sql`, `00003_*.sql`, etc.

**Rollback (manual):**
```bash
goose -dir internal/db/migrations sqlite3 ./inventory.db down
```

## Connection Management

Database opened via `db.Open(cfg)` with pragmas for better SQLite performance:

```go
PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;
```

Connection closed on application exit via `defer database.Close()`.
