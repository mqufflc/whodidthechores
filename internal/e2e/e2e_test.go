//go:build e2e

package e2e

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// TestE2ESuite runs the E2E test suite
func TestE2ESuite(t *testing.T) {
	suite.Run(t, new(E2ESuite))
}