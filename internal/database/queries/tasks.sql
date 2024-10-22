-- name: ListTasks :many
SELECT * FROM tasks
ORDER BY started_at;

-- name: GetTask :one
SELECT * FROM tasks
WHERE id = $1;

-- name: CreateTask :one
INSERT INTO tasks (
    user_id, chore_id, started_at, duration_mn, description
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1;

-- name: UpdateTask :one
UPDATE tasks SET 
user_id = $2,
chore_id = $3,
started_at = $4,
duration_mn = $5
WHERE id = $1
RETURNING *;

-- name: GetUserTasks :many
SELECT sqlc.embed(tasks), sqlc.embed(chores), sqlc.embed(users)
FROM tasks
JOIN chores ON tasks.chore_id = chores.id
JOIN users ON tasks.user_id = users.id
WHERE users.id = $1;

-- name: ListUsersTasks :many
SELECT sqlc.embed(tasks), sqlc.embed(chores), sqlc.embed(users)
FROM tasks
JOIN chores ON tasks.chore_id = chores.id
JOIN users ON tasks.user_id = users.id;