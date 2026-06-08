-- name: GetItem :one
SELECT id, name, item_type_id, location_id, quantity, unit_type, created_at
FROM items
WHERE id = ?;

-- name: ListItems :many
SELECT id, name, item_type_id, location_id, quantity, unit_type, created_at
FROM items
ORDER BY name;

-- name: ListItemsByType :many
SELECT id, name, item_type_id, location_id, quantity, unit_type, created_at
FROM items
WHERE item_type_id = ?
ORDER BY name;

-- name: ListItemsWithType :many
SELECT
    i.id,
    i.name,
    i.item_type_id,
    i.location_id,
    i.quantity,
    i.unit_type,
    i.created_at,
    t.name as type_name,
    t.parent_id as type_parent_id,
    t.depth as type_depth
FROM items i
JOIN item_types t ON i.item_type_id = t.id
ORDER BY i.name;

-- name: ListItemsWithTypePath :many
SELECT
    i.id,
    i.name,
    i.item_type_id,
    i.location_id,
    i.quantity,
    i.unit_type,
    i.created_at,
    t.name as type_name,
    t.parent_id as type_parent_id,
    t.depth as type_depth
FROM items i
JOIN item_types t ON i.item_type_id = t.id
ORDER BY t.depth, t.name, i.name;

-- name: CountItemsByType :one
SELECT COUNT(*) AS count
FROM items
WHERE item_type_id = ?;

-- name: CreateItem :one
INSERT INTO items (name, item_type_id, location_id, quantity, unit_type)
VALUES (?, ?, ?, ?, ?)
RETURNING id, name, item_type_id, location_id, quantity, unit_type, created_at;

-- name: DeleteItem :exec
DELETE FROM items WHERE id = ?;

-- name: UpdateItem :exec
UPDATE items
SET name = ?, item_type_id = ?, location_id = ?, quantity = ?, unit_type = ?
WHERE id = ?;
