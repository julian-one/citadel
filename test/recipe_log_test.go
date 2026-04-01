package test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"citadel/internal/session"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- Bookmark Tests ----------

func TestBookmarkRecipe(t *testing.T) {
	req, err := http.NewRequest("PUT", server.URL+"/recipes/"+td.RecipeId+"/bookmark", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]bool
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.True(t, result["bookmarked"])
}

func TestBookmarkRecipe_Duplicate(t *testing.T) {
	// Bookmark again — should be idempotent
	req, err := http.NewRequest("PUT", server.URL+"/recipes/"+td.RecipeId+"/bookmark", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetBookmarkStatus_Bookmarked(t *testing.T) {
	// First bookmark
	req, _ := http.NewRequest("PUT", server.URL+"/recipes/"+td.RecipeId+"/bookmark", nil)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.Admin.Session})
	resp, _ := http.DefaultClient.Do(req)
	resp.Body.Close()

	// Check status
	req, err := http.NewRequest("GET", server.URL+"/recipes/"+td.RecipeId+"/bookmark", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.Admin.Session})

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]bool
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.True(t, result["bookmarked"])
}

func TestGetBookmarkStatus_NotBookmarked(t *testing.T) {
	// Create a new recipe that won't be bookmarked
	payload := `{"title": "Unbookmarked Recipe", "ingredients": [], "instructions": []}`
	req, _ := http.NewRequest("POST", server.URL+"/recipes", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})
	resp, _ := http.DefaultClient.Do(req)
	var created map[string]string
	json.NewDecoder(resp.Body).Decode(&created)
	resp.Body.Close()
	newRecipeId := created["recipe_id"]

	// Check bookmark status (should be false)
	req, err := http.NewRequest("GET", server.URL+"/recipes/"+newRecipeId+"/bookmark", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.Admin.Session})

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]bool
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.False(t, result["bookmarked"])
}

func TestUnbookmarkRecipe(t *testing.T) {
	// Bookmark first
	req, _ := http.NewRequest("PUT", server.URL+"/recipes/"+td.RecipeId+"/bookmark", nil)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})
	resp, _ := http.DefaultClient.Do(req)
	resp.Body.Close()

	// Unbookmark
	req, err := http.NewRequest("DELETE", server.URL+"/recipes/"+td.RecipeId+"/bookmark", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestBookmarkRecipe_Unauthenticated(t *testing.T) {
	req, err := http.NewRequest("PUT", server.URL+"/recipes/"+td.RecipeId+"/bookmark", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// ---------- Recipe Log Tests ----------

func TestCreateRecipeLog(t *testing.T) {
	payload := `{
		"notes": "Added extra garlic. Cooked at altitude — needed 5 extra mins.",
		"rating": 4.5,
		"duration": 2700000000000,
		"intensity": 2
	}`
	req, err := http.NewRequest(
		"POST",
		server.URL+"/recipes/"+td.RecipeId+"/logs",
		strings.NewReader(payload),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(t, result["log_id"])
}

func TestCreateRecipeLog_MinimalFields(t *testing.T) {
	payload := `{}`
	req, err := http.NewRequest(
		"POST",
		server.URL+"/recipes/"+td.RecipeId+"/logs",
		strings.NewReader(payload),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(t, result["log_id"])
}

func TestListRecipeLogs(t *testing.T) {
	// Create a log first
	payload := `{"notes": "Test log for listing", "rating": 3.0}`
	req, _ := http.NewRequest(
		"POST",
		server.URL+"/recipes/"+td.RecipeId+"/logs",
		strings.NewReader(payload),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})
	resp, _ := http.DefaultClient.Do(req)
	resp.Body.Close()

	// List logs
	req, err := http.NewRequest("GET", server.URL+"/recipes/"+td.RecipeId+"/logs", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var logs []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&logs))
	assert.NotEmpty(t, logs)
}

func TestDeleteRecipeLog(t *testing.T) {
	// Create a log
	payload := `{"notes": "To be deleted"}`
	req, _ := http.NewRequest(
		"POST",
		server.URL+"/recipes/"+td.RecipeId+"/logs",
		strings.NewReader(payload),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})
	resp, _ := http.DefaultClient.Do(req)
	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()
	logId := result["log_id"]

	// Delete it
	req, err := http.NewRequest("DELETE", server.URL+"/recipe-logs/"+logId, nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestDeleteRecipeLog_OtherUser(t *testing.T) {
	// Create a log as regular user
	payload := `{"notes": "User's log"}`
	req, _ := http.NewRequest(
		"POST",
		server.URL+"/recipes/"+td.RecipeId+"/logs",
		strings.NewReader(payload),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})
	resp, _ := http.DefaultClient.Do(req)
	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()
	logId := result["log_id"]

	// Try to delete as admin (different user)
	req, err := http.NewRequest("DELETE", server.URL+"/recipe-logs/"+logId, nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.Admin.Session})

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestRecipeLog_Unauthenticated(t *testing.T) {
	payload := `{"notes": "Should fail"}`
	req, err := http.NewRequest(
		"POST",
		server.URL+"/recipes/"+td.RecipeId+"/logs",
		strings.NewReader(payload),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
