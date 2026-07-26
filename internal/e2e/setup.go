//go:build e2e

package e2e

import (
	"context"
	"log"
	"log/slog"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mqufflc/whodidthechores/internal/api"
	"github.com/mqufflc/whodidthechores/internal/config"
	"github.com/mqufflc/whodidthechores/internal/database"
	"github.com/mqufflc/whodidthechores/internal/repository"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestServer struct {
	Server     *httptest.Server
	Repository *repository.Repository
	DBPool     *pgxpool.Pool
	Container  testcontainers.Container
	Context    context.Context
}

func SetupTestServer(t *testing.T) *TestServer {
	t.Helper()

	ctx := context.Background()

	// Create PostgreSQL container
	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("whodidthechores_test"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
			wait.ForListeningPort("5432/tcp")),
	)
	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	// Create logger for migrations
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Run migrations
	err = database.Migrate(connStr, logger)
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create database pool
	dbPool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("Failed to create database pool: %v", err)
	}

	// Create repository
	repo := repository.New(repository.NewRepositoryParams{DB: dbPool})

	// Create test config
	testConfig := config.Config{
		Port:     8080,
		TimeZone: "UTC",
		Database: config.DbConfig{
			Username: "postgres",
			Password: "postgres",
			Hostname: "localhost",
			Port:     5432,
			Database: "whodidthechores_test",
			SslMode:  "disable",
		},
	}

	// Create API handler
	handler := api.New(ctx, repo, testConfig, logger)

	// Create test server
	server := httptest.NewServer(handler)

	return &TestServer{
		Server:     server,
		Repository: repo,
		DBPool:     dbPool,
		Container:  pgContainer,
		Context:    ctx,
	}
}

func (ts *TestServer) Cleanup() {
	ts.Server.Close()
	ts.DBPool.Close()
	if err := ts.Container.Terminate(ts.Context); err != nil {
		log.Printf("Warning: failed to terminate container: %v", err)
	}
}