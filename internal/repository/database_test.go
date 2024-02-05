package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestChoreRepository(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := postgres.RunContainer(ctx, testcontainers.WithImage("postgres:15.3-alpine"), postgres.WithDatabase("whodidthechores"), postgres.WithUsername("postgres"), postgres.WithPassword("postgres"), testcontainers.WithWaitStrategy((wait.ForLog("database system is ready to accept connections").
		WithOccurrence(2).WithStartupTimeout(5 * time.Second))))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate pgContainer: %s", err)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	assert.NoError(t, err)

	choreRepo, err := NewService(connStr)
	assert.NoError(t, err)

	err = choreRepo.Migrate()
	assert.NoError(t, err)

	c, err := choreRepo.CreateChore(Chore{
		ID:          "ikhsyetd64",
		Name:        "Vaisselle",
		Description: "Faire la vaisselle",
	})

	assert.NoError(t, err)
	assert.NotNil(t, c)
}
