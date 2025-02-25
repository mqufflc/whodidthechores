package repository

import (
	"context"
	"log"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mqufflc/whodidthechores/internal/database"
	"github.com/mqufflc/whodidthechores/internal/repository/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresContainer struct {
	*tcpostgres.PostgresContainer
	ConnectionString string
}

func CreatePostgesContainer(ctx context.Context) (*PostgresContainer, error) {
	pgContainer, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("whodidthechores"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
			wait.ForListeningPort("5432/tcp")))
	if err != nil {
		return nil, err
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, err
	}

	return &PostgresContainer{
		PostgresContainer: pgContainer,
		ConnectionString:  connStr,
	}, nil
}

type RepositoryTestSuite struct {
	suite.Suite
	pgContainer *PostgresContainer
	repository  *Repository
	dbpool      *pgxpool.Pool
	ctx         context.Context
}

func TestRepositoryTestSuit(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

func (suite *RepositoryTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	pgContainer, err := CreatePostgesContainer(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}
	suite.pgContainer = pgContainer

	err = database.Migrate(pgContainer.ConnectionString)
	if err != nil {
		log.Fatalf("unable to apply database migrations: %v", err)
	}

	suite.dbpool, err = pgxpool.New(suite.ctx, pgContainer.ConnectionString)
	if err != nil {
		log.Fatalf("unable to open a connection to the database: %s", err)
	}

	suite.repository = New(NewRepositoryParams{DB: suite.dbpool})
}

func (suite *RepositoryTestSuite) TearDownSuite() {
	suite.dbpool.Close()
	if err := suite.pgContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating postgres container : %s", err)
	}
}

func (suite *RepositoryTestSuite) TestCreateChore() {
	t := suite.T()

	chore, err := suite.repository.CreateChore(suite.ctx, postgres.CreateChoreParams{
		Name:        "Dishes",
		Description: "Wahing dishes",
	})

	assert.NoError(t, err)
	assert.NotNil(t, chore)

	_, err = suite.repository.CreateChore(suite.ctx, postgres.CreateChoreParams{
		Name:        "Dishes",
		Description: "Wahing dishes",
	})

	assert.ErrorContains(t, err, "chore already exists")

	_, err = suite.repository.CreateChore(suite.ctx, postgres.CreateChoreParams{
		Name:        "",
		Description: "Wahing dishes",
	})

	assert.ErrorContains(t, err, "invalid chore name")
}
