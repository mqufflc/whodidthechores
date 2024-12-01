package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mqufflc/whodidthechores/internal/config"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

type Effector func(str string) error

func retry(effector Effector, retries int, delay time.Duration) Effector {
	return func(str string) error {
		for r := 0; ; r++ {
			err := effector(str)
			if err == nil || r >= retries {
				return err
			}
			slog.Warn(fmt.Sprintf("Attempt %d failed; retrying in %v", r+1, delay))
			<-time.After(delay)
		}
	}
}

func checkDatabaseConnectiviy(connString string) error {
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return fmt.Errorf("unable to create database connection: %w", err)
	}
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("connectivity test error: %w", err)
	}
	err = db.Close()
	if err != nil {
		slog.Error("unable to close connection after database connectiviy test")
		return err
	}
	return nil
}

func Migrate(connString string) error {
	slog.Info("checking database connectivity")
	r := retry(checkDatabaseConnectiviy, 3, 10*time.Second)
	err := r(connString)
	if err != nil {
		return fmt.Errorf("all attempts to connect to database failed: %w", err)
	}
	slog.Info("applying migrations")
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return fmt.Errorf("unable to connect to database before migrations: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			slog.Warn("failed to close migration db connection")
		}
	}()

	pg_driver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		return err
	}

	source_driver, err := iofs.New(embedMigrations, "migrations")
	if err != nil {
		return fmt.Errorf("unable to access embeded migrations: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source_driver, "pgx5", pg_driver)
	if err != nil {
		return fmt.Errorf("unable to create migration instance: %w", err)
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
		slog.Info("migrations applied")
		return nil
	}
}

func Connect(ctx context.Context, config config.DbConfig) (*pgxpool.Pool, error) {
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", config.Username, url.QueryEscape(config.Password), config.Hostname, config.Port, config.Database, config.SslMode)

	if err := Migrate(connectionString); err != nil {
		return nil, fmt.Errorf("applying migrations failed: %w", err)
	}

	dbpool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		return nil, fmt.Errorf("unable to open a connection to the database: %w", err)
	}

	return dbpool, nil
}
