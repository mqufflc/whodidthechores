package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lib/pq"
)

type Service struct {
	db *sql.DB
}

func NewService(connStr string) (*Service, error) {
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, err
	}

	err = db.Ping()

	if err != nil {
		return nil, fmt.Errorf("unable to ping the database : %w", err)
	}

	return &Service{db: db}, nil
}

func (service *Service) Migrate() error {
	driver, err := postgres.WithInstance(service.db, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil {
		switch {
		case errors.Is(err, migrate.ErrNoChange):
			log.Println("No new migration to apply.")
		default:
			return err
		}
	}
	return nil
}

func chorePgError(err error) error {
	var pgErr *pq.Error
	if !errors.As(err, &pgErr) {
		return nil
	}
	if pgErr.Code.Name() == "unique_violation" {
		return errors.New("chore already exists")
	}
	if pgErr.Code.Name() == "check_violation" {
		switch pgErr.Constraint {
		case "chores_id_check":
			return errors.New("invalid chore ID")
		case "chores_name_check":
			return errors.New("invalid chore name")
		}
	}
	fmt.Printf("%v", pgErr.Code.Name())
	return err
}

func (service *Service) GetChore(id string) (*Chore, error) {
	chore := Chore{}

	query, err := service.db.Prepare("SELECT id, name, description FROM chores WHERE id = $1")

	if err != nil {
		return &chore, err
	}

	sqlErr := query.QueryRow(id).Scan(&chore.ID, &chore.Name, &chore.Description)

	if sqlErr != nil {
		if sqlErr == sql.ErrNoRows {
			return nil, nil
		}
		return &chore, sqlErr
	}

	return &chore, nil
}

func (service *Service) CreateChore(params ChoreParams) (*Chore, error) {
	var createdChore Chore

	tx, err := service.db.Begin()
	if err != nil {
		return &createdChore, err
	}
	defer tx.Rollback()

	query, err := tx.Prepare("INSERT INTO chores (id, name, description) VALUES ($1, $2, $3) RETURNING id, name, description, created_at, modified_at")

	if err != nil {
		return &createdChore, err
	}

	err = query.QueryRow(params.ID, params.Name, params.Description).Scan(&createdChore.ID, &createdChore.Name, &createdChore.Description, &createdChore.CreatedAt, &createdChore.ModifiedAt)

	if err != nil {
		if sqlErr := chorePgError(err); err != nil {
			return &createdChore, sqlErr
		}
		return &createdChore, err
	}

	err = tx.Commit()

	if err != nil {
		return &createdChore, err
	}

	return &createdChore, nil
}

func (service *Service) ListChores() (*[]Chore, error) {
	rows, err := service.db.Query("SELECT id, name, description FROM chores")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	chores := make([]Chore, 0)

	for rows.Next() {
		chore := Chore{}
		err = rows.Scan(&chore.ID, &chore.Name, &chore.Description)

		if err != nil {
			return nil, err
		}

		chores = append(chores, chore)
	}

	err = rows.Err()

	if err != nil {
		return nil, err
	}

	return &chores, nil
}

func (service *Service) UpdateChore(params ChoreParams) (*Chore, error) {

	chore := Chore{}

	tx, err := service.db.Begin()

	if err != nil {
		return &chore, err
	}

	defer tx.Rollback()

	query, err := tx.Prepare(`UPDATE chores SET name = COALESCE($2, name), description = COALESCE($3, description), modified_at = NOW() WHERE id = $1 RETURNING id, name, description, created_at, modified_at`)

	if err != nil {
		return &chore, err
	}

	sqlErr := query.QueryRow(params.ID, params.Name, params.Description).Scan(&chore.ID, &chore.Name, &chore.Description, &chore.CreatedAt, &chore.ModifiedAt)

	if sqlErr != nil {
		return &chore, sqlErr
	}

	err = tx.Commit()

	if err != nil {
		return &chore, err
	}

	return &chore, nil
}

func (service *Service) DeleteChore(id string) (bool, error) {
	tx, err := service.db.Begin()

	if err != nil {
		return false, err
	}

	defer tx.Rollback()

	query, err := tx.Prepare("DELETE FROM chores WHERE id = $1")

	if err != nil {
		return false, err
	}

	defer query.Close()

	_, err = query.Exec(id)

	if err != nil {
		return false, err
	}

	err = tx.Commit()

	if err != nil {
		return false, err
	}

	return true, nil
}
