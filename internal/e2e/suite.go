//go:build e2e

package e2e

import (
	"github.com/stretchr/testify/suite"
)

// E2ESuite is the main test suite for end-to-end tests
type E2ESuite struct {
	suite.Suite
	TestServer *TestServer
}

// SetupSuite runs once before all tests in the suite
func (s *E2ESuite) SetupSuite() {
	s.T().Log("Setting up E2E test suite...")
	s.TestServer = SetupTestServer(s.T())
	s.T().Log("Test suite setup complete")
}

// TearDownSuite runs once after all tests in the suite
func (s *E2ESuite) TearDownSuite() {
	s.T().Log("Tearing down E2E test suite...")
	s.TestServer.Cleanup()
	s.T().Log("Test suite teardown complete")
}

// SetupTest runs before each individual test
func (s *E2ESuite) SetupTest() {
	// Optional: Add any per-test setup here
	// For example, could clear specific database tables between tests
}

// TearDownTest runs after each individual test  
func (s *E2ESuite) TearDownTest() {
	// Optional: Add any per-test cleanup here
}