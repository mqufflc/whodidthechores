package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mqufflc/whodidthechores/internal/repository/postgres"
)

type Repository struct {
	db *pgxpool.Pool
	q  *postgres.Queries
}

type NewRepositoryParams struct {
	DB *pgxpool.Pool
}

func New(p NewRepositoryParams) *Repository {
	return &Repository{
		db: p.DB,
		q:  postgres.New(p.DB),
	}
}
