package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

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
		return errors.New("chore already exists")
	case "chores_name_check":
		return errors.New("invalid chore name")
	}
	slog.Error(fmt.Sprintf("uncaught chore pg error: %v", pgErr.Code))
	return err
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

func (r *Repository) UpdateChore(ctx context.Context, params postgres.UpdateChoreParams) (postgres.Chore, error) {
	chore, err := r.q.UpdateChore(ctx, params)
	if err != nil {
		if sqlErr := chorePgError(err); sqlErr != nil {
			return postgres.Chore{}, sqlErr
		}
		return postgres.Chore{}, err
	}
	return chore, nil
}
