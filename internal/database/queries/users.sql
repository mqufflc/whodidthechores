-- name: ListUsers :many
SELECT * FROM users
ORDER BY id;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (
    name
) VALUES (
    $1
)
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: UpdateUser :one
UPDATE users SET 
name = $2
WHERE id = $1
RETURNING *;
