package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mqufflc/whodidthechores/internal/repository/postgres"
)

func taskPgError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}
	switch pgErr.ConstraintName {
	case "tasks_name_key":
		return errors.New("task already exists")
	case "tasks_name_check":
		return errors.New("invalid task name")
	}
	slog.Error(fmt.Sprintf("uncaught task pg error: %v", pgErr.Code))
	return err
}

func (r *Repository) CreateTask(ctx context.Context, params postgres.CreateTaskParams) (postgres.Task, error) {
	newtask, err := r.q.CreateTask(ctx, params)
	if err != nil {
		if sqlErr := taskPgError(err); sqlErr != nil {
			return postgres.Task{}, sqlErr
		}
		return postgres.Task{}, err
	}

	return newtask, nil
}

func (r *Repository) ListTasks(ctx context.Context) ([]postgres.Task, error) {
	tasks, err := r.q.ListTasks(ctx)
	if err != nil {
		if sqlErr := taskPgError(err); sqlErr != nil {
			return nil, sqlErr
		}
		return nil, err
	}
	return tasks, nil
}

func (r *Repository) Gettask(ctx context.Context, id uuid.UUID) (postgres.Task, error) {
	task, err := r.q.GetTask(ctx, id)
	if err != nil {
		if sqlErr := taskPgError(err); sqlErr != nil {
			return postgres.Task{}, sqlErr
		}
		return postgres.Task{}, err
	}
	return task, nil
}

func (r *Repository) Updatetask(ctx context.Context, params postgres.UpdateTaskParams) (postgres.Task, error) {
	task, err := r.q.UpdateTask(ctx, params)
	if err != nil {
		if sqlErr := taskPgError(err); sqlErr != nil {
			return postgres.Task{}, sqlErr
		}
		return postgres.Task{}, err
	}
	return task, nil
}

func (r *Repository) ListUsersTasks(ctx context.Context) ([]postgres.ListUsersTasksRow, error) {
	tasks, err := r.q.ListUsersTasks(ctx)
	if err != nil {
		if sqlErr := taskPgError(err); sqlErr != nil {
			return nil, sqlErr
		}
		return nil, err
	}
	return tasks, nil
}
