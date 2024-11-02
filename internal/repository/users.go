package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mqufflc/whodidthechores/internal/repository/postgres"
)

func userPgError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}
	switch pgErr.ConstraintName {
	case "users_name_key":
		return ErrDuplicateName
	case "users_name_check":
		return ErrInvalidName
	case "tasks_user_id_fkey":
		return ErrStillInUse
	}
	slog.Error(fmt.Sprintf("uncaught user pg error: %v", pgErr.Code))
	return err
}

type UserParams struct {
	ID     int32
	Name   string
	Errors UserParamsError
}

type UserParamsError struct {
	Name string
}

func (r *Repository) ValidateUser(ctx context.Context, userParams *UserParams) (string, error) {
	isErr := false
	if err := r.ValidateUserName(ctx, userParams.Name, userParams.ID); err != nil {
		isErr = true
		switch {
		case errors.Is(err, ErrInvalidName):
			userParams.Errors.Name = "Name can't be empty"
		case errors.Is(err, ErrDuplicateName):
			userParams.Errors.Name = "Name already taken, please chose another one"
		default:
			slog.Error(fmt.Sprintf("Unable to validate a name: %v", err))
			userParams.Errors.Name = "Unable to validate this name, please try again"
		}
	}
	if isErr {
		return "", ErrValidation
	}
	return userParams.Name, nil
}

func (r *Repository) ValidateUserName(ctx context.Context, name string, id int32) error {
	existingUsers, err := r.q.ListUsers(ctx)
	if err != nil {
		return fmt.Errorf("unable to get existing users: %w", err)
	}

	if name == "" {
		return ErrInvalidName
	}

	for _, user := range existingUsers {
		if user.Name == name && user.ID != id {
			return ErrDuplicateName
		}
	}

	return nil
}

func (r *Repository) CreateUser(ctx context.Context, name string) (postgres.User, error) {
	newuser, err := r.q.CreateUser(ctx, name)
	if err != nil {
		if sqlErr := userPgError(err); sqlErr != nil {
			return postgres.User{}, sqlErr
		}
		return postgres.User{}, err
	}

	return newuser, nil
}

func (r *Repository) ListUsers(ctx context.Context) ([]postgres.User, error) {
	users, err := r.q.ListUsers(ctx)
	if err != nil {
		if sqlErr := userPgError(err); sqlErr != nil {
			return nil, sqlErr
		}
		return nil, err
	}
	return users, nil
}

func (r *Repository) GetUser(ctx context.Context, id int32) (postgres.User, error) {
	user, err := r.q.GetUser(ctx, id)
	if err != nil {
		if sqlErr := userPgError(err); sqlErr != nil {
			return postgres.User{}, sqlErr
		}
		return postgres.User{}, err
	}
	return user, nil
}

func (r *Repository) UpdateUser(ctx context.Context, id int32, name string) (postgres.User, error) {
	params := postgres.UpdateUserParams{
		ID:   id,
		Name: name,
	}
	user, err := r.q.UpdateUser(ctx, params)
	if err != nil {
		if sqlErr := userPgError(err); sqlErr != nil {
			return postgres.User{}, sqlErr
		}
		return postgres.User{}, err
	}
	return user, nil
}

func (r *Repository) DeleteUser(ctx context.Context, id int32) error {
	err := r.q.DeleteUser(ctx, id)
	if err != nil {
		if sqlErr := userPgError(err); sqlErr != nil {
			return sqlErr
		}
		return err
	}
	return nil
}
