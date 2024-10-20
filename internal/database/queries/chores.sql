-- name: ListChores :many
SELECT * FROM chores
ORDER BY id;

-- name: GetChore :one
SELECT * FROM chores
WHERE id = $1;

-- name: CreateChore :one
INSERT INTO chores (
    name, description, default_duration_mn
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: UpdateChore :one
UPDATE chores SET 
name = $2,
description = $3,
default_duration_mn = $4
WHERE id = $1
RETURNING *;

-- name: DeleteChore :exec
DELETE FROM chores
WHERE id = $1;