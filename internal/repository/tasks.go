package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

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

type TaskParams struct {
	ID          uuid.UUID
	UserID      string
	ChoreID     string
	StartedAt   string
	DurationMn  string
	Description string
	Errors      TaskParamsError
}

type TaskParamsError struct {
	UserID      string
	ChoreID     string
	StartedAt   string
	DurationMn  string
	Description string
}

func (r *Repository) ValidateTask(ctx context.Context, taskParams *TaskParams, timezone time.Location) (postgres.CreateTaskParams, error) {
	isErr := false
	choreId, err := strconv.Atoi(taskParams.ChoreID)
	if err != nil {
		isErr = true
		taskParams.Errors.ChoreID = "Please select an existing chore"
	} else if err = r.ValidateTaskChoreId(ctx, choreId); err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			taskParams.Errors.ChoreID = "Chore not found"
		default:
			slog.Error(fmt.Sprintf("unable to validate a chore id: %v", err))
			taskParams.Errors.ChoreID = "Unable to validate this chore, please try again"
		}
	}
	userId, err := strconv.Atoi(taskParams.UserID)
	if err != nil {
		isErr = true
		taskParams.Errors.UserID = "Please select an existing user"
	} else if err = r.ValidateTaskUserId(ctx, userId); err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			taskParams.Errors.UserID = "User not found"
		default:
			slog.Error(fmt.Sprintf("unable to validate a task id: %v", err))
			taskParams.Errors.UserID = "Unable to validate this task, please try again"
		}
	}
	duration, err := strconv.Atoi(taskParams.DurationMn)
	if err != nil {
		isErr = true
		taskParams.Errors.DurationMn = "Please enter a number"
	} else if err = r.ValidateTaskDuration(duration); err != nil {
		switch {
		case errors.Is(err, ErrTooSmall):
			taskParams.Errors.DurationMn = "Duration can't be negative"
		case errors.Is(err, ErrTooBig):
			taskParams.Errors.DurationMn = "Duration too big, please select a smaller number"
		default:
			slog.Error(fmt.Sprintf("Unable to validate a task duration: %v", err))
			taskParams.Errors.DurationMn = "Unable to validate this duration, please try again"
		}
	}

	startedAt, err := time.ParseInLocation("2006-01-02T15:04", taskParams.StartedAt, &timezone)
	if err != nil {
		isErr = true
		slog.Warn(fmt.Sprintf("Unable to parse started time: %v", err))
		taskParams.Errors.StartedAt = "Please enter a valid date"
	}
	if isErr {
		return postgres.CreateTaskParams{}, ErrValidation
	}
	return postgres.CreateTaskParams{ChoreID: int32(choreId), UserID: int32(userId), Description: taskParams.Description, DurationMn: int32(duration), StartedAt: startedAt}, nil

}

func (r *Repository) ValidateTaskChoreId(ctx context.Context, choreId int) error {
	if choreId < 0 {
		return ErrNotFound
	}
	if choreId > 2147483647 {
		return ErrNotFound
	}
	chore, err := r.GetChore(ctx, int32(choreId))
	if err != nil {
		return fmt.Errorf("unable to get existing chore: %w", err)
	}
	if chore == (postgres.Chore{}) {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) ValidateTaskUserId(ctx context.Context, userId int) error {
	if userId < 0 {
		return ErrNotFound
	}
	if userId > 2147483647 {
		return ErrNotFound
	}
	user, err := r.GetUser(ctx, int32(userId))
	if err != nil {
		return fmt.Errorf("unable to get existing user: %w", err)
	}
	if user == (postgres.User{}) {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) ValidateTaskDuration(default_duration int) error {
	if default_duration < 0 {
		return ErrTooSmall
	}
	if default_duration > 2147483647 {
		return ErrTooBig
	}
	return nil
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

func (r *Repository) GetTask(ctx context.Context, id uuid.UUID) (postgres.Task, error) {
	task, err := r.q.GetTask(ctx, id)
	if err != nil {
		if sqlErr := taskPgError(err); sqlErr != nil {
			return postgres.Task{}, sqlErr
		}
		return postgres.Task{}, err
	}
	return task, nil
}

func (r *Repository) UpdateTask(ctx context.Context, id uuid.UUID, taskParams postgres.CreateTaskParams) (postgres.Task, error) {
	params := postgres.UpdateTaskParams{
		ID:          id,
		UserID:      taskParams.UserID,
		ChoreID:     taskParams.ChoreID,
		StartedAt:   taskParams.StartedAt,
		DurationMn:  taskParams.DurationMn,
		Description: taskParams.Description,
	}
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

func (r *Repository) DeleteTask(ctx context.Context, id uuid.UUID) error {
	err := r.q.DeleteTask(ctx, id)
	if err != nil {
		if sqlErr := taskPgError(err); sqlErr != nil {
			return sqlErr
		}
		return err
	}
	return nil
}

// func (r *Repository) GetUsersReport(ctx context.Context) (map[string][]ChoreReport, error) {
// 	reports, err := r.q.TasksReport(ctx)
// 	if err != nil {
// 		if sqlErr := taskPgError(err); sqlErr != nil {
// 			return nil, sqlErr
// 		}
// 		return nil, err
// 	}
// 	choreReports := make([]TaskReport, len(reports))
// 	for index, report := range reports {
// 		choreReports[index] = TaskReport{
// 			User:  User(report.User),
// 			Chore: Chore(report.Chore),
// 			Sum:   report.Sum,
// 		}
// 	}
// 	return GenerateUserReport(choreReports), nil
// }

func (r *Repository) GetChoreReport(ctx context.Context) (Report, error) {
	reports, err := r.q.TasksReport(ctx)
	if err != nil {
		if sqlErr := taskPgError(err); sqlErr != nil {
			return Report{}, sqlErr
		}
		return Report{}, err
	}
	choreReports := make([]TaskReport, len(reports))
	for index, report := range reports {
		choreReports[index] = TaskReport{
			User:  User(report.User),
			Chore: Chore(report.Chore),
			Sum:   report.Sum,
		}
	}
	return GenerateReport(choreReports), nil
}
