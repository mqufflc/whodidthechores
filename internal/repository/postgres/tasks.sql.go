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
    user_id, chore_id, started_at, duration_mn
) VALUES (
    $1, $2, $3, $4
)
RETURNING id, user_id, chore_id, started_at, duration_mn
`

type CreateTaskParams struct {
	UserID     int32
	ChoreID    int32
	StartedAt  time.Time
	DurationMn int32
}

func (q *Queries) CreateTask(ctx context.Context, arg CreateTaskParams) (Task, error) {
	row := q.db.QueryRow(ctx, createTask,
		arg.UserID,
		arg.ChoreID,
		arg.StartedAt,
		arg.DurationMn,
	)
	var i Task
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.ChoreID,
		&i.StartedAt,
		&i.DurationMn,
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
SELECT id, user_id, chore_id, started_at, duration_mn FROM tasks
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
	)
	return i, err
}

const listTasks = `-- name: ListTasks :many
SELECT id, user_id, chore_id, started_at, duration_mn FROM tasks
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
