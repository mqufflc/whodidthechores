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
		return errors.New("user already exists")
	case "users_name_check":
		return errors.New("invalid user name")
	}
	slog.Error(fmt.Sprintf("uncaught user pg error: %v", pgErr.Code))
	return err
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

func (r *Repository) UpdateUser(ctx context.Context, params postgres.UpdateUserParams) (postgres.User, error) {
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
