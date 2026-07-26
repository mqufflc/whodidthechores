//go:build e2e

package e2e

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// ExtractChoreIDFromList extracts the first chore ID from the chores list page
func ExtractChoreIDFromList(t *testing.T, resp *http.Response) string {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	defer resp.Body.Close()

	bodyStr := string(body)
	// Look for pattern like "/chores/123"
	startIdx := strings.Index(bodyStr, "/chores/")
	if startIdx == -1 {
		t.Skip("No chores found in list")
		return ""
	}
	startIdx += len("/chores/")
	endIdx := startIdx
	for endIdx < len(bodyStr) && bodyStr[endIdx] >= '0' && bodyStr[endIdx] <= '9' {
		endIdx++
	}
	return bodyStr[startIdx:endIdx]
}

// ExtractUserIDFromList extracts the first user ID from the users list page
func ExtractUserIDFromList(t *testing.T, resp *http.Response) string {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	defer resp.Body.Close()

	bodyStr := string(body)
	// Look for pattern like "/users/123"
	startIdx := strings.Index(bodyStr, "/users/")
	if startIdx == -1 {
		t.Skip("No users found in list")
		return ""
	}
	startIdx += len("/users/")
	endIdx := startIdx
	for endIdx < len(bodyStr) && bodyStr[endIdx] >= '0' && bodyStr[endIdx] <= '9' {
		endIdx++
	}
	return bodyStr[startIdx:endIdx]
}
