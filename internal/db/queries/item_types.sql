-- name: GetItemType :one
SELECT id, parent_id, name, description, depth, created_at
FROM item_types
WHERE id = ?;

-- name: ListItemTypes :many
SELECT id, parent_id, name, description, depth, created_at
FROM item_types
ORDER BY depth, name;

-- name: ListRootItemTypes :many
SELECT id, parent_id, name, description, depth, created_at
FROM item_types
WHERE parent_id IS NULL
ORDER BY name;

-- name: ListChildItemTypes :many
SELECT id, parent_id, name, description, depth, created_at
FROM item_types
WHERE parent_id = ?
ORDER BY name;

-- name: ListLeafItemTypes :many
SELECT it.id, it.parent_id, it.name, it.description, it.depth, it.created_at
FROM item_types it
WHERE NOT EXISTS (
    SELECT 1 FROM item_types AS children WHERE children.parent_id = it.id
)
ORDER BY it.name;

-- name: GetItemTypeAncestors :many
WITH RECURSIVE ancestors AS (
    SELECT id, parent_id, name, description, depth, created_at
    FROM item_types
    WHERE item_types.id = sqlc.arg(type_id)
    UNION ALL
    SELECT it.id, it.parent_id, it.name, it.description, it.depth, it.created_at
    FROM item_types it, ancestors a
    WHERE it.id = a.parent_id
)
SELECT * FROM ancestors
ORDER BY depth;

-- name: GetItemTypePath :many
WITH RECURSIVE path AS (
    SELECT id, parent_id, name, description, depth, created_at
    FROM item_types
    WHERE item_types.id = sqlc.arg(type_id)
    UNION ALL
    SELECT it.id, it.parent_id, it.name, it.description, it.depth, it.created_at
    FROM item_types it, path p
    WHERE it.id = p.parent_id
)
SELECT * FROM path
ORDER BY depth ASC;

-- name: IsLeafItemType :one
SELECT CASE
    WHEN COUNT(*) = 0 THEN 1
    ELSE 0
END AS is_leaf
FROM item_types
WHERE parent_id = ?;

-- name: CountChildItemTypes :one
SELECT COUNT(*) AS count
FROM item_types
WHERE parent_id = ?;

-- name: CreateItemType :one
INSERT INTO item_types (parent_id, name, description, depth)
VALUES (?, ?, ?, ?)
RETURNING id, parent_id, name, description, depth, created_at;

-- name: CreateRootItemType :one
INSERT INTO item_types (parent_id, name, description, depth)
VALUES (NULL, ?, ?, 0)
RETURNING id, parent_id, name, description, depth, created_at;

-- name: DeleteItemType :exec
DELETE FROM item_types WHERE id = ?;

-- name: GetItemTypeDescendants :many
WITH RECURSIVE descendants AS (
    SELECT id, parent_id, name, description, depth, created_at
    FROM item_types
    WHERE item_types.id = sqlc.arg(type_id)
    UNION ALL
    SELECT it.id, it.parent_id, it.name, it.description, it.depth, it.created_at
    FROM item_types it, descendants d
    WHERE it.parent_id = d.id
)
SELECT * FROM descendants
ORDER BY depth, name;
