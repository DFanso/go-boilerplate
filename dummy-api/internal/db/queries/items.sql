-- name: CreateItem :one
INSERT INTO items (
    owner_id,
    name,
    description
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: ListItems :many
SELECT * FROM items
WHERE owner_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetItem :one
SELECT * FROM items WHERE id = $1 AND owner_id = $2;

-- name: DeleteItem :exec
DELETE FROM items WHERE id = $1 AND owner_id = $2;
