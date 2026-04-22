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

// ---------- Recipe Review Tests ----------

func createTestRecipe(t *testing.T) string {
	payload := `{
		"title": "Review Recipe",
		"ingredients": [{"amount": 1, "unit": "cup", "item": "water"}],
		"instructions": ["boil"]
	}`
	req, _ := http.NewRequest("POST", server.URL+"/recipes", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	return result["recipe_id"]
}

func TestCreateRecipeReview(t *testing.T) {
	recipeID := createTestRecipe(t)
	payload := `{
		"notes": "Added extra garlic. Cooked at altitude — needed 5 extra mins.",
		"rating": 4,
		"duration": 2700000000000,
		"difficulty": 2
	}`
	req, err := http.NewRequest(
		"POST",
		server.URL+"/recipes/"+recipeID+"/reviews",
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
	assert.NotEmpty(t, result["review_id"])
}

func TestCreateRecipeReview_MinimalFields(t *testing.T) {
	recipeID := createTestRecipe(t)
	payload := `{"rating": 3}`
	req, err := http.NewRequest(
		"POST",
		server.URL+"/recipes/"+recipeID+"/reviews",
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
	assert.NotEmpty(t, result["review_id"])
}

func TestListRecipeReviews(t *testing.T) {
	recipeID := createTestRecipe(t)
	// Create a review first
	payload := `{"notes": "Test review for listing", "rating": 3, "duration": 1800000000000, "difficulty": 2}`
	req, _ := http.NewRequest(
		"POST",
		server.URL+"/recipes/"+recipeID+"/reviews",
		strings.NewReader(payload),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})
	resp, _ := http.DefaultClient.Do(req)
	resp.Body.Close()

	// List reviews (Public)
	req, err := http.NewRequest("GET", server.URL+"/recipes/"+recipeID+"/reviews", nil)
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var reviews []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&reviews))
	assert.NotEmpty(t, reviews)
	assert.NotEmpty(t, reviews[0]["username"])
}

func TestDeleteRecipeReview(t *testing.T) {
	recipeID := createTestRecipe(t)
	// Create a review
	payload := `{"notes": "To be deleted", "rating": 3, "duration": 1800000000000, "difficulty": 2}`
	req, _ := http.NewRequest(
		"POST",
		server.URL+"/recipes/"+recipeID+"/reviews",
		strings.NewReader(payload),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})
	resp, _ := http.DefaultClient.Do(req)
	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()
	reviewId := result["review_id"]

	// Delete it
	req, err := http.NewRequest("DELETE", server.URL+"/recipe-reviews/"+reviewId, nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestDeleteRecipeReview_OtherUser(t *testing.T) {
	recipeID := createTestRecipe(t)
	// Create a review as regular user
	payload := `{"notes": "User's review", "rating": 3, "duration": 1800000000000, "difficulty": 2}`
	req, _ := http.NewRequest(
		"POST",
		server.URL+"/recipes/"+recipeID+"/reviews",
		strings.NewReader(payload),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})
	resp, _ := http.DefaultClient.Do(req)
	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()
	reviewId := result["review_id"]

	// Try to delete as admin (different user)
	req, err := http.NewRequest("DELETE", server.URL+"/recipe-reviews/"+reviewId, nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.Admin.Session})

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestRecipeReview_Unauthenticated(t *testing.T) {
	payload := `{"notes": "Should fail"}`
	req, err := http.NewRequest(
		"POST",
		server.URL+"/recipes/"+td.RecipeId+"/reviews",
		strings.NewReader(payload),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
