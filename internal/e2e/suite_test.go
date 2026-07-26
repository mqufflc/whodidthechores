//go:build e2e

package e2e

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestChoreWorkflow tests the complete chore creation and management flow
func (s *E2ESuite) TestChoreWorkflow() {
	t := s.T()
	ts := s.TestServer

	t.Run("Create chore", func(t *testing.T) {
		// Test GET create form
		resp, err := ts.Server.Client().Get(ts.Server.URL + "/chores/new")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Test POST create chore
		formData := url.Values{
			"name":             {"Wash dishes"},
			"description":      {"Washing all the dishes"},
			"default_duration": {"30"},
		}

		// Create a client that doesn't follow redirects automatically
		testClient := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		resp, err = testClient.PostForm(ts.Server.URL+"/chores/new", formData)
		require.NoError(t, err)
		// Check if redirect header is set correctly
		assert.Equal(t, "/chores", resp.Header.Get("Location"))
		// Check status code - should be 303 SeeOther for redirect
		assert.Equal(t, http.StatusSeeOther, resp.StatusCode, "Expected redirect status code")
	})

	t.Run("List chores", func(t *testing.T) {
		resp, err := ts.Server.Client().Get(ts.Server.URL + "/chores")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "Wash dishes")
	})

	t.Run("View chore details", func(t *testing.T) {
		// First get the chore ID from list
		resp, err := ts.Server.Client().Get(ts.Server.URL + "/chores")
		require.NoError(t, err)

		choreID := ExtractChoreIDFromList(t, resp)
		if choreID == "" {
			t.Skip("No chores available to view")
			return
		}

		// Test view chore
		resp, err = ts.Server.Client().Get(ts.Server.URL + "/chores/" + choreID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Edit chore", func(t *testing.T) {
		// Get chore list to find a chore
		resp, err := ts.Server.Client().Get(ts.Server.URL + "/chores")
		require.NoError(t, err)

		choreID := ExtractChoreIDFromList(t, resp)
		if choreID == "" {
			t.Skip("No chores available to edit")
			return
		}

		// Test HTMX PUT request to edit chore
		formData := url.Values{
			"name":             {"Updated chore name"},
			"description":      {"Updated description"},
			"default_duration": {"45"},
		}

		req, err := http.NewRequest("PUT", ts.Server.URL+"/chores/"+choreID+"/edit", strings.NewReader(formData.Encode()))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("HX-Request", "true")

		resp, err = ts.Server.Client().Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("HX-Location"), "/chores/"+choreID)
	})
}

// TestUserWorkflow tests the complete user creation and management flow
func (s *E2ESuite) TestUserWorkflow() {
	t := s.T()
	ts := s.TestServer

	t.Run("Create user", func(t *testing.T) {
		formData := url.Values{
			"name": {"John Doe"},
		}
		var resp *http.Response
		// Create a client that doesn't follow redirects automatically
		testClient := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		resp, err := testClient.PostForm(ts.Server.URL+"/users/new", formData)

		require.NoError(t, err)
		assert.Equal(t, http.StatusSeeOther, resp.StatusCode)
		assert.Equal(t, "/users", resp.Header.Get("Location"))
	})

	t.Run("List users", func(t *testing.T) {
		resp, err := ts.Server.Client().Get(ts.Server.URL + "/users")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "John Doe")
	})

	t.Run("View user details", func(t *testing.T) {
		// First get the user ID from list
		resp, err := ts.Server.Client().Get(ts.Server.URL + "/users")
		require.NoError(t, err)

		userID := ExtractUserIDFromList(t, resp)
		if userID == "" {
			t.Skip("No users available to view")
			return
		}

		// Test view user
		resp, err = ts.Server.Client().Get(ts.Server.URL + "/users/" + userID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Edit user", func(t *testing.T) {
		// Get user list to find a user
		resp, err := ts.Server.Client().Get(ts.Server.URL + "/users")
		require.NoError(t, err)

		userID := ExtractUserIDFromList(t, resp)
		if userID == "" {
			t.Skip("No users available to edit")
			return
		}

		// Test HTMX PUT request to edit user
		formData := url.Values{
			"name": {"Updated user name"},
		}

		req, err := http.NewRequest("PUT", ts.Server.URL+"/users/"+userID+"/edit", strings.NewReader(formData.Encode()))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("HX-Request", "true")

		resp, err = ts.Server.Client().Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("HX-Location"), "/users/"+userID)
	})
}

// TestTaskWorkflow tests the complete task creation and management flow
func (s *E2ESuite) TestTaskWorkflow() {
	t := s.T()
	ts := s.TestServer

	t.Run("Create prerequisites", func(t *testing.T) {
		// Create a client that doesn't follow redirects automatically
		testClient := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		choreForm := url.Values{"name": {"Test chore"}, "description": {"Test"}, "default_duration": {"30"}}
		_, err := testClient.PostForm(ts.Server.URL+"/chores/new", choreForm)
		require.NoError(t, err)

		userForm := url.Values{"name": {"Test user"}}
		_, err = testClient.PostForm(ts.Server.URL+"/users/new", userForm)
		require.NoError(t, err)
	})

	t.Run("Create task", func(t *testing.T) {
		// Get the actual chore and user IDs from the database
		chores, err := ts.Repository.ListChores(ts.Context)
		require.NoError(t, err)
		require.True(t, len(chores) > 0, "No chores found in database")
		choreID := strconv.FormatInt(int64(chores[0].ID), 10)

		users, err := ts.Repository.ListUsers(ts.Context)
		require.NoError(t, err)
		require.True(t, len(users) > 0, "No users found in database")
		userID := strconv.FormatInt(int64(users[0].ID), 10)

		// Get the task creation page first to see available chores/users
		resp, err := ts.Server.Client().Get(ts.Server.URL + "/tasks/new")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Create task
		formData := url.Values{
			"chore-id":    {choreID},
			"user-id":     {userID},
			"start-time":  {time.Now().Format("2006-01-02T15:04")},
			"duration":    {"30"},
			"description": {"Test task"},
		}

		// Create a client that doesn't follow redirects automatically
		testClient := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		resp, err = testClient.PostForm(ts.Server.URL+"/tasks/new", formData)
		require.NoError(t, err)
		assert.Equal(t, http.StatusSeeOther, resp.StatusCode)
		assert.Equal(t, "/tasks", resp.Header.Get("Location"))
	})

	t.Run("List tasks", func(t *testing.T) {
		resp, err := ts.Server.Client().Get(ts.Server.URL + "/tasks")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "Test task")
	})

	t.Run("Edit task", func(t *testing.T) {
		// Get the task, user and chore ID from the database
		tasks, err := ts.Repository.ListTasks(ts.Context)
		require.NoError(t, err)
		require.True(t, len(tasks) > 0, "No tasks found in database")
		taskID := tasks[0].ID
		taskIDString := tasks[0].ID.String()
		choreID := tasks[0].ChoreID
		choreIDString := strconv.Itoa(int(choreID))
		userID := tasks[0].UserID
		userIDString := strconv.Itoa(int(userID))

		// Test GET edit form
		resp, err := ts.Server.Client().Get(ts.Server.URL + "/tasks/" + taskIDString)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Test PUT request to edit task
		formData := url.Values{
			"chore-id":    {choreIDString}, // Using same chore for simplicity
			"user-id":     {userIDString},  // Using same user for simplicity
			"start-time":  {time.Now().Add(1 * time.Hour).Format("2006-01-02T15:04")},
			"duration":    {"45"},
			"description": {"Updated task description"},
		}

		req, err := http.NewRequest("PUT", ts.Server.URL+"/tasks/"+taskIDString, strings.NewReader(formData.Encode()))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err = ts.Server.Client().Do(req)
		require.NoError(t, err)
		// API returns 200 and renders the edit form again after successful update
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify the task was updated by checking the description in the database
		tasksAfterUpdate, err := ts.Repository.GetTask(ts.Context, taskID)
		require.NoError(t, err)
		assert.Equal(t, "Updated task description", tasksAfterUpdate.Description, "Task was not updated with new description")
	})

	t.Run("Delete task", func(t *testing.T) {
		// Get the task ID from the database
		tasks, err := ts.Repository.ListTasks(ts.Context)
		require.NoError(t, err)
		require.True(t, len(tasks) > 0, "No tasks found in database")
		taskID := tasks[0].ID.String()

		// Test HTMX DELETE request
		req, err := http.NewRequest("DELETE", ts.Server.URL+"/tasks/"+taskID, nil)
		require.NoError(t, err)
		req.Header.Set("HX-Request", "true")

		resp, err := ts.Server.Client().Do(req)
		require.NoError(t, err)
		// API returns 204 No Content for HTMX DELETE requests
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		// HTMX uses HX-Location header instead of Location
		assert.Equal(t, "/tasks", resp.Header.Get("HX-Location"))

		// Verify task was deleted by checking it's no longer in the list
		tasksAfterDelete, err := ts.Repository.ListTasks(ts.Context)
		require.NoError(t, err)
		require.True(t, len(tasksAfterDelete) < len(tasks), "Task was not deleted")
	})

	t.Run("Task creation with negative duration should fail validation", func(t *testing.T) {
		// Get the actual chore and user IDs from the database
		chores, err := ts.Repository.ListChores(ts.Context)
		require.NoError(t, err)
		require.True(t, len(chores) > 0, "No chores found in database")
		choreID := strconv.FormatInt(int64(chores[0].ID), 10)

		users, err := ts.Repository.ListUsers(ts.Context)
		require.NoError(t, err)
		require.True(t, len(users) > 0, "No users found in database")
		userID := strconv.FormatInt(int64(users[0].ID), 10)

		// Test with negative duration (should be rejected by validation)
		formData := url.Values{
			"chore-id":    {choreID},
			"user-id":     {userID},
			"start-time":  {time.Now().Format("2006-01-02T15:04")},
			"duration":    {"-1"}, // Invalid negative duration
			"description": {"Invalid negative duration task"},
		}

		resp, err := ts.Server.Client().PostForm(ts.Server.URL+"/tasks/new", formData)
		require.NoError(t, err)
		// Should return to form with validation errors, not redirect
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		bodyStr := string(body)
		// Check that the form is being displayed (indicating validation error)
		assert.Contains(t, bodyStr, "Task Values")
		// Check that the invalid duration value is still in the form
		assert.Contains(t, bodyStr, "value=\"-1\"")
		// Check that a validation error message is displayed
		assert.Contains(t, bodyStr, "Duration can&#39;t be negative")
	})

	t.Run("Task creation with non-existent chore ID should fail validation", func(t *testing.T) {
		users, err := ts.Repository.ListUsers(ts.Context)
		require.NoError(t, err)
		require.True(t, len(users) > 0, "No users found in database")
		userID := strconv.FormatInt(int64(users[0].ID), 10)

		formData := url.Values{
			"chore-id":    {"999999"}, // Non-existent chore ID
			"user-id":     {userID},
			"start-time":  {time.Now().Format("2006-01-02T15:04")},
			"duration":    {"30"},
			"description": {"Task with non-existent chore"},
		}

		resp, err := ts.Server.Client().PostForm(ts.Server.URL+"/tasks/new", formData)
		require.NoError(t, err)
		// Should return to form with validation errors, not redirect
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		bodyStr := string(body)
		// Check that the form is being displayed (indicating validation error)
		assert.Contains(t, bodyStr, "Task Values")
		// Check that the specific "not found" validation error message is displayed
		assert.Contains(t, bodyStr, "Chore not found")
	})

	t.Run("Task creation with non-existent user ID should fail validation", func(t *testing.T) {
		chores, err := ts.Repository.ListChores(ts.Context)
		require.NoError(t, err)
		require.True(t, len(chores) > 0, "No chores found in database")
		choreID := strconv.FormatInt(int64(chores[0].ID), 10)

		formData := url.Values{
			"chore-id":    {choreID},
			"user-id":     {"999999"}, // Non-existent user ID
			"start-time":  {time.Now().Format("2006-01-02T15:04")},
			"duration":    {"30"},
			"description": {"Task with non-existent user"},
		}

		resp, err := ts.Server.Client().PostForm(ts.Server.URL+"/tasks/new", formData)
		require.NoError(t, err)
		// Should return to form with validation errors, not redirect
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		bodyStr := string(body)
		// Check that the form is being displayed (indicating validation error)
		assert.Contains(t, bodyStr, "Task Values")
		// Check that the specific "not found" validation error message is displayed
		assert.Contains(t, bodyStr, "User not found")
	})
}

// TestIndexPage tests the main index page functionality
func (s *E2ESuite) TestIndexPage() {
	t := s.T()
	ts := s.TestServer

	resp, err := ts.Server.Client().Get(ts.Server.URL + "/")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "Who Did The Chores")
}

// TestHTMXEndpoints tests HTMX-specific functionality
func (s *E2ESuite) TestHTMXEndpoints() {
	t := s.T()
	ts := s.TestServer

	// Create test data
	choreForm := url.Values{"name": {"HTMX chore"}, "description": {"Test"}, "default_duration": {"30"}}
	_, err := ts.Server.Client().PostForm(ts.Server.URL+"/chores/new", choreForm)
	require.NoError(t, err)

	t.Run("HTMX chore edit", func(t *testing.T) {
		// Get chore list to find ID
		resp, err := ts.Server.Client().Get(ts.Server.URL + "/chores")
		require.NoError(t, err)

		choreID := ExtractChoreIDFromList(t, resp)
		if choreID == "" {
			t.Skip("No chores found")
			return
		}

		// Test HTMX PUT request
		formData := url.Values{
			"name":             {"Updated HTMX chore"},
			"description":      {"Updated description"},
			"default_duration": {"45"},
		}

		req, err := http.NewRequest("PUT", ts.Server.URL+"/chores/"+choreID+"/edit", strings.NewReader(formData.Encode()))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("HX-Request", "true")

		resp, err = ts.Server.Client().Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("HX-Location"), "/chores/"+choreID)
	})

	t.Run("HTMX chore delete", func(t *testing.T) {
		// Get chore list to find ID
		resp, err := ts.Server.Client().Get(ts.Server.URL + "/chores")
		require.NoError(t, err)

		choreID := ExtractChoreIDFromList(t, resp)
		if choreID == "" {
			t.Skip("No chores found")
			return
		}

		// Test HTMX DELETE request
		req, err := http.NewRequest("DELETE", ts.Server.URL+"/chores/"+choreID+"/edit", nil)
		require.NoError(t, err)
		req.Header.Set("HX-Request", "true")

		resp, err = ts.Server.Client().Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		assert.Equal(t, "/chores", resp.Header.Get("HX-Location"))
	})
}

// TestErrorHandling tests various error scenarios
func (s *E2ESuite) TestErrorHandling() {
	t := s.T()
	ts := s.TestServer

	t.Run("Invalid chore creation", func(t *testing.T) {
		formData := url.Values{
			"name":             {""}, // Empty name should fail validation
			"description":      {"Test"},
			"default_duration": {"30"},
		}
		resp, err := ts.Server.Client().PostForm(ts.Server.URL+"/chores/new", formData)
		require.NoError(t, err)
		// Should return to form with validation errors, not redirect
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		bodyStr := string(body)
		assert.Contains(t, bodyStr, "Name can&#39;t be empty")
	})

	t.Run("Non-existent chore", func(t *testing.T) {
		resp, err := ts.Server.Client().Get(ts.Server.URL + "/chores/99999")
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Non-existent user", func(t *testing.T) {
		resp, err := ts.Server.Client().Get(ts.Server.URL + "/users/99999")
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
