-- name: GetLocation :one
SELECT id, parent_id, name, description, depth, created_at
FROM locations
WHERE id = ?;

-- name: ListLocations :many
SELECT id, parent_id, name, description, depth, created_at
FROM locations
ORDER BY depth, name;

-- name: ListRootLocations :many
SELECT id, parent_id, name, description, depth, created_at
FROM locations
WHERE parent_id IS NULL
ORDER BY name;

-- name: ListChildLocations :many
SELECT id, parent_id, name, description, depth, created_at
FROM locations
WHERE parent_id = ?
ORDER BY name;

-- name: ListLeafLocations :many
SELECT l.id, l.parent_id, l.name, l.description, l.depth, l.created_at
FROM locations l
WHERE NOT EXISTS (
    SELECT 1 FROM locations AS children WHERE children.parent_id = l.id
)
ORDER BY l.name;

-- name: GetLocationAncestors :many
WITH RECURSIVE ancestors AS (
    SELECT id, parent_id, name, description, depth, created_at
    FROM locations
    WHERE locations.id = sqlc.arg(location_id)
    UNION ALL
    SELECT l.id, l.parent_id, l.name, l.description, l.depth, l.created_at
    FROM locations l, ancestors a
    WHERE l.id = a.parent_id
)
SELECT * FROM ancestors
ORDER BY depth;

-- name: GetLocationPath :many
WITH RECURSIVE path AS (
    SELECT id, parent_id, name, description, depth, created_at
    FROM locations
    WHERE locations.id = sqlc.arg(location_id)
    UNION ALL
    SELECT l.id, l.parent_id, l.name, l.description, l.depth, l.created_at
    FROM locations l, path p
    WHERE l.id = p.parent_id
)
SELECT * FROM path
ORDER BY depth ASC;

-- name: IsLeafLocation :one
SELECT CASE
    WHEN COUNT(*) > 0 THEN 1
    ELSE 0
END AS is_leaf
FROM items
WHERE location_id = ?;

-- name: CountChildLocations :one
SELECT COUNT(*) AS count
FROM locations
WHERE parent_id = ?;

-- name: CreateLocation :one
INSERT INTO locations (parent_id, name, description, depth)
VALUES (?, ?, ?, ?)
RETURNING id, parent_id, name, description, depth, created_at;

-- name: CreateRootLocation :one
INSERT INTO locations (parent_id, name, description, depth)
VALUES (NULL, ?, ?, 0)
RETURNING id, parent_id, name, description, depth, created_at;

-- name: DeleteLocation :exec
DELETE FROM locations WHERE id = ?;

-- name: GetLocationDescendants :many
WITH RECURSIVE descendants AS (
    SELECT id, parent_id, name, description, depth, created_at
    FROM locations
    WHERE locations.id = sqlc.arg(location_id)
    UNION ALL
    SELECT l.id, l.parent_id, l.name, l.description, l.depth, l.created_at
    FROM locations l, descendants d
    WHERE l.parent_id = d.id
)
SELECT * FROM descendants
ORDER BY depth, name;

-- name: ListItemsByLocation :many
SELECT id, name, item_type_id, location_id, quantity, unit_type, created_at
FROM items
WHERE location_id = ?
ORDER BY name;

-- name: CountItemsByLocation :one
SELECT COUNT(*) AS count
FROM items
WHERE location_id = ?;
