package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mqufflc/whodidthechores/internal/repository/postgres"
)

func chorePgError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}
	switch pgErr.ConstraintName {
	case "chores_name_key":
		return fmt.Errorf("%w: chore already exists", ErrDuplicateName)
	case "chores_name_check":
		return fmt.Errorf("%w: invalid chore name", ErrInvalidName)
	case "tasks_chore_id_fkey":
		return fmt.Errorf("%w: chore linked to existing task", ErrStillInUse)
	}
	slog.Error(fmt.Sprintf("uncaught chore pg error: %v", pgErr))
	return fmt.Errorf("%w: %w", ErrSQL, err)
}

type ChoreParams struct {
	ID                int32
	Name              string
	Description       string
	DefaultDurationMn string
	Errors            ChoreParamsError
}

type ChoreParamsError struct {
	Name              string
	Description       string
	DefaultDurationMn string
}

func (r *Repository) ValidateChore(ctx context.Context, choreParams *ChoreParams) (postgres.CreateChoreParams, error) {
	isErr := false
	if err := r.ValidateChoreName(ctx, choreParams.Name, choreParams.ID); err != nil {
		isErr = true
		switch {
		case errors.Is(err, ErrInvalidName):
			choreParams.Errors.Name = "Name can't be empty"
		case errors.Is(err, ErrDuplicateName):
			choreParams.Errors.Name = "Name already taken, please chose another one"
		default:
			slog.Error(fmt.Sprintf("Unable to validate a name: %v", err))
			choreParams.Errors.Name = "Unable to validate this name, please try again"
		}
	}
	if err := r.ValidateChoreDescription(choreParams.Description); err != nil {
		isErr = true
		switch {
		case errors.Is(err, ErrInvalidName):
			choreParams.Errors.Description = "Description can't be empty"
		default:
			slog.Error(fmt.Sprintf("Unable to validate a description: %v", err))
			choreParams.Errors.Description = "Unable to validate this description, please try again"
		}
	}
	default_duration, err := strconv.Atoi(choreParams.DefaultDurationMn)
	if err != nil {
		isErr = true
		choreParams.Errors.DefaultDurationMn = "Please enter a number"
	} else if err = r.ValidateChoreDefaultDuration(default_duration); err != nil {
		switch {
		case errors.Is(err, ErrTooSmall):
			choreParams.Errors.DefaultDurationMn = "Default duration can't be negative"
		case errors.Is(err, ErrTooBig):
			choreParams.Errors.DefaultDurationMn = "Default duration too big, please select a smaller number"
		default:
			slog.Error(fmt.Sprintf("Unable to validate a default duration: %v", err))
			choreParams.Errors.DefaultDurationMn = "Unable to validate this duration, please try again"
		}
	}
	if isErr {
		return postgres.CreateChoreParams{}, ErrValidation
	}
	return postgres.CreateChoreParams{Name: choreParams.Name, Description: choreParams.Description, DefaultDurationMn: int32(default_duration)}, nil
}

func (r *Repository) ValidateChoreName(ctx context.Context, name string, id int32) error {
	existingChores, err := r.q.ListChores(ctx)
	if err != nil {
		return fmt.Errorf("unable to get existing chores: %w", err)
	}

	if name == "" {
		return ErrInvalidName
	}

	for _, chore := range existingChores {
		if chore.Name == name && chore.ID != id {
			return ErrDuplicateName
		}
	}

	return nil
}

func (r *Repository) ValidateChoreDescription(name string) error {
	if name == "" {
		return ErrInvalidName
	}

	return nil
}

func (r *Repository) ValidateChoreDefaultDuration(default_duration int) error {
	if default_duration < 0 {
		return ErrTooSmall
	}
	if default_duration > 2147483647 {
		return ErrTooBig
	}
	return nil
}

func (r *Repository) CreateChore(ctx context.Context, params postgres.CreateChoreParams) (postgres.Chore, error) {
	newChore, err := r.q.CreateChore(ctx, params)
	if err != nil {
		if sqlErr := chorePgError(err); sqlErr != nil {
			return postgres.Chore{}, sqlErr
		}
		return postgres.Chore{}, err
	}

	return newChore, nil
}

func (r *Repository) ListChores(ctx context.Context) ([]postgres.Chore, error) {
	chores, err := r.q.ListChores(ctx)
	if err != nil {
		if sqlErr := chorePgError(err); sqlErr != nil {
			return nil, sqlErr
		}
		return nil, err
	}
	return chores, nil
}

func (r *Repository) GetChore(ctx context.Context, id int32) (postgres.Chore, error) {
	chore, err := r.q.GetChore(ctx, id)
	if err != nil {
		if sqlErr := chorePgError(err); sqlErr != nil {
			return postgres.Chore{}, sqlErr
		}
		return postgres.Chore{}, err
	}
	return chore, nil
}

func (r *Repository) UpdateChore(ctx context.Context, id int32, choreParams postgres.CreateChoreParams) (postgres.Chore, error) {
	params := postgres.UpdateChoreParams{
		ID:                id,
		Name:              choreParams.Name,
		Description:       choreParams.Description,
		DefaultDurationMn: choreParams.DefaultDurationMn,
	}
	chore, err := r.q.UpdateChore(ctx, params)
	if err != nil {
		if sqlErr := chorePgError(err); sqlErr != nil {
			return postgres.Chore{}, sqlErr
		}
		return postgres.Chore{}, err
	}
	return chore, nil
}

func (r *Repository) DeleteChore(ctx context.Context, id int32) error {
	err := r.q.DeleteChore(ctx, id)
	if err != nil {
		if sqlErr := chorePgError(err); sqlErr != nil {
			return sqlErr
		}
		return err
	}
	return nil
}
