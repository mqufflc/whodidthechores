package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/google/uuid"

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

func (service *Service) Migrate(migration_file_path string) error {
	driver, err := postgres.WithInstance(service.db, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance("file://"+migration_file_path, "postgres", driver)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil {
		switch {
		case errors.Is(err, migrate.ErrNoChange):
			slog.Info("No new migration to apply.")
			return nil
		default:
			return err
		}
	} else {
		slog.Info("Mgrations applied")
		return nil
	}
}

func userPgError(err error) error {
	var pgErr *pq.Error
	if !errors.As(err, &pgErr) {
		return nil
	}
	if pgErr.Code.Name() == "unique_violation" {
		return errors.New("user already exists")
	}
	if pgErr.Code.Name() == "check_violation" {
		switch pgErr.Constraint {
		case "users_id_check":
			return errors.New("invalid user ID")
		case "users_name_check":
			return errors.New("invalid user name")
		}
	}
	slog.Error(fmt.Sprintf("uncaught user pg error: %v", pgErr.Code.Name()))
	return err
}

func (service *Service) CreateUser(params UserParams) (createdUser *User, err error) {
	createdUser = &User{}

	tx, err := service.db.Begin()
	if err != nil {
		return createdUser, err
	}
	defer func() {
		rollbackErr := tx.Rollback()
		if errors.Is(rollbackErr, sql.ErrTxDone) {
			return
		}
		if err != nil {
			if rollbackErr != nil {
				log.Printf("failed to rollback user creation: %v", err)
			}
			return
		}
		err = rollbackErr
	}()

	query, err := tx.Prepare("INSERT INTO users (id, name, hash) VALUES ($1, $2, $3) RETURNING id, name, hash, created_at, modified_at")

	if err != nil {
		return createdUser, err
	}

	err = query.QueryRow(params.ID, params.Name, params.Hash).Scan(&createdUser.ID, &createdUser.Name, &createdUser.Hash, &createdUser.CreatedAt, &createdUser.ModifiedAt)

	if err != nil {
		if sqlErr := userPgError(err); sqlErr != nil {
			return createdUser, sqlErr
		}
		return createdUser, err
	}

	err = tx.Commit()

	if err != nil {
		return createdUser, err
	}

	return createdUser, nil
}

func (service *Service) GetUser(id string) (user *User, err error) {
	user = &User{}

	query, err := service.db.Prepare("SELECT id, name, hash FROM users WHERE id = $1")

	defer func() {
		if closeErr := query.Close(); closeErr != nil {
			log.Printf("failed to close query: %v\n", err)
		}
	}()

	if err != nil {
		return user, err
	}

	sqlErr := query.QueryRow(id).Scan(&user.ID, &user.Name, &user.Hash)

	if sqlErr != nil {
		if sqlErr == sql.ErrNoRows {
			return nil, nil
		}
		return user, sqlErr
	}

	return user, nil
}

func (service *Service) ListUser(id string) (*[]User, error) {
	rows, err := service.db.Query("SELECT id, name, hash FROM users")

	if err != nil {
		return nil, err
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			log.Printf("failed to close list users query: %v\n", err)
		}
	}()

	users := make([]User, 0)

	for rows.Next() {
		user := User{}
		err = rows.Scan(&user.ID, &user.Name, &user.Hash)

		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	err = rows.Err()

	if err != nil {
		return nil, err
	}

	return &users, nil
}

func (service *Service) SearchUserByName(name string) (user *User, err error) {
	user = &User{}

	query, err := service.db.Prepare("SELECT id, name, hash FROM users WHERE name = $1")

	defer func() {
		if closeErr := query.Close(); closeErr != nil {
			log.Printf("failed to close search query: %v\n", err)
		}
	}()

	if err != nil {
		return user, err
	}

	sqlErr := query.QueryRow(name).Scan(&user.ID, &user.Name, &user.Hash)

	if sqlErr != nil {
		if sqlErr == sql.ErrNoRows {
			return nil, nil
		}
		return user, sqlErr
	}

	return user, nil
}

func (service *Service) CreateSession(params SessionParams) (session *Session, err error) {
	session = &Session{}
	session.User = &User{}

	tx, err := service.db.Begin()
	if err != nil {
		return session, err
	}
	defer func() {
		rollbackErr := tx.Rollback()
		if errors.Is(rollbackErr, sql.ErrTxDone) {
			return
		}
		if err != nil {
			if rollbackErr != nil {
				log.Printf("failed to rollback session creation: %v", err)
			}
			return
		}
		err = rollbackErr
	}()

	query, err := tx.Prepare(`WITH inserted_session as (INSERT INTO user_sessions (id, user_id, expires_at) VALUES ($1, $2, $3) RETURNING *) SELECT inserted_session.id, inserted_session.created_at, inserted_session.last_used_at, inserted_session.expires_at, users.id, users.name, users.hash FROM inserted_session JOIN users ON inserted_session.user_id = users.id`)

	if err != nil {
		return session, err
	}

	err = query.QueryRow(params.ID, params.UserID, params.ExpiresAt).Scan(&session.ID, &session.CreatedAt, &session.LastUsedAt, &session.ExpiresAt, &session.User.ID, &session.User.Name, &session.User.Hash)

	if err != nil {
		return session, err
	}

	err = tx.Commit()

	if err != nil {
		return session, err
	}

	return session, nil
}

func (service *Service) GetSession(id uuid.UUID) (session *Session, err error) {
	session = &Session{}
	session.User = &User{}

	query, err := service.db.Prepare(`SELECT user_sessions.id, user_sessions.created_at, user_sessions.last_used_at, user_sessions.expires_at, users.id, users.name, users.hash 
		FROM user_sessions JOIN users ON user_sessions.user_id = users.id WHERE user_sessions.id = $1`)

	if err != nil {
		return session, err
	}

	defer func() {
		if closeErr := query.Close(); closeErr != nil {
			log.Printf("failed to close get session query: %v\n", err)
		}
	}()

	sqlErr := query.QueryRow(id).Scan(&session.ID, &session.CreatedAt, &session.LastUsedAt, &session.ExpiresAt, &session.User.ID, &session.User.Name, &session.User.Hash)

	if sqlErr != nil {
		if sqlErr == sql.ErrNoRows {
			return nil, nil
		}
		return session, sqlErr
	}

	return session, nil
}

func (service *Service) UseSession(id uuid.UUID) (session *Session, err error) {
	session = &Session{}
	session.User = &User{}

	tx, err := service.db.Begin()

	if err != nil {
		return session, err
	}

	defer func() {
		rollbackErr := tx.Rollback()
		if errors.Is(rollbackErr, sql.ErrTxDone) {
			return
		}
		if err != nil {
			if rollbackErr != nil {
				log.Printf("failed to rollback session update: %v", err)
			}
			return
		}
		err = rollbackErr
	}()

	query, err := tx.Prepare(`WITH updated_session as (UPDATE user_sessions SET last_used_at = $2 WHERE id = $1 RETURNING *)
		SELECT updated_session.id, updated_session.created_at, updated_session.last_used_at, updated_session.expires_at, user.id, user.name, user.hash 
		FROM updated_session JOIN users as user ON updated_session.user_id = user.id`)

	if err != nil {
		return session, err
	}

	sqlErr := query.QueryRow(id, time.Now()).Scan(&session.ID, &session.CreatedAt, &session.LastUsedAt, &session.ExpiresAt, &session.User.ID, &session.User.Name, &session.User.Hash)

	if sqlErr != nil {
		if sqlErr == sql.ErrNoRows {
			return nil, nil
		}
		return session, sqlErr
	}

	err = tx.Commit()

	if err != nil {
		return session, err
	}

	return session, nil

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

func (service *Service) GetChore(id string) (chore *Chore, err error) {
	chore = &Chore{}

	query, err := service.db.Prepare("SELECT id, name, description FROM chores WHERE id = $1")

	defer func() {
		if closeErr := query.Close(); closeErr != nil {
			log.Printf("failed to close get session query: %v\n", err)
		}
	}()

	if err != nil {
		return chore, err
	}

	sqlErr := query.QueryRow(id).Scan(&chore.ID, &chore.Name, &chore.Description)

	if sqlErr != nil {
		if sqlErr == sql.ErrNoRows {
			return nil, nil
		}
		return chore, sqlErr
	}

	return chore, nil
}

func (service *Service) CreateChore(params ChoreParams) (chore *Chore, err error) {
	chore = &Chore{}

	tx, err := service.db.Begin()
	if err != nil {
		return
	}
	defer func() {
		rollbackErr := tx.Rollback()
		if errors.Is(rollbackErr, sql.ErrTxDone) {
			return
		}
		if err != nil {
			if rollbackErr != nil {
				log.Printf("failed to rollback chore update: %v", err)
			}
			return
		}
		err = rollbackErr
	}()

	query, err := tx.Prepare("INSERT INTO chores (id, name, description) VALUES ($1, $2, $3) RETURNING id, name, description, created_at, modified_at")

	if err != nil {
		return
	}

	err = query.QueryRow(params.ID, params.Name, params.Description).Scan(&chore.ID, &chore.Name, &chore.Description, &chore.CreatedAt, &chore.ModifiedAt)

	if err != nil {
		if sqlErr := chorePgError(err); sqlErr != nil {
			return chore, sqlErr
		}
		return
	}

	err = tx.Commit()

	if err != nil {
		return
	}

	return
}

func (service *Service) ListChores() (*[]Chore, error) {
	rows, err := service.db.Query("SELECT id, name, description FROM chores")

	if err != nil {
		return nil, err
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			log.Printf("failed to close list chore query: %v\n", err)
		}
	}()

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

func (service *Service) UpdateChore(params ChoreParams) (chore *Chore, err error) {
	chore = &Chore{}

	tx, err := service.db.Begin()

	if err != nil {
		return chore, err
	}

	defer func() {
		rollbackErr := tx.Rollback()
		if errors.Is(rollbackErr, sql.ErrTxDone) {
			return
		}
		if err != nil {
			if rollbackErr != nil {
				log.Printf("failed to rollback chore update: %v", err)
			}
			return
		}
		err = rollbackErr
	}()

	query, err := tx.Prepare(`UPDATE chores SET name = COALESCE($2, name), description = COALESCE($3, description), modified_at = NOW() WHERE id = $1 RETURNING id, name, description, created_at, modified_at`)

	if err != nil {
		return chore, err
	}

	sqlErr := query.QueryRow(params.ID, params.Name, params.Description).Scan(&chore.ID, &chore.Name, &chore.Description, &chore.CreatedAt, &chore.ModifiedAt)

	if sqlErr != nil {
		return chore, sqlErr
	}

	err = tx.Commit()

	if err != nil {
		return chore, err
	}

	return chore, nil
}

func (service *Service) DeleteChore(id string) (deleted bool, err error) {
	tx, err := service.db.Begin()

	if err != nil {
		return false, err
	}

	defer func() {
		rollbackErr := tx.Rollback()
		if errors.Is(rollbackErr, sql.ErrTxDone) {
			return
		}
		if err != nil {
			if rollbackErr != nil {
				log.Printf("failed to rollback chore delete: %v", err)
			}
			return
		}
		err = rollbackErr
	}()

	query, err := tx.Prepare("DELETE FROM chores WHERE id = $1")

	if err != nil {
		return false, err
	}

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
