package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
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

func (db *Service) Migrate() error {
	driver, err := postgres.WithInstance(db.db, &postgres.Config{})
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

func (db *Service) GetChore(id int) (*Chore, error) {
	chore := Chore{}

	query, err := db.db.Prepare("SELECT id, name, description FROM chores WHERE id = $1;")

	if err != nil {
		return &chore, err
	}

	sqlErr := query.QueryRow(id).Scan(&chore.ID, &chore.Name, &chore.Description)

	if sqlErr != nil {
		if sqlErr == sql.ErrNoRows {
			return &chore, nil
		}
		return &chore, sqlErr
	}

	return &chore, nil
}

func (db *Service) CreateChore(chore Chore) (*Chore, error) {
	var createdChore Chore

	tx, err := db.db.Begin()
	if err != nil {
		return &createdChore, err
	}
	defer tx.Rollback()

	query, err := tx.Prepare("INSERT INTO chores (id, name, description) VALUES ($1, $2, $3) RETURNING id, name, description, created_at, modified_at;")

	if err != nil {
		return &createdChore, err
	}

	err = query.QueryRow(chore.ID, chore.Name, chore.Description).Scan(&createdChore.ID, &createdChore.Name, &createdChore.Description, &createdChore.CreatedAt, &createdChore.ModifiedAt)

	if err != nil {
		return &createdChore, err
	}

	err = tx.Commit()

	if err != nil {
		return &createdChore, err
	}

	return &createdChore, nil
}

func (db *Service) ListChores() (*[]Chore, error) {
	rows, err := db.db.Query("SELECT id, name, description FROM chores;")

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

func (db *Service) UpdateChore(id int, updatedChore Chore) (*Chore, error) {

	chore := Chore{}

	tx, err := db.db.Begin()

	if err != nil {
		return &chore, err
	}

	defer tx.Rollback()

	query, err := tx.Prepare("UPDATE chores SET name = $2, description = $3 WHERE id = $1 RETURNING id, name, description")

	if err != nil {
		return &chore, err
	}

	sqlErr := query.QueryRow(id, updatedChore.Name, updatedChore.Description).Scan(&chore.ID, &chore.Name, &chore.Description)

	if sqlErr != nil {
		return &chore, sqlErr
	}

	err = tx.Commit()

	if err != nil {
		return &chore, err
	}

	return &chore, nil
}

func (db *Service) DeleteChore(id int) (bool, error) {
	tx, err := db.db.Begin()

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
