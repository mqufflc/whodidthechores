-- name: ListTasks :many
SELECT * FROM tasks
ORDER BY started_at;

-- name: GetTask :one
SELECT * FROM tasks
WHERE id = $1;

-- name: CreateTask :one
INSERT INTO tasks (
    user_id, chore_id, started_at, duration_mn
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1;