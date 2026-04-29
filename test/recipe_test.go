package test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"citadel/internal/recipe"
	"citadel/internal/session"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRecipe_Authenticated(t *testing.T) {
	req, err := http.NewRequest("GET", server.URL+"/recipes/"+td.Recipe, nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var r recipe.Recipe
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&r))
	assert.Equal(t, td.Recipe, r.ID)
	assert.Equal(t, "Test Recipe", r.Title)
	require.Len(t, r.Components, 1)
	assert.Len(t, r.Components[0].Ingredients, 1)
	assert.Equal(t, "Water", r.Components[0].Ingredients[0].Item)
	assert.Len(t, r.Components[0].Instructions, 1)
	assert.Equal(t, "Boil water", r.Components[0].Instructions[0])
}

func TestGetRecipe_Unauthenticated(t *testing.T) {
	req, err := http.NewRequest("GET", server.URL+"/recipes/"+td.Recipe, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var r recipe.Recipe
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&r))
	assert.Equal(t, td.Recipe, r.ID)
}

func TestListRecipes_Authenticated(t *testing.T) {
	req, err := http.NewRequest("GET", server.URL+"/recipes", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		Items []recipe.Recipe `json:"items"`
		Total int             `json:"total"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(t, result.Items)
	assert.Greater(t, result.Total, 0)
}

func TestCreateRecipe_Authenticated(t *testing.T) {
	// 10 mins = 600,000,000,000 nanoseconds
	payload := `{
		"title": "New Recipe",
		"description": "New Description",
		"components": [{
			"ingredients": [{"amount": 2, "unit": "whole", "item": "Eggs"}],
			"instructions": ["Crack eggs", "Fry eggs"]
		}],
		"cook_time": 600000000000,
		"serves": 2,
		"cuisine": "American",
		"category": "Main"
	}`
	req, err := http.NewRequest("POST", server.URL+"/recipes", strings.NewReader(payload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(t, result["recipe_id"])
}

func TestUpdateRecipe_OwnRecipe(t *testing.T) {
	payload := `{"title": "Updated Title"}`
	req, err := http.NewRequest(
		"PATCH",
		server.URL+"/recipes/"+td.Recipe,
		strings.NewReader(payload),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify update
	req, _ = http.NewRequest("GET", server.URL+"/recipes/"+td.Recipe, nil)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})
	resp, _ = http.DefaultClient.Do(req)
	defer resp.Body.Close()
	var r recipe.Recipe
	json.NewDecoder(resp.Body).Decode(&r)
	assert.Equal(t, "Updated Title", r.Title)
}

func TestUpdateRecipe_OtherUser(t *testing.T) {
	payload := `{"title": "Hacked"}`
	req, err := http.NewRequest(
		"PATCH",
		server.URL+"/recipes/"+td.Recipe,
		strings.NewReader(payload),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(
		&http.Cookie{Name: session.CookieName, Value: td.Admin.Session},
	) // Admin is not the owner

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestDeleteRecipe_OwnRecipe(t *testing.T) {
	// Create a recipe to delete
	payload := `{"title": "Delete Me", "components": [], "cuisine": "American", "category": "Main"}`
	req, _ := http.NewRequest("POST", server.URL+"/recipes", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})
	resp, _ := http.DefaultClient.Do(req)
	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	recipeID := result["recipe_id"]
	resp.Body.Close()

	// Delete it
	req, err := http.NewRequest("DELETE", server.URL+"/recipes/"+recipeID, nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify deletion — soft-deleted recipes are filtered by ById, returning 404
	req, _ = http.NewRequest("GET", server.URL+"/recipes/"+recipeID, nil)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})
	resp, _ = http.DefaultClient.Do(req)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDeleteRecipe_OtherUser(t *testing.T) {
	req, err := http.NewRequest("DELETE", server.URL+"/recipes/"+td.Recipe, nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.Admin.Session})

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
