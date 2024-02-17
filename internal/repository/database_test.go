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

	err = repositoryService.Migrate("../../migrations")
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

	chore, err := suite.repositoryService.CreateChore(ChoreParams{
		ID:          "ikhsyetd64",
		Name:        "Dishes",
		Description: "Wahing dishes",
	})

	assert.NoError(t, err)
	assert.NotNil(t, chore)

	_, err = suite.repositoryService.CreateChore(ChoreParams{
		ID:          "ikhsyetd64",
		Name:        "Dishes cloned",
		Description: "Wahing dishes",
	})

	assert.ErrorContains(t, err, "chore already exists")

	_, err = suite.repositoryService.CreateChore(ChoreParams{
		ID:          "ikhfyetd64",
		Name:        "Dishes",
		Description: "Wahing dishes",
	})

	assert.ErrorContains(t, err, "chore already exists")

	_, err = suite.repositoryService.CreateChore(ChoreParams{
		ID:          "",
		Name:        "Dishes",
		Description: "Wahing dishes",
	})

	assert.ErrorContains(t, err, "invalid chore ID")

	_, err = suite.repositoryService.CreateChore(ChoreParams{
		ID:          "ikhdyetd64",
		Name:        "",
		Description: "Wahing dishes",
	})

	assert.ErrorContains(t, err, "invalid chore name")
}

func (suite *ChoreRepoTestSuite) TestGetChore() {
	t := suite.T()

	chore, err := suite.repositoryService.CreateChore(ChoreParams{
		ID:          "ikhspetd64",
		Name:        "Dishes2",
		Description: "Washing dishes a second time",
	})

	assert.NoError(t, err)
	assert.NotNil(t, chore)

	getChore, err := suite.repositoryService.GetChore("ikhspetd64")
	assert.NoError(t, err)
	assert.Equal(t, chore.ID, getChore.ID)
	assert.Equal(t, chore.Name, getChore.Name)
	assert.Equal(t, chore.Description, getChore.Description)

	emptyChore, err := suite.repositoryService.GetChore("lopddddddd")

	assert.NoError(t, err)
	assert.Nil(t, emptyChore)
}

func (suite *ChoreRepoTestSuite) TestUpdateChore() {
	t := suite.T()

	chore, err := suite.repositoryService.CreateChore(ChoreParams{
		ID:          "ikhsperd64",
		Name:        "Dishes3",
		Description: "Washing dishes a third time",
	})

	assert.NoError(t, err)
	assert.NotNil(t, chore)

	updatedChore := ChoreParams{
		ID:          "ikhsperd64",
		Name:        "Dishes4",
		Description: "Washing dishes a fourth time",
	}

	storedUpdatedChore, err := suite.repositoryService.UpdateChore(updatedChore)
	assert.NoError(t, err)
	assert.Equal(t, chore.ID, storedUpdatedChore.ID)
	assert.Equal(t, updatedChore.Name, storedUpdatedChore.Name)
	assert.Equal(t, updatedChore.Description, storedUpdatedChore.Description)
	assert.NotEqual(t, chore.Name, storedUpdatedChore.Name)
	assert.NotEqual(t, chore.Description, storedUpdatedChore.Description)
	assert.NotEqual(t, chore.ModifiedAt, storedUpdatedChore.ModifiedAt)
	assert.Equal(t, chore.CreatedAt, storedUpdatedChore.CreatedAt)
}
