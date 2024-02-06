package repository

import (
	"context"
	"log"
	"testing"

	"github.com/mqufflc/whodidthechores/internal/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ChoreRepoTestSuite struct {
	suite.Suite
	pgContainer       *testhelpers.PostgresContainer
	repositoryService *Service
	ctx               context.Context
}

func TestChoreRepoTestSuit(t *testing.T) {
	suite.Run(t, new(ChoreRepoTestSuite))
}

func (suite *ChoreRepoTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	pgContainer, err := testhelpers.CreatePostgesContainer(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}
	suite.pgContainer = pgContainer

	repositoryService, err := NewService(suite.pgContainer.ConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	err = repositoryService.Migrate()
	if err != nil {
		log.Fatalf("error during database migration : %s", err)
	}
	suite.repositoryService = repositoryService
}

func (suite *ChoreRepoTestSuite) TearDownSuite() {
	if err := suite.pgContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating postgres container : %s", err)
	}
}

func (suite *ChoreRepoTestSuite) TestCreateChore() {
	t := suite.T()

	customer, err := suite.repositoryService.CreateChore(Chore{
		ID:          "ikhsyetd64",
		Name:        "Vaisselle",
		Description: "Faire la vaisselle",
	})

	assert.NoError(t, err)
	assert.NotNil(t, customer)
}
