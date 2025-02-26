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
duration_mn = $5,
description = $6
WHERE id = $1
RETURNING *;

-- name: GetUserTasks :many
SELECT sqlc.embed(tasks), sqlc.embed(chores)
FROM tasks
JOIN chores ON tasks.chore_id = chores.id
JOIN users ON tasks.user_id = users.id
WHERE users.id = $1
ORDER BY tasks.started_at DESC;

-- name: GetChoreTasks :many
SELECT sqlc.embed(tasks), sqlc.embed(users)
FROM tasks
JOIN users ON tasks.user_id = users.id
WHERE tasks.chore_id = $1
ORDER BY tasks.started_at DESC;

-- name: ListUsersTasks :many
SELECT sqlc.embed(tasks), sqlc.embed(chores), sqlc.embed(users)
FROM tasks
JOIN chores ON tasks.chore_id = chores.id
JOIN users ON tasks.user_id = users.id
ORDER BY tasks.started_at DESC;

-- name: TasksReport :many
SELECT sqlc.embed(users), sqlc.embed(chores), SUM(duration_mn)
FROM tasks
JOIN chores ON tasks.chore_id = chores.id
JOIN users ON tasks.user_id = users.id
WHERE tasks.started_at > sqlc.arg(not_before) AND tasks.started_at < sqlc.arg(not_after)
GROUP BY chores.id, users.id;
