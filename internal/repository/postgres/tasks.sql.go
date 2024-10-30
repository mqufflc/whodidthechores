// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: tasks.sql

package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createTask = `-- name: CreateTask :one
INSERT INTO tasks (
    user_id, chore_id, started_at, duration_mn, description
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING id, user_id, chore_id, started_at, duration_mn, description
`

type CreateTaskParams struct {
	UserID      int32
	ChoreID     int32
	StartedAt   time.Time
	DurationMn  int32
	Description string
}

func (q *Queries) CreateTask(ctx context.Context, arg CreateTaskParams) (Task, error) {
	row := q.db.QueryRow(ctx, createTask,
		arg.UserID,
		arg.ChoreID,
		arg.StartedAt,
		arg.DurationMn,
		arg.Description,
	)
	var i Task
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.ChoreID,
		&i.StartedAt,
		&i.DurationMn,
		&i.Description,
	)
	return i, err
}

const deleteTask = `-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1
`

func (q *Queries) DeleteTask(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteTask, id)
	return err
}

const getTask = `-- name: GetTask :one
SELECT id, user_id, chore_id, started_at, duration_mn, description FROM tasks
WHERE id = $1
`

func (q *Queries) GetTask(ctx context.Context, id uuid.UUID) (Task, error) {
	row := q.db.QueryRow(ctx, getTask, id)
	var i Task
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.ChoreID,
		&i.StartedAt,
		&i.DurationMn,
		&i.Description,
	)
	return i, err
}

const getUserTasks = `-- name: GetUserTasks :many
SELECT tasks.id, tasks.user_id, tasks.chore_id, tasks.started_at, tasks.duration_mn, tasks.description, chores.id, chores.name, chores.description, chores.default_duration_mn, users.id, users.name
FROM tasks
JOIN chores ON tasks.chore_id = chores.id
JOIN users ON tasks.user_id = users.id
WHERE users.id = $1
`

type GetUserTasksRow struct {
	Task  Task
	Chore Chore
	User  User
}

func (q *Queries) GetUserTasks(ctx context.Context, id int32) ([]GetUserTasksRow, error) {
	rows, err := q.db.Query(ctx, getUserTasks, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetUserTasksRow
	for rows.Next() {
		var i GetUserTasksRow
		if err := rows.Scan(
			&i.Task.ID,
			&i.Task.UserID,
			&i.Task.ChoreID,
			&i.Task.StartedAt,
			&i.Task.DurationMn,
			&i.Task.Description,
			&i.Chore.ID,
			&i.Chore.Name,
			&i.Chore.Description,
			&i.Chore.DefaultDurationMn,
			&i.User.ID,
			&i.User.Name,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listTasks = `-- name: ListTasks :many
SELECT id, user_id, chore_id, started_at, duration_mn, description FROM tasks
ORDER BY started_at
`

func (q *Queries) ListTasks(ctx context.Context) ([]Task, error) {
	rows, err := q.db.Query(ctx, listTasks)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Task
	for rows.Next() {
		var i Task
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.ChoreID,
			&i.StartedAt,
			&i.DurationMn,
			&i.Description,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listUsersTasks = `-- name: ListUsersTasks :many
SELECT tasks.id, tasks.user_id, tasks.chore_id, tasks.started_at, tasks.duration_mn, tasks.description, chores.id, chores.name, chores.description, chores.default_duration_mn, users.id, users.name
FROM tasks
JOIN chores ON tasks.chore_id = chores.id
JOIN users ON tasks.user_id = users.id
ORDER BY tasks.started_at
`

type ListUsersTasksRow struct {
	Task  Task
	Chore Chore
	User  User
}

func (q *Queries) ListUsersTasks(ctx context.Context) ([]ListUsersTasksRow, error) {
	rows, err := q.db.Query(ctx, listUsersTasks)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListUsersTasksRow
	for rows.Next() {
		var i ListUsersTasksRow
		if err := rows.Scan(
			&i.Task.ID,
			&i.Task.UserID,
			&i.Task.ChoreID,
			&i.Task.StartedAt,
			&i.Task.DurationMn,
			&i.Task.Description,
			&i.Chore.ID,
			&i.Chore.Name,
			&i.Chore.Description,
			&i.Chore.DefaultDurationMn,
			&i.User.ID,
			&i.User.Name,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateTask = `-- name: UpdateTask :one
UPDATE tasks SET 
user_id = $2,
chore_id = $3,
started_at = $4,
duration_mn = $5,
description = $6
WHERE id = $1
RETURNING id, user_id, chore_id, started_at, duration_mn, description
`

type UpdateTaskParams struct {
	ID          uuid.UUID
	UserID      int32
	ChoreID     int32
	StartedAt   time.Time
	DurationMn  int32
	Description string
}

func (q *Queries) UpdateTask(ctx context.Context, arg UpdateTaskParams) (Task, error) {
	row := q.db.QueryRow(ctx, updateTask,
		arg.ID,
		arg.UserID,
		arg.ChoreID,
		arg.StartedAt,
		arg.DurationMn,
		arg.Description,
	)
	var i Task
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.ChoreID,
		&i.StartedAt,
		&i.DurationMn,
		&i.Description,
	)
	return i, err
}
